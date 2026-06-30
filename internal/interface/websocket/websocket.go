package websocket

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"sync"
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

type client struct {
	ws   *WSSystem
	conn *websocket.Conn
	send chan WSMessage
}

// writePump handles writing messages to the websocket client.
// Launching a dedicated writer per client avoids blocking the main broadcast hub.
func (c *client) writePump() {
	defer func() {
		err := c.conn.Close()
		if err != nil {
			log.Printf("failed to close websocket: %v", err)
		}
	}()
	for msg := range c.send {
		if err := c.conn.WriteJSON(msg); err != nil {
			return
		}
	}
}

type WSSystem struct {
	clients     map[*client]bool
	register    chan *client
	unregister  chan *client
	broadcastCh chan WSMessage

	history   []WSMessage
	historyMu sync.RWMutex
}

func NewWSSystem(port string) *WSSystem {
	ws := &WSSystem{
		clients:     make(map[*client]bool),
		register:    make(chan *client),
		unregister:  make(chan *client),
		broadcastCh: make(chan WSMessage, 256),
	}

	go ws.run()

	http.HandleFunc("/logo.png", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		_, err := w.Write(logoBytes)
		if err != nil {
			log.Printf("failed to write logo: %v", err)
		}
	})
	http.HandleFunc("/echo", ws.handleWebSocket)
	http.HandleFunc("/", ws.handleHome)

	go func() {
		log.Fatal(http.ListenAndServe("localhost:"+port, nil))
	}()

	return ws
}

// run handles the lifecycle, registration and message broadcasting for all connections.
func (ws *WSSystem) run() {
	for {
		select {
		case client := <-ws.register:
			ws.clients[client] = true
		case client := <-ws.unregister:
			if _, ok := ws.clients[client]; ok {
				delete(ws.clients, client)
				close(client.send)
			}
		case msg := <-ws.broadcastCh:
			ws.historyMu.Lock()
			ws.history = append(ws.history, msg)
			ws.historyMu.Unlock()

			for client := range ws.clients {
				select {
				case client.send <- msg:
				default:
					close(client.send)
					delete(ws.clients, client)
					err := client.conn.Close()
					if err != nil {
						log.Printf("failed to close websocket: %v", err)
					}
				}
			}
		}
	}
}

func (ws *WSSystem) handleHome(w http.ResponseWriter, r *http.Request) {
	err := homeTemplate.Execute(w, "ws://"+r.Host+"/echo")
	if err != nil {
		http.Error(w, "Error executing template", http.StatusInternalServerError)
	}
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

	// 1. Replay history to the new client before registering.
	// This avoids any concurrent writes from other goroutines while history is replayed.
	ws.historyMu.RLock()
	msgs := make([]WSMessage, len(ws.history))
	copy(msgs, ws.history)
	ws.historyMu.RUnlock()

	for _, msg := range msgs {
		if err := c.WriteJSON(msg); err != nil {
			log.Println("write history err:", err)
			err = c.Close()
			if err != nil {
				log.Printf("failed to close websocket: %v", err)
			}
			return
		}
	}

	client := &client{
		ws:   ws,
		conn: c,
		send: make(chan WSMessage, 256),
	}
	ws.register <- client

	// Start writing pump for this client
	go client.writePump()

	defer func() {
		ws.unregister <- client
	}()

	// Keep connection alive and read incoming messages
	for {
		_, _, err := c.ReadMessage()
		if err != nil {
			break
		}
	}
}

func (ws *WSSystem) broadcast(msg WSMessage) {
	ws.broadcastCh <- msg
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
