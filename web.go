package main

import (
	"github.com/gorilla/websocket"
	"log"
	"net"
	"net/http"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

func startWebServer(w *World) {
	mux := http.NewServeMux()
	mux.HandleFunc("/", serveHome)
	mux.HandleFunc("/ws", handleWebSocket)

	log.Println("Web Portal active on http://0.0.0.0:8080")
	if err := http.ListenAndServe("0.0.0.0:8080", mux); err != nil {
		log.Fatal("Web server error:", err)
	}
}

func serveHome(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte(htmlClient))
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}
	defer ws.Close()

	tcpConn, err := net.Dial("tcp", "localhost:2323")
	if err != nil {
		ws.WriteMessage(websocket.TextMessage, []byte("Error: Could not connect to Matrix Construct\r\n"))
		return
	}
	defer tcpConn.Close()

	go func() {
		buf := make([]byte, 4096)
		for {
			n, err := tcpConn.Read(buf)
			if err != nil {
				return
			}
			if err := ws.WriteMessage(websocket.TextMessage, buf[:n]); err != nil {
				return
			}
		}
	}()

	for {
		_, message, err := ws.ReadMessage()
		if err != nil {
			break
		}
		if _, err := tcpConn.Write(message); err != nil {
			break
		}
	}
}

const htmlClient = `
<!DOCTYPE html>
<html>
<head>
    <title>The Construct</title>
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0, user-scalable=no" />
    <link rel="stylesheet" href="https://cdnjs.cloudflare.com/ajax/libs/xterm/3.14.5/xterm.min.css" />
    <script src="https://cdnjs.cloudflare.com/ajax/libs/xterm/3.14.5/xterm.min.js"></script>
    <script src="https://cdnjs.cloudflare.com/ajax/libs/xterm/3.14.5/addons/fit/fit.min.js"></script>
    <style>
        :root { --matrix-green: #00FF41; --dark-bg: #0D0208; }
        body { 
            background-color: var(--dark-bg); 
            margin: 0; 
            overflow: hidden; 
            font-family: 'Courier New', monospace;
            display: flex;
            flex-direction: column;
            height: 100vh;
        }
        
        /* CRT Effect */
        #screen-container {
            flex: 1;
            position: relative;
            border-bottom: 2px solid #333;
            padding: 5px;
            overflow: hidden;
        }
        
        /* Mobile Controls */
        #controls {
            height: 220px;
            background: #111;
            display: grid;
            grid-template-columns: 1fr 1.5fr 1fr;
            gap: 5px;
            padding: 10px;
            box-sizing: border-box;
        }

        .panel { display: grid; gap: 5px; }
        
        /* D-Pad */
        .dpad { grid-template-columns: repeat(3, 1fr); grid-template-rows: repeat(2, 1fr); }
        .btn {
            background: #222;
            border: 1px solid #444;
            color: var(--matrix-green);
            font-family: monospace;
            font-size: 14px;
            cursor: pointer;
            border-radius: 4px;
            text-transform: uppercase;
            display: flex;
            align-items: center;
            justify-content: center;
            user-select: none;
        }
        .btn:active { background: var(--matrix-green); color: #000; }
        
        /* Specific Button Layouts */
        #btn-n { grid-column: 2; grid-row: 1; }
        #btn-w { grid-column: 1; grid-row: 2; }
        #btn-s { grid-column: 2; grid-row: 2; }
        #btn-e { grid-column: 3; grid-row: 2; }

        .actions { grid-template-columns: 1fr 1fr; }
        .utility { grid-template-columns: 1fr; }

        /* Desktop adjustments */
        @media (min-width: 768px) {
            #controls { height: 0; padding: 0; overflow: hidden; } /* Hide controls on desktop */
        }
    </style>
</head>
<body>
    <div id="screen-container">
        <div id="terminal"></div>
    </div>

    <div id="controls">
        <!-- Movement -->
        <div class="panel dpad">
            <div class="btn" id="btn-n" onclick="cmd('north')">N</div>
            <div class="btn" id="btn-w" onclick="cmd('west')">W</div>
            <div class="btn" id="btn-s" onclick="cmd('south')">S</div>
            <div class="btn" id="btn-e" onclick="cmd('east')">E</div>
        </div>

        <!-- Actions -->
        <div class="panel actions">
            <div class="btn" onclick="cmd('look')">LOOK</div>
            <div class="btn" onclick="cmd('inv')">INV</div>
            <div class="btn" onclick="cmd('get all')">GET ALL</div>
            <div class="btn" onclick="cmd('score')">SCORE</div>
            <div class="btn" style="grid-column: span 2; border-color: #800; color: #f55;" onclick="cmd('kill')">ATTACK TARGET</div>
        </div>

        <!-- Utility -->
        <div class="panel utility">
            <div class="btn" onclick="cmd('help')">HELP</div>
            <div class="btn" onclick="cmd('who')">WHO</div>
            <div class="btn" onclick="cmd('save')">SAVE</div>
        </div>
    </div>

    <script>
        Terminal.applyAddon(fit);
        const term = new Terminal({
            cursorBlink: true,
            fontFamily: 'Courier New, monospace',
            fontSize: 16,
            theme: { background: '#0D0208', foreground: '#00FF41', cursor: '#00FF41' }
        });
        term.open(document.getElementById('terminal'));
        term.fit();

        window.onresize = function() { term.fit(); };

        const protocol = window.location.protocol === 'https:' ? 'wss://' : 'ws://';
        const socket = new WebSocket(protocol + window.location.host + '/ws');

        socket.onopen = () => { term.writeln('\x1b[37m[Signal established...]\x1b[0m'); };
        socket.onmessage = (event) => { term.write(event.data); };
        
        // Helper to send commands via buttons
        function cmd(str) {
            socket.send(str + '\n');
            term.focus();
        }

        // Keyboard Input
        let currentLine = "";
        term.on('key', (key, ev) => {
            const printable = !ev.altKey && !ev.altGraphKey && !ev.ctrlKey && !ev.metaKey;
            if (ev.keyCode === 13) { 
                term.write('\r\n'); 
                socket.send(currentLine + '\n'); 
                currentLine = ""; 
            } else if (ev.keyCode === 8) { 
                if (currentLine.length > 0) { 
                    currentLine = currentLine.slice(0, -1); 
                    term.write('\b \b'); 
                } 
            } else if (printable) { 
                currentLine += key; 
                term.write(key); 
            }
        });
    </script>
</body>
</html>
`
