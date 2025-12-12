// Web client with device-adaptive Matrix intro
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
	if origin == "" || Config.AllowedOrigins == "*" { return true }
	for _, a := range strings.Split(Config.AllowedOrigins, ",") {
		if strings.TrimSpace(a) == origin { return true }
	}
	return false
}

func filterTelnetIAC(data []byte) []byte {
	result := make([]byte, 0, len(data))
	for i := 0; i < len(data); {
		if data[i] == 255 && i+1 < len(data) {
			switch data[i+1] {
			case 255: result = append(result, 255); i += 2
			case 251, 252, 253, 254: i += 3
			case 250: i += 2; for i < len(data) && !(data[i] == 255 && i+1 < len(data) && data[i+1] == 240) { i++ }; i += 2
			default: i += 2
			}
		} else { result = append(result, data[i]); i++ }
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
	if ws == nil { return }
	defer ws.Close()
	tcpConn, err := net.Dial("tcp", "localhost:"+Config.TelnetPort)
	if err != nil { ws.WriteMessage(websocket.TextMessage, []byte("Connection failed\r\n")); return }
	defer tcpConn.Close()
	go func() {
		buf := make([]byte, 4096)
		for { n, e := tcpConn.Read(buf); if e != nil { return }; if f := filterTelnetIAC(buf[:n]); len(f) > 0 { if ws.WriteMessage(websocket.TextMessage, f) != nil { return } } }
	}()
	for { _, m, e := ws.ReadMessage(); if e != nil { break }; tcpConn.Write(m) }
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
@font-face{font-family:'Glass TTY VT220';src:url('https://raw.githubusercontent.com/svofski/glasstty/master/Glass_TTY_VT220.ttf') format('truetype')}
*{margin:0;padding:0;box-sizing:border-box}
html,body{height:100%;width:100%;background:#000;overflow:hidden;position:fixed}
#app{display:flex;flex-direction:column;height:100dvh;height:100vh;width:100%}
#crt{flex:1;min-height:0;position:relative;margin:2px;border-radius:6px;background:#020202;box-shadow:0 0 30px rgba(0,255,65,0.06);overflow:hidden}
#terminal{position:absolute;top:2px;left:2px;right:2px;bottom:2px}
.xterm{height:100%!important}.xterm-screen{height:100%!important}
.xterm .xterm-cursor-block{background-color:#33ff33 !important;box-shadow:0 0 8px #33ff33,0 0 16px #22dd22 !important}
.xterm-cursor-blink{animation:phosphor-blink 0.7s steps(1) infinite !important}
@keyframes phosphor-blink{0%,40%{opacity:1}41%,100%{opacity:0.1}}
#scanlines{position:absolute;top:0;left:0;right:0;bottom:0;pointer-events:none;z-index:10;background:repeating-linear-gradient(0deg,rgba(0,0,0,0.04) 0px,rgba(0,0,0,0.04) 1px,transparent 1px,transparent 2px)}
#input-bar{background:#050505;padding:4px;border-top:1px solid #111;display:flex;gap:4px}
#input{flex:1;padding:8px;font-size:16px;font-family:'Glass TTY VT220',monospace;background:#020202;border:1px solid #0a0;border-radius:3px;color:#0f0;outline:none;caret-color:#33ff33;animation:input-glow 0.7s steps(1) infinite}
@keyframes input-glow{0%,40%{box-shadow:inset 0 0 5px rgba(51,255,51,0.3)}41%,100%{box-shadow:inset 0 0 2px rgba(51,255,51,0.1)}}
#send-btn{padding:8px 12px;background:#0a0;border:none;border-radius:3px;color:#000;font-family:'Glass TTY VT220',monospace;font-weight:bold}
#controls{display:flex;flex-wrap:wrap;gap:2px;padding:3px;background:#030303}
.qbtn{flex:1;min-width:38px;padding:6px 2px;background:#080808;border:1px solid #181818;border-radius:3px;color:#0a0;font-family:'Glass TTY VT220',monospace;font-size:9px;text-align:center}
.qbtn:active{background:#0a0;color:#000}
@media(min-width:600px){#controls{display:none}}
</style>
</head>
<body>
<div id="app">
<div id="crt"><div id="terminal"></div><div id="scanlines"></div></div>
<div id="controls">
<div class="qbtn" ontouchend="cmd('n')">N</div><div class="qbtn" ontouchend="cmd('s')">S</div><div class="qbtn" ontouchend="cmd('e')">E</div><div class="qbtn" ontouchend="cmd('w')">W</div>
<div class="qbtn" ontouchend="cmd('look')">LOOK</div><div class="qbtn" ontouchend="cmd('inv')">INV</div><div class="qbtn" ontouchend="cmd('get')">GET</div><div class="qbtn" ontouchend="cmd('score')">SCORE</div>
</div>
<div id="input-bar"><input type="text" id="input" placeholder=">" autocomplete="off" autocapitalize="off" enterkeyhint="send"/><button id="send-btn" onclick="sendInput()">SEND</button></div>
</div>
<script>
// Payphone ASCII art - wide angle street scene
const artDesktop = [
"                                                                  ;;++++;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;",
"                                                                 ;+++;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;",
"                                                                 ;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;;; ;;;;;;;;;;;;;;;; ;;;;;;;;;",
"                                                                 ;+;                                                           ;;",
"                                                                  ;;;;  .. .   .                                          .   .;",
"                                                                  ;+;;              ;;;; ;;;;;;;;;;;;;;;;  ;;;;;;        ;.  ;;;;",
"                                                                  ;*+;:.; ;.  ;;;;+#@##*;*##+*@@#*+#*+*#*;+#@@@@#+;..    ;;:.:+;;+*+;",
"                                                                  ;*;;:.; ;   ;  ;;+;;;;.+;+;;;*;;;;;;;*;;++;;;;+*  ;    ;;; ;;;;**##*;",
"                                                                  ;+;;  ; ;      ;+;;; ;;+;+;+;*;;;+*;+;+;*+****;*;      ;.   .;;*+;;*+",
"                                                                  ;++;::; ;.. ...;++;;+;+;;+;;;*;;:+;*;+;.++;;;;;+;..  . ;;:..;;;+++;++",
"                                                                  ;+;;;;; ; ; ; ;;+*#*++;+*#;;***;;;*+*;  ;***#**+;     ;;;; ;:;;++;;;;",
"                                                                  ;+:;  ;;;;;;;;;;;  ; ;   ;      ;;;   ;          ;;;;;;; ; ;.; +**+;;",
"                                                                  ;*+;;;;;;+;:;+;:;::++;:++;.:;;;;;;+;::;;:;:.:;;.;;;::::;;;.;++; ;;;;;",
"                                                                  ;+;;;;;;..; ;.;;  ;;;  ..  .......;;;;;;;;; ;;; ;;:... .;;;;;;;",
"                                                                  ;+:    ;;+;;;++*+;  ;;++++;;;;;;;;;;;;;;;;;;++++;;;;;+;;;    ;+;;++;",
"                                                                  ;*;;;;;;:;.;;;;;;;;;;;;;;;;;+++*******+++++;;;;;;;;;;;;;+; ;;;;#;;;;;",
"                                                                  ;*;;:;;;: ;+;;**;;:;;;;***;*;;+*+*******+++***;;+++++++;;; ;:;;*;;;;",
"                                                                  ;+;;  ;;;;;.;;++;+++++;+++;*;         ;    *#*;;+******;.  ;.;**+;;",
"                                                                  ;++;:;; ;;;:;;*;;;;;;;;;**;+ ;;;;;;    . . *#*;+;      ;;. ;+;+*;;;;  ;",
"                                                                  ;+;;;;;  ;;;+;;+;;.:.;++;;+* ;;+**+;  ;;;;;*#;+*;;.    ;;;;;;;+*;;;;  ;",
"                                                                   ;.   ;;+; ;;;;+;;;+;;;;:+;+ ;;+++;+       +*+;+;      ; ;   ;;+ ;;;  ;",
"                                                                  ;+;;;;;;; .;+;;*;;;;; ++;;+*;;.;;;;+       *#;+*;      ;;;.;;;;+;;+;  ;",
"                                                                  ;+;;;;;;.  ;+;;+;;;;;;;;;;**;;;;;;;;  ;;   *#;**; ;;; ;+;;;;;;++ ;+;; ;",
"                                                                  ;+.     ;;; ;;;;;;;;;;;;:;+*      ;;       +*+;+;      ;     ;;; ;;;",
"                                                                  ;+;;;;;.;+;.;;+++++;;;;;;;+*;;;. :;        +#;+*;     ;;;; ;;;;+ ;;;; ;",
"                                                                  ;++;;+;;;;;;;;+++++++++++;+*;;;;  ;        +#;+*;     ;+;; ;; ;+ ++   ;",
"                                                                  ;+:  .; ;:; ;;+;++;++;;;:;++;;;::;;        +*+++;     ;;    . ;; ;+   ;",
"                                                                  ;++;:+; ;+;.;;++++++++++++++ ;;;;;         +**++      ;;;; ;; ;; ;+   ;",
"                                                                  ;+;;;;;;;;;;;;;;;;;;;;;;;;++  ;;;          +*;*+      ;;;;;;;;;;;;+;; ;",
"                                                                  ;;.   ; ;   ;;:....... . ;;+               ;+.;;       ;       ;;:+",
"                                                                  ;+;;;;;;;;;;;;;;+++;;;;;;;;;;;             ;*;++      ;+;;;;;;;;;++ ;",
"                                                                  ;++;;;;;;;;;;;+++***+;;;;;;;:;             ;*;++      ;+;;.;+ +;;++ ;",
"                                                                  ;+.  .;;;;;;;+#*++++;;;;;;;  ;;            ;+:+;      ;;   ;. .;;.;",
"                                                                  ;+;;;;;;; ;+;+;    ;;+;;;;;; ;;            ;+;+;      ;+;;;;;;;;;;+ ;",
"                                                                  ;++.;+;;; ;+++;     ;++;++;..;;            ;+++;      ;++;.;+.+;;++",
"                                                                  ;+: ;:;;  ;+;+      ;;;:;;;;.;;            ;+;+;      ;;.  ;. .;;.+",
"                                                                  ;++.;+;;  ;+++;     ;;;;+;;;;;;            ;+++;      ;;:; ;; :;;++ ; ;",
"                                                                  ;+;;;;;; ;;;+;;;    ;;++;   ;;;            ;+;+;      ;;;;.;; :;;++ ; ;",
"                                                                  ;;.;; ;;;;;   ;;                           ;;.;; ;;;; ;; ; ;.  ;;.;",
"                                                                  ;;;;;;;;;;;;;;;+                           ;+;;;;;++; ;;;;;;; .;;;; ;;;",
"                                                                  ;;+.:;;;;;::.;;+               ;+**+;      ;++; ;;;++ ;;+; ;; :;;:; ;.",
"                                                                  ;;: ;.;;;;;  ;;+               ;+++;;;     ;;;;  .;;; ;;.; ;.  ;;.; ;",
"                                                                  ;+;;;.;;;;.  ;;;    ;;+++;  ;; ;;++;;;     ;;.;  .;;  ;;.; ;;;.;;.; ;",
"                                                                  ;++.;:;.;:. .:;;    ;+;;++  ;+    ;;;;     ;;+; :;;;  ;;;..;+ :;;:; ;",
"                                                                  ;++;...  .   ;;;    ;+;;;;  ;;;;;   ;      ;;:.  .;;  ;;:; ;: .;;.; ;",
" ;;                         ;                                     ;++;.:; ..    +;    ;++++;                 ;;.. ..;;  ;;:; ;; .;;:; ;             ;",
"                    ;                                             ;++.:;;;;;;:.;;;;    ;;;;                  ;;+;;;;;;  ;;+:.;+.:;;:; ;.",
"                                                                  ;+;;;;;;;;;;;;                            ;*;;+;;;;;  ;;;;;;;; ;;.; ;",
];

// Mobile closeup payphone
const artMobile = [
"      ..                .......",
"      ..             . .,,..,..",
"      ..             ...,. ....",
"      ..             .....   ..",
"       .     .:::::, ..........",
"       .     .**##*+ ... . .,..",
"       .     .*#%%#*,....  ....",
"      ..   . .#%#*%*........  ,",
"      ..     .#%#:+#:  ....  ..",
"   .  ..     .%%*,:*,..  .....,",
"    . ..     .##;.:*:      .,.,",
"   .  ... .  .*:,.+*:.       .;,",
"   :,,,:,... ,#:.,+*,.  ....,:*;:",
" ,.+;:+*+:;;,,*;+***,. .::;+**%#*",
" :,+++#%*++;::+...,*:,.;+*#%%%%%#",
" ;:*+*#%#+*;:;;.. .+::,+##%%%%%%#",
" +;#**#%#;++;:+,,,:*::;+*#%%%%%%*",
" ;:+**##*;;::,*+****:;;+*##%%%%#*",
" . .,+##*;:..,;,. ,*:,.,:+*###%#+",
"   . *+*+**;,:;.. .+:,,:;;++++*+:",
"   . ::;:::,.:;.:::+,.,:,;;::;+:.",
" ,           .#####+      ,,,,:.",
" ;.......    .#####+   ........",
" ++++;++;+;;:;;++++;:;++;;;:;::,,"
];

const matrixChars = "ﾊﾐﾋｰｳｼﾅﾓﾆｻﾜﾂｵﾘｱﾎﾃﾏｹﾒｴｶｷﾑﾕﾗｾﾈｽﾀﾇﾍ01";

// Detect device
const isMobile = /iPhone|Android/i.test(navigator.userAgent) && window.innerWidth < 800;
console.log('Device:', isMobile ? 'mobile' : 'desktop', window.innerWidth, 'x', window.innerHeight);

// Use smaller font during intro for more detail, then resize for gameplay
const introFontSize = isMobile ? 8 : 8;
const gameFontSize = isMobile ? (window.innerWidth < 380 ? 8 : 10) : (window.innerWidth < 1200 ? 10 : 12);

const term = new Terminal({
    cursorBlink: true,
    cursorStyle: 'block',
    fontFamily: '"Glass TTY VT220", monospace',
    fontSize: introFontSize,
    lineHeight: 1.0,
    scrollback: 500,
    theme: { background: '#020202', foreground: '#33ff33', cursor: '#33ff33' }
});

const fitAddon = new FitAddon.FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));
setTimeout(() => { fitAddon.fit(); console.log('Terminal:', term.cols, 'x', term.rows); runIntro(); }, 150);
window.addEventListener('resize', () => fitAddon.fit());

let socket = null, introComplete = false;
const sleep = ms => new Promise(r => setTimeout(r, ms));
const green = '\x1b[32m', bright = '\x1b[92m', white = '\x1b[97m', dim = '\x1b[2m\x1b[32m', reset = '\x1b[0m';
const randChar = () => matrixChars[Math.floor(Math.random() * matrixChars.length)];

async function runIntro() {
    const art = isMobile ? artMobile : artDesktop;
    const cols = term.cols;
    const rows = term.rows;
    const artW = Math.max(...art.map(l => l.length));
    const artH = art.length;
    console.log('Art:', artW, 'x', artH, 'Terminal:', cols, 'x', rows);
    
    term.clear();
    
    // THE CONSTRUCT banner config
    const banner = 'T H E   C O N S T R U C T';
    const bannerPad = Math.max(0, Math.floor((cols - banner.length) / 2));
    const lineW = Math.min(cols - 4, banner.length + 10);
    const linePad = Math.max(0, Math.floor((cols - lineW) / 2));
    
    // Position art below banner - banner takes ~12 rows at top
    const bannerHeight = 12;
    const artOffsetX = Math.max(0, Math.floor((cols - artW) / 2));
    const artOffsetY = bannerHeight + Math.max(0, Math.floor((rows - bannerHeight - artH - 2) / 2));
    
    // Initialize combined buffer for banner + art
    let buffer = [];
    for (let y = 0; y < rows; y++) {
        buffer[y] = [];
        for (let x = 0; x < cols; x++) {
            buffer[y][x] = { char: ' ', style: 'art', revealed: false };
        }
    }
    
    // Place banner in buffer (top portion)
    const topLine = '═'.repeat(lineW);
    const bannerLines = [
        { y: 2, text: ' '.repeat(linePad) + topLine, style: 'white' },
        { y: 5, text: ' '.repeat(bannerPad) + banner, style: 'dim' },
        { y: 6, text: ' '.repeat(bannerPad) + banner, style: 'green' },
        { y: 7, text: ' '.repeat(bannerPad) + banner, style: 'bright' },
        { y: 8, text: ' '.repeat(bannerPad) + banner, style: 'green' },
        { y: 9, text: ' '.repeat(bannerPad) + banner, style: 'dim' },
        { y: 11, text: ' '.repeat(linePad) + topLine, style: 'white' }
    ];
    
    for (const bl of bannerLines) {
        for (let x = 0; x < bl.text.length && x < cols; x++) {
            if (bl.y < rows && bl.text[x] !== ' ') {
                buffer[bl.y][x] = { char: bl.text[x], style: bl.style, revealed: false };
            }
        }
    }
    
    // Place art in buffer (below banner)
    for (let y = 0; y < artH; y++) {
        const line = art[y] || '';
        for (let x = 0; x < line.length; x++) {
            const bx = artOffsetX + x;
            const by = artOffsetY + y;
            if (by < rows && bx < cols && line[x] !== ' ') {
                buffer[by][bx] = { char: line[x], style: 'art', revealed: false };
            }
        }
    }
    
    // Rain drops
    let drops = [];
    for (let x = 0; x < cols; x++) {
        drops[x] = { y: -Math.random() * rows * 2, speed: 0.3 + Math.random() * 0.4 };
    }
    
    // Phase 1: Pure rain for ~2 seconds (60 frames)
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
    
    // Phase 2: Rain reveals banner + art simultaneously (~6 seconds, 180 frames)
    for (let frame = 0; frame < 180; frame++) {
        term.write('\x1b[H');
        for (let y = 0; y < rows - 1; y++) {
            let line = '';
            for (let x = 0; x < cols; x++) {
                const cell = buffer[y][x];
                const d = drops[x].y - y;
                
                // Rain head reveals content as it passes
                if (d >= 0 && d < 2 && !cell.revealed) {
                    cell.revealed = true;
                }
                
                if (cell.revealed && cell.char !== ' ') {
                    // Show revealed character with appropriate style
                    const c = cell.char;
                    const s = cell.style;
                    if (s === 'white') line += white + c + reset;
                    else if (s === 'bright') line += bright + c + reset;
                    else if (s === 'dim') line += dim + c + reset;
                    else if (s === 'green') line += green + c + reset;
                    else if (s === 'art') {
                        // Art character coloring
                        if ('.,:'.includes(c)) line += dim + c + reset;
                        else if (';+*'.includes(c)) line += green + c + reset;
                        else if ('#%@'.includes(c)) line += bright + c + reset;
                        else line += green + c + reset;
                    } else line += green + c + reset;
                } else if (d >= 0 && d < 1) {
                    line += white + randChar() + reset;
                } else if (d >= 1 && d < 3) {
                    line += bright + randChar() + reset;
                } else if (d >= 3 && d < 7) {
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
    
    // Phase 3: Show final clean image (banner + art)
    await sleep(500);
    term.clear();
    for (let y = 0; y < rows - 1; y++) {
        let line = '';
        for (let x = 0; x < cols; x++) {
            const cell = buffer[y][x];
            if (cell.char !== ' ') {
                const c = cell.char;
                const s = cell.style;
                if (s === 'white') line += white + c + reset;
                else if (s === 'bright') line += bright + c + reset;
                else if (s === 'dim') line += dim + c + reset;
                else if (s === 'green') line += green + c + reset;
                else if (s === 'art') {
                    if ('.,:'.includes(c)) line += dim + c + reset;
                    else if (';+*'.includes(c)) line += green + c + reset;
                    else if ('#%@'.includes(c)) line += bright + c + reset;
                    else line += green + c + reset;
                } else line += green + c + reset;
            } else {
                line += ' ';
            }
        }
        term.writeln(line);
    }
    
    // 5-second pause to admire the full reveal
    await sleep(5000);
    
    // Transition to gameplay
    term.writeln('');
    term.writeln(green + '                    [ Connection established ]' + reset);
    
    await sleep(1000);
    
    // Resize font for gameplay
    term.options.fontSize = gameFontSize;
    fitAddon.fit();
    term.clear();
    
    introComplete = true;
    connect();
}

function connect() {
    if (socket && socket.readyState === WebSocket.OPEN) return;
    socket = new WebSocket((location.protocol === 'https:' ? 'wss://' : 'ws://') + location.host + '/ws');
    socket.onmessage = e => term.write(e.data);
    socket.onclose = () => { if (introComplete) { term.writeln('\x1b[31m[ Signal lost ]\x1b[0m'); setTimeout(connect, 3000); } };
}
document.addEventListener('visibilitychange', () => { if (!document.hidden && introComplete && (!socket || socket.readyState !== WebSocket.OPEN)) connect(); });

const input = document.getElementById('input');
function sendInput() { const t = input.value.trim(); if (t && socket && socket.readyState === WebSocket.OPEN) { term.write(t + '\r\n'); socket.send(t + '\n'); input.value = ''; } input.focus(); }
function cmd(s) { if (socket && socket.readyState === WebSocket.OPEN) { term.write(s + '\r\n'); socket.send(s + '\n'); } }
input.addEventListener('keydown', e => { if (e.key === 'Enter') { e.preventDefault(); sendInput(); } });

let line = '';
term.onKey(({ key, domEvent }) => {
    if (window.innerWidth < 600 || !introComplete || !socket || socket.readyState !== WebSocket.OPEN) return;
    if (domEvent.keyCode === 13) { term.write('\r\n'); socket.send(line + '\n'); line = ''; }
    else if (domEvent.keyCode === 8) { if (line.length > 0) { line = line.slice(0, -1); term.write('\b \b'); } }
    else if (!domEvent.altKey && !domEvent.ctrlKey && !domEvent.metaKey) { line += key; term.write(key); }
});
document.getElementById('crt').addEventListener('click', () => { if (introComplete) connect(); });
</script>
</body>
</html>`
