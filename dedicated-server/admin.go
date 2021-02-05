package main

import (
	"fmt"
	"log"
	"net/http"
	"time"
)

// AdminServer enables a web interface for viewing and kicking players.
type AdminServer struct {
	server     *http.Server
	room       kicker
	playerList playerlistSubscriber
}

// Satisfied by Room. Allows dependency injection for testing.
type kicker interface {
	Kick(string) error
}

// Satisfied by PlayerList. Allows dependency injection for testing.
type playerlistSubscriber interface {
	Subscribe(chan string)
	Unsubscribe(chan string)
}

// NewAdminServer creates a new AdminServer that will run a http server on the
// specified port.
func NewAdminServer(port int, room kicker, playerList playerlistSubscriber) *AdminServer {
	s := &AdminServer{nil, room, playerList}
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write([]byte(adminTemplate))
	})
	mux.HandleFunc("/kick/", s.handleKick)
	mux.HandleFunc("/playerlist", s.handlePlayerlist)
	s.server = &http.Server{
		Addr:    fmt.Sprintf("localhost:%d", port),
		Handler: mux,
	}
	go func() {
		if err := s.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Printf("Admin server failed: %v", err)
		}
	}()
	return s
}

// Close shuts down the AdminServer.
func (s *AdminServer) Close() error {
	return s.server.Close()
}

// Handles a /playerlist request.
func (s *AdminServer) handlePlayerlist(w http.ResponseWriter, r *http.Request) {
	f, ok := w.(http.Flusher)
	if !ok {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Set headers as necessary to enable Server-sent Events.
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")
	c := make(chan string, 1)
	s.playerList.Subscribe(c)
	defer s.playerList.Unsubscribe(c)
	t := time.NewTicker(10 * time.Second)
	defer t.Stop()
	for {
		select {
		case playerlist := <-c:
			if _, err := fmt.Fprintf(w, "data: %s\n\n", playerlist); err != nil {
				log.Printf("playerlist write failure: %v", err)
				return
			}
			f.Flush()
		case <-t.C:
			// Send data to keep the connection alive. A line starting with
			// a colon is ignored.
			if _, err := w.Write([]byte(":\n")); err != nil {
				log.Printf("playerlist keep-alive failure: %v", err)
				return
			}
			f.Flush()
		case <-r.Context().Done():
			return
		}
	}
}

// Handles a /kick request.
func (s *AdminServer) handleKick(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Cache-Control", "no-cache")
	userName := r.URL.Path[len("/kick/"):]
	if err := s.room.Kick(userName); err != nil {
		log.Printf("Kick failed: %v", err)
		w.WriteHeader(http.StatusNotFound)
	}
}

// TODO(bmclarnon): Should we add game-specific logic to display what items
// each player has collected?
const adminTemplate = `<!doctype html>
<title>BizHawk Co-op Admin</title>
<style>
body {
	color: #444;
	font-family: monospace;	
}
.header span {
	box-sizing: border-box;
	display: inline-block;
	min-width: 3ch;
	margin-right: 2ch;
}
.playerlist ul {
	display: inline-block;
	list-style: none;
	margin: 0.5em 0;
	padding: 0;
	min-width: 24ch;
}
.playerlist li {
	display: block;
	line-height: 1.5em;
	padding: 0 0.5ch 0 4.5ch;
}
.playerlist li::before {
	display: inline-block;
	content: "[ ]";
	width: 4ch;
	margin-left: -4ch;
}
.playerlist li.r::before { content: "[X]"; }
.playerlist li:nth-child(even) { background: #FFF; }
.playerlist li:nth-child(odd) { background: #EEE; }
.playerlist li:hover { background: #CCC; }
.playerlist li span {
	display: inline-block;
	min-width: 4ch;
	text-align: right;
}
.playerlist li button {
	cursor: pointer;
	background: none;
	border: 0;
	padding: 0;
	margin: 0;
	float: right;
}
.playerlist li:hover button::after, .playerlist li button:focus::after {
 	content: "\1F97E";
	color: transparent;
	text-shadow: 0 0 0 #444;
}
</style>
<h1>BizHawk Co-op Admin</h1>
<div class="header">Players: <span id="pc">0</span>Ready: <span id="rc">0</span></div>
<div class="playerlist"><ul id="pl"></ul></div>
<script>
(function() {
	const pl = document.getElementById('pl');
	const pc = document.getElementById('pc');
	const rc = document.getElementById('rc');
	const parse = function(data) {
		if (data == "") return [];
		const m = new Map();
		for (let e of data.split(',')) {
			if (e == "") continue;
			const parts = e.split(':');
			if (!m.has(parts[1])) {
				m.set(parts[1], {name: parts[1]});
			}
			m.get(parts[1])[parts[2]] = parts[3];
		}
		return Array.from(m.values()).sort((x, y) => x.num - y.num);
	};
	const e = new EventSource('/playerlist');
	e.onmessage = function(e) {
		const players = parse(e.data)
		pl.innerHTML = '';
		var ready = 0;
		for (let p of players) {
			const li = document.createElement('li');
			if (p.status != 'Unready') {
				++ready;
				li.classList.add('r');
			}
			const b = document.createElement('button');
			b.type = 'button';
			b.onclick = function() {
				fetch("/kick/" + encodeURI(p.name))
			};
			const span = document.createElement('span');
			span.append('P' + p.num)
			li.append(span, ': ' + p.name, b);
			pl.append(li);
		}
		pc.textContent = players.length;
		rc.textContent = ready;
	};
})()
</script>
`
