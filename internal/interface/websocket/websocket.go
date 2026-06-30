package websocket

import (
	_ "embed"
	"flag"
	"html/template"
	"log"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"tiercelieux-llm-go/internal/domain"

	"github.com/gorilla/websocket"
)

//go:embed templates/index.html
var indexHTML string

//go:embed templates/logo.png
var logoBytes []byte

var homeTemplate = template.Must(template.New("").Parse(indexHTML))

type PlayerInfo struct {
	Name        string `json:"name"`
	Temperament string `json:"temperament"`
	Role        string `json:"role"`
	IsAlive     bool   `json:"isAlive"`
}

type SummaryPayload struct {
	Turn    int          `json:"turn"`
	Phase   string       `json:"phase"` // "day" or "night"
	Players []PlayerInfo `json:"players"`
}

type TalkPayload struct {
	Content string `json:"content"`
}

type VotePayload struct {
	Voter  string `json:"voter"`
	Target string `json:"target"`
}

type VictimPayload struct {
	Victim string `json:"victim"`
	Role   string `json:"role"`
}

type VictoryPayload struct {
	Winner string `json:"winner"`
}

type WSMessage struct {
	Type    string `json:"type"`
	Payload any    `json:"payload,omitempty"`
}

type WSSystem struct {
	clients   map[*websocket.Conn]bool
	clientsMu sync.Mutex
	history   []WSMessage
	historyMu sync.Mutex
}

func NewWSSystem(port string) *WSSystem {
	flag.Parse()
	log.SetFlags(0)

	ws := &WSSystem{
		clients: make(map[*websocket.Conn]bool),
	}

	http.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(logoBytes)
	})
	http.HandleFunc("/echo", ws.handleWebSocket)
	http.HandleFunc("/", ws.handleHome)

	go func() {
		log.Fatal(http.ListenAndServe("localhost:"+port, nil))
	}()

	// Graceful shutdown on termination signals (e.g. Ctrl+C, SIGTERM, Ctrl+D if shell exits)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigChan
		log.Println("Shutdown signal received, closing WebSocket connections...")
		ws.Close()
		os.Exit(0)
	}()

	return ws
}

func (ws *WSSystem) Close() {
	ws.clientsMu.Lock()
	conns := make([]*websocket.Conn, 0, len(ws.clients))
	for client := range ws.clients {
		conns = append(conns, client)
	}
	ws.clientsMu.Unlock()

	for _, client := range conns {
		// Send Close frame to client
		err := client.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, "Server shutting down"),
		)
		if err != nil {
			log.Println("Error sending WS close message:", err)
		}
		client.Close()
	}
}

func (ws *WSSystem) handleHome(w http.ResponseWriter, r *http.Request) {
	homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (ws *WSSystem) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Print("upgrade:", err)
		return
	}

	ws.clientsMu.Lock()
	ws.clients[c] = true
	ws.clientsMu.Unlock()

	defer func() {
		ws.clientsMu.Lock()
		delete(ws.clients, c)
		ws.clientsMu.Unlock()
		c.Close()
	}()

	// Replay history to the new client so they see everything that happened
	ws.historyMu.Lock()
	for _, msg := range ws.history {
		if err := c.WriteJSON(msg); err != nil {
			log.Println("write history err:", err)
			ws.historyMu.Unlock()
			return
		}
	}
	ws.historyMu.Unlock()

	// Keep connection alive
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (ws *WSSystem) broadcast(msg WSMessage) {
	// Store in history
	ws.historyMu.Lock()
	ws.history = append(ws.history, msg)
	ws.historyMu.Unlock()

	// Send to all clients
	ws.clientsMu.Lock()
	defer ws.clientsMu.Unlock()
	for client := range ws.clients {
		if err := client.WriteJSON(msg); err != nil {
			log.Printf("error writing to client: %v", err)
			client.Close()
			delete(ws.clients, client)
		}
	}
}

func (w *WSSystem) DisplaySummary(game *domain.Game) {
	players := make([]PlayerInfo, 0, len(game.Villagers)+len(game.Werewolves)+len(game.Deceased))

	// Add villagers
	for _, p := range game.Villagers {
		players = append(players, PlayerInfo{
			Name:        p.Name(),
			Temperament: p.Temperament().String(),
			Role:        p.Role().String(),
			IsAlive:     true,
		})
	}
	// Add werewolves
	for _, p := range game.Werewolves {
		players = append(players, PlayerInfo{
			Name:        p.Name(),
			Temperament: p.Temperament().String(),
			Role:        p.Role().String(),
			IsAlive:     true,
		})
	}
	// Add deceased players
	for _, p := range game.Deceased {
		players = append(players, PlayerInfo{
			Name:        p.Name(),
			Temperament: p.Temperament().String(),
			Role:        p.Role().String(),
			IsAlive:     false,
		})
	}

	phase := "day"
	if game.GameState == domain.GameStateNight {
		phase = "night"
	}

	payload := SummaryPayload{
		Turn:    game.Turn,
		Phase:   phase,
		Players: players,
	}

	w.broadcast(WSMessage{
		Type:    "summary",
		Payload: payload,
	})
}

func (w *WSSystem) Display(content string) {
	w.broadcast(WSMessage{
		Type:    "talk",
		Payload: TalkPayload{Content: content},
	})
}

func (w *WSSystem) DisplayVillagerVictory() {
	w.broadcast(WSMessage{
		Type:    "victory",
		Payload: VictoryPayload{Winner: "villagers"},
	})
}

func (w *WSSystem) DisplayWerewolfVictory() {
	w.broadcast(WSMessage{
		Type:    "victory",
		Payload: VictoryPayload{Winner: "werewolves"},
	})
}

func (w *WSSystem) DisplayVictim(victim string, role domain.Role) {
	w.broadcast(WSMessage{
		Type: "victim",
		Payload: VictimPayload{
			Victim: victim,
			Role:   role.String(),
		},
	})
}

func (w *WSSystem) DisplayVote(voter string, target string) {
	w.broadcast(WSMessage{
		Type: "vote",
		Payload: VotePayload{
			Voter:  voter,
			Target: target,
		},
	})
}
