// Web client with device-adaptive Matrix intro and CRT effects
package main

import (
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/yourusername/matrix-mud/pkg/admin"
	"github.com/yourusername/matrix-mud/pkg/logging"
	"github.com/yourusername/matrix-mud/pkg/metrics"
)

var upgrader = websocket.Upgrader{CheckOrigin: checkWebSocketOrigin}

func checkWebSocketOrigin(r *http.Request) bool {
	origin := r.Header.Get("Origin")
	if origin == "" || Config.AllowedOrigins == "*" {
		return true
	}
	for _, a := range strings.Split(Config.AllowedOrigins, ",") {
		if strings.TrimSpace(a) == origin {
			return true
		}
	}
	return false
}

func filterTelnetIAC(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		if data[i] == 255 && i+1 < len(data) {
			switch data[i+1] {
			case 255:
				result = append(result, 255)
				i += 2
			case 251, 252, 253, 254:
				i += 3
			case 250:
				i += 2
				for i < len(data) && !(data[i] == 255 && i+1 < len(data) && data[i+1] == 240) {
					i++
				}
				i += 2
			default:
				i += 2
			}
		} else {
			result = append(result, data[i])
			i++
		}
	}
	return result
}

func startWebServer(w *World) {
	startTime := time.Now()
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/ws", handleWebSocket)
	mux.HandleFunc("/health", handleHealth)
	mux.Handle("/metrics", metrics.Handler())
	mux.Handle("/admin/dashboard", admin.Handler(Version, startTime))
	logging.Info().Str("port", Config.WebPort).Msg("Web Portal active")
	http.ListenAndServe("0.0.0.0:"+Config.WebPort, mux)
}

func handleHealth(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"status":"healthy","version":"` + Version + `","service":"matrix-mud"}`))
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(htmlClient))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, _ := upgrader.Upgrade(w, r, nil)
	if ws == nil {
		return
	}
	defer ws.Close()
	tcpConn, err := net.Dial("tcp", "localhost:"+Config.TelnetPort)
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Connection failed\r\n"))
		return
	}
	defer tcpConn.Close()
	go func() {
		buf := make([]byte, 4096)
		for {
			n, e := tcpConn.Read(buf)
			if e != nil {
				return
			}
			if f := filterTelnetIAC(buf[:n]); len(f) > 0 {
				if ws.WriteMessage(websocket.TextMessage, f) != nil {
					return
				}
			}
		}
	}()
	for {
		_, m, e := ws.ReadMessage()
		if e != nil {
			break
		}
		tcpConn.Write(m)
	}
}

const htmlClient = `<!DOCTYPE html>
<html>
<head>
<title>The Construct</title>
<meta name="viewport" content="width=device-width,initial-scale=1,maximum-scale=1,user-scalable=no,viewport-fit=cover"/>
<meta name="apple-mobile-web-app-capable" content="yes">
<link rel="stylesheet" href="https://cdn.jsdelivr.net/npm/xterm@5.3.0/css/xterm.min.css"/>
<script src="https://cdn.jsdelivr.net/npm/xterm@5.3.0/lib/xterm.min.js"></script>
<script src="https://cdn.jsdelivr.net/npm/@xterm/addon-fit@0.10.0/lib/addon-fit.min.js"></script>
<style>
@font-face{font-family:'Glass TTY VT220';src:url('https://cdn.jsdelivr.net/gh/svofski/glasstty@master/Glass_TTY_VT220.ttf') format('truetype')}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%;width:100%;background:#0a0a0a;overflow:hidden;position:fixed}

/* Main app container */
#app{display:flex;flex-direction:column;height:100dvh;height:100vh;width:100%;padding:4px}

/* CRT Monitor frame */
#crt{
    flex:1;min-height:0;position:relative;
    border-radius:20px;
    background:linear-gradient(145deg,#1a1a1a,#0d0d0d);
    box-shadow:
        inset 0 0 80px rgba(0,0,0,0.8),
        0 0 20px rgba(0,255,65,0.1),
        0 0 60px rgba(0,255,65,0.05);
    overflow:hidden;
    border:3px solid #222;
}

/* Screen area with curvature effect */
#screen{
    position:absolute;
    top:8px;left:8px;right:8px;bottom:8px;
    border-radius:12px;
    overflow:hidden;
    background:#000800;
}

/* Terminal container */
#terminal{
    position:absolute;
    top:0;left:0;right:0;bottom:0;
    padding:4px;
}

.xterm{height:100%!important}
.xterm-screen{height:100%!important}
.xterm-viewport{overflow-y:hidden!important}

/* Phosphor cursor glow */
.xterm .xterm-cursor-block{
    background-color:#33ff33 !important;
    box-shadow:0 0 10px #33ff33,0 0 20px #22cc22,0 0 30px #119911 !important;
}
.xterm-cursor-blink{animation:phosphor-blink 0.6s steps(1) infinite !important}
@keyframes phosphor-blink{0%,45%{opacity:1}46%,100%{opacity:0.15}}

/* Scanlines overlay */
#scanlines{
    position:absolute;
    top:0;left:0;right:0;bottom:0;
    pointer-events:none;
    z-index:10;
    background:repeating-linear-gradient(
        0deg,
        transparent 0px,
        transparent 1px,
        rgba(0,0,0,0.3) 1px,
        rgba(0,0,0,0.3) 2px
    );
    animation:scanline-flicker 0.05s infinite;
}
@keyframes scanline-flicker{
    0%{opacity:0.9}
    50%{opacity:1}
    100%{opacity:0.9}
}

/* Screen glow/bloom effect */
#glow{
    position:absolute;
    top:0;left:0;right:0;bottom:0;
    pointer-events:none;
    z-index:11;
    background:radial-gradient(ellipse at center,transparent 0%,rgba(0,20,0,0.2) 80%,rgba(0,0,0,0.5) 100%);
    mix-blend-mode:multiply;
}

/* Phosphor flicker */
#flicker{
    position:absolute;
    top:0;left:0;right:0;bottom:0;
    pointer-events:none;
    z-index:9;
    background:transparent;
    animation:screen-flicker 0.1s infinite;
    opacity:0.03;
}
@keyframes screen-flicker{
    0%{background:rgba(0,255,0,0.02)}
    50%{background:transparent}
    100%{background:rgba(0,255,0,0.01)}
}

/* RGB subpixel effect on text (subtle) */
#terminal::before{
    content:'';
    position:absolute;
    top:0;left:0;right:0;bottom:0;
    background:repeating-linear-gradient(
        90deg,
        rgba(255,0,0,0.03) 0px,
        rgba(0,255,0,0.03) 1px,
        rgba(0,0,255,0.03) 2px,
        transparent 3px
    );
    pointer-events:none;
    z-index:5;
}

/* Input bar styled as part of CRT */
#input-bar{
    background:linear-gradient(180deg,#0d0d0d,#1a1a1a);
    padding:6px 8px;
    border-top:2px solid #222;
    display:flex;
    gap:6px;
    border-radius:0 0 8px 8px;
}

#input-wrapper{
    flex:1;
    position:relative;
    display:flex;
    align-items:center;
}

#prompt{
    color:#33ff33;
    font-family:'Glass TTY VT220',monospace;
    font-size:16px;
    padding:0 4px;
    text-shadow:0 0 8px #33ff33;
}

#input{
    flex:1;
    padding:8px 4px;
    font-size:16px;
    font-family:'Glass TTY VT220',monospace;
    background:transparent;
    border:none;
    color:#33ff33;
    outline:none;
    caret-color:transparent;
    text-shadow:0 0 8px #33ff33;
}

/* Custom blinking cursor that follows input */
#cursor{
    display:inline-block;
    width:10px;
    height:18px;
    background:#33ff33;
    animation:cursor-blink 0.6s steps(1) infinite;
    box-shadow:0 0 10px #33ff33,0 0 20px #22cc22;
    vertical-align:middle;
    margin-left:2px;
}
@keyframes cursor-blink{0%,45%{opacity:1}46%,100%{opacity:0.15}}

#send-btn{
    padding:8px 16px;
    background:linear-gradient(180deg,#1a1a1a,#0d0d0d);
    border:1px solid #33ff33;
    border-radius:4px;
    color:#33ff33;
    font-family:'Glass TTY VT220',monospace;
    font-weight:bold;
    text-shadow:0 0 8px #33ff33;
    cursor:pointer;
}
#send-btn:hover{background:#33ff33;color:#000;text-shadow:none}
#send-btn:active{transform:scale(0.98)}

/* Mobile quick buttons */
#controls{
    display:flex;
    flex-wrap:wrap;
    gap:3px;
    padding:4px;
    background:#0d0d0d;
}
.qbtn{
    flex:1;
    min-width:40px;
    padding:8px 2px;
    background:linear-gradient(180deg,#1a1a1a,#0d0d0d);
    border:1px solid #1a3a1a;
    border-radius:4px;
    color:#33ff33;
    font-family:'Glass TTY VT220',monospace;
    font-size:10px;
    text-align:center;
    text-shadow:0 0 5px #33ff33;
}
.qbtn:active{background:#33ff33;color:#000;text-shadow:none}

@media(min-width:600px){#controls{display:none}}
@media(max-width:599px){
    #crt{border-radius:8px}
    #screen{top:4px;left:4px;right:4px;bottom:4px;border-radius:6px}
}
</style>
</head>
<body>
<div id="app">
<div id="crt">
    <div id="screen">
        <div id="terminal"></div>
        <div id="flicker"></div>
        <div id="scanlines"></div>
        <div id="glow"></div>
    </div>
</div>
<div id="controls">
<div class="qbtn" ontouchend="cmd('n')">N</div>
<div class="qbtn" ontouchend="cmd('s')">S</div>
<div class="qbtn" ontouchend="cmd('e')">E</div>
<div class="qbtn" ontouchend="cmd('w')">W</div>
<div class="qbtn" ontouchend="cmd('look')">LOOK</div>
<div class="qbtn" ontouchend="cmd('inv')">INV</div>
<div class="qbtn" ontouchend="cmd('get')">GET</div>
<div class="qbtn" ontouchend="cmd('score')">SCORE</div>
</div>
<div id="input-bar">
<div id="input-wrapper">
    <span id="prompt">&gt;</span>
    <input type="text" id="input" autocomplete="off" autocapitalize="off" enterkeyhint="send"/>
    <span id="cursor"></span>
</div>
<button id="send-btn" onclick="sendInput()">SEND</button>
</div>
</div>

<script>
const matrixChars = "ﾊﾐﾋｰｳｼﾅﾓﾆｻﾜﾂｵﾘｱﾎﾃﾏｹﾒｴｶｷﾑﾕﾗｾﾈｽﾀﾇﾍ01234567890";
const isMobile = /iPhone|Android/i.test(navigator.userAgent) && window.innerWidth < 800;
const fontSize = isMobile ? (window.innerWidth < 380 ? 9 : 11) : (window.innerWidth < 1200 ? 12 : 14);

const term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontFamily: '"Glass TTY VT220", "Courier New", monospace',
    fontSize: fontSize,
    lineHeight: 1.1,
    scrollback: 1000,
    theme: {
        background: '#000800',
        foreground: '#33ff33',
        cursor: '#33ff33',
        cursorAccent: '#000000',
        selectionBackground: '#33ff3344'
    }
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));

// Track if intro has been shown this session
let introShown = sessionStorage.getItem('introShown') === 'true';
let socket = null;
let introComplete = false;

setTimeout(() => {
    fitAddon.fit();
    if (!introShown) {
        runIntro();
    } else {
        introComplete = true;
        connect();
    }
}, 100);

window.addEventListener('resize', () => fitAddon.fit());

const sleep = ms => new Promise(r => setTimeout(r, ms));
const green = '\x1b[32m', bright = '\x1b[92m', white = '\x1b[97m', dim = '\x1b[2m\x1b[32m', reset = '\x1b[0m';
const randChar = () => matrixChars[Math.floor(Math.random() * matrixChars.length)];

async function runIntro() {
    const cols = term.cols;
    const rows = term.rows;
    
    term.clear();
    
    // Initialize rain drops
    let drops = [];
    for (let x = 0; x < cols; x++) {
        drops[x] = { y: -Math.random() * rows * 2, speed: 0.4 + Math.random() * 0.4 };
    }
    
    // Banner text
    const banner = 'T H E   C O N S T R U C T';
    const bannerPad = Math.max(0, Math.floor((cols - banner.length) / 2));
    const lineW = Math.min(cols - 4, banner.length + 10);
    const linePad = Math.max(0, Math.floor((cols - lineW) / 2));
    const topLine = '═'.repeat(lineW);
    
    // Buffer for reveal
    let buffer = [];
    for (let y = 0; y < rows; y++) {
        buffer[y] = [];
        for (let x = 0; x < cols; x++) {
            buffer[y][x] = { char: ' ', style: '', revealed: false };
        }
    }
    
    // Place banner in buffer
    const bannerY = Math.floor(rows / 2) - 2;
    const bannerLines = [
        { y: bannerY - 3, text: ' '.repeat(linePad) + topLine, style: 'white' },
        { y: bannerY - 1, text: ' '.repeat(bannerPad) + banner, style: 'dim' },
        { y: bannerY, text: ' '.repeat(bannerPad) + banner, style: 'green' },
        { y: bannerY + 1, text: ' '.repeat(bannerPad) + banner, style: 'bright' },
        { y: bannerY + 2, text: ' '.repeat(bannerPad) + banner, style: 'green' },
        { y: bannerY + 3, text: ' '.repeat(bannerPad) + banner, style: 'dim' },
        { y: bannerY + 5, text: ' '.repeat(linePad) + topLine, style: 'white' }
    ];
    
    for (const bl of bannerLines) {
        if (bl.y >= 0 && bl.y < rows) {
            for (let x = 0; x < bl.text.length && x < cols; x++) {
                if (bl.text[x] !== ' ') {
                    buffer[bl.y][x] = { char: bl.text[x], style: bl.style, revealed: false };
                }
            }
        }
    }
    
    // Phase 1: Rain (2 seconds)
    for (let frame = 0; frame < 60; frame++) {
        term.write('\x1b[H');
        for (let y = 0; y < rows - 1; y++) {
            let line = '';
            for (let x = 0; x < cols; x++) {
                const d = drops[x].y - y;
                if (d >= 0 && d < 1) line += white + randChar() + reset;
                else if (d >= 1 && d < 3) line += bright + randChar() + reset;
                else if (d >= 3 && d < 8) line += green + randChar() + reset;
                else if (d >= 8 && d < 12) line += dim + randChar() + reset;
                else line += ' ';
            }
            term.writeln(line);
        }
        for (let x = 0; x < cols; x++) {
            drops[x].y += drops[x].speed;
            if (drops[x].y > rows + 12) drops[x].y = -Math.random() * 10;
        }
        await sleep(33);
    }
    
    // Phase 2: Rain reveals banner (4 seconds)
    for (let frame = 0; frame < 120; frame++) {
        term.write('\x1b[H');
        for (let y = 0; y < rows - 1; y++) {
            let line = '';
            for (let x = 0; x < cols; x++) {
                const cell = buffer[y][x];
                const d = drops[x].y - y;
                
                if (d >= 0 && d < 2) cell.revealed = true;
                
                if (cell.revealed && cell.char !== ' ') {
                    const c = cell.char;
                    const s = cell.style;
                    if (s === 'white') line += white + c + reset;
                    else if (s === 'bright') line += bright + c + reset;
                    else if (s === 'dim') line += dim + c + reset;
                    else line += green + c + reset;
                } else if (d >= 0 && d < 1) {
                    line += white + randChar() + reset;
                } else if (d >= 1 && d < 3) {
                    line += bright + randChar() + reset;
                } else if (d >= 3 && d < 6) {
                    line += green + randChar() + reset;
                } else {
                    line += ' ';
                }
            }
            term.writeln(line);
        }
        for (let x = 0; x < cols; x++) {
            drops[x].y += drops[x].speed;
            if (drops[x].y > rows + 7) drops[x].y = -Math.random() * 8;
        }
        await sleep(33);
    }
    
    // Phase 3: Clean banner display
    await sleep(300);
    term.clear();
    for (let y = 0; y < rows - 1; y++) {
        let line = '';
        for (let x = 0; x < cols; x++) {
            const cell = buffer[y][x];
            if (cell.char !== ' ') {
                const s = cell.style;
                if (s === 'white') line += white + cell.char + reset;
                else if (s === 'bright') line += bright + cell.char + reset;
                else if (s === 'dim') line += dim + cell.char + reset;
                else line += green + cell.char + reset;
            } else {
                line += ' ';
            }
        }
        term.writeln(line);
    }
    
    // Hold banner
    await sleep(2500);
    
    // Transition
    term.writeln('');
    term.writeln(green + '              [ Establishing connection... ]' + reset);
    await sleep(1000);
    
    term.clear();
    sessionStorage.setItem('introShown', 'true');
    introShown = true;
    introComplete = true;
    connect();
}

function connect() {
    if (socket && socket.readyState === WebSocket.OPEN) return;
    const proto = location.protocol === 'https:' ? 'wss:' : 'ws:';
    socket = new WebSocket(proto + '//' + location.host + '/ws');
    socket.onmessage = e => term.write(e.data);
    socket.onclose = () => {
        if (introComplete) {
            term.writeln('\r\n\x1b[31m[ Signal lost - Reconnecting... ]\x1b[0m');
            setTimeout(connect, 3000);
        }
    };
    socket.onerror = () => {};
}

document.addEventListener('visibilitychange', () => {
    if (!document.hidden && introComplete && (!socket || socket.readyState !== WebSocket.OPEN)) {
        connect();
    }
});

const input = document.getElementById('input');
const cursor = document.getElementById('cursor');

function sendInput() {
    const t = input.value.trim();
    if (t && socket && socket.readyState === WebSocket.OPEN) {
        socket.send(t + '\n');
        input.value = '';
    }
    input.focus();
}

function cmd(s) {
    if (socket && socket.readyState === WebSocket.OPEN) {
        socket.send(s + '\n');
    }
}

input.addEventListener('keydown', e => {
    if (e.key === 'Enter') {
        e.preventDefault();
        sendInput();
    }
});

// Keep input focused
input.focus();
document.addEventListener('click', () => input.focus());

// Reconnect on CRT click
document.getElementById('crt').addEventListener('click', () => {
    if (introComplete && (!socket || socket.readyState !== WebSocket.OPEN)) {
        connect();
    }
});
</script>
</body>
</html>`
