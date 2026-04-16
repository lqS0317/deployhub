package ws

import "sync"

// Hub 管理按 room 划分的 WebSocket 连接
type Hub struct {
	rooms      map[string]map[*Client]bool
	register   chan *subscription
	unregister chan *subscription
	broadcast  chan *message
	mu         sync.RWMutex
}

type subscription struct {
	room   string
	client *Client
}

type message struct {
	room    string
	payload []byte
}

// NewHub 创建 WebSocket Hub
func NewHub() *Hub {
	return &Hub{
		rooms:      make(map[string]map[*Client]bool),
		register:   make(chan *subscription, 256),
		unregister: make(chan *subscription, 256),
		broadcast:  make(chan *message, 256),
	}
}

// Run 启动 Hub 事件循环，需在单独的 goroutine 中运行
func (h *Hub) Run() {
	for {
		select {
		case sub := <-h.register:
			h.mu.Lock()
			if _, ok := h.rooms[sub.room]; !ok {
				h.rooms[sub.room] = make(map[*Client]bool)
			}
			h.rooms[sub.room][sub.client] = true
			h.mu.Unlock()

		case sub := <-h.unregister:
			h.mu.Lock()
			if clients, ok := h.rooms[sub.room]; ok {
				if _, exists := clients[sub.client]; exists {
					delete(clients, sub.client)
					close(sub.client.send)
					if len(clients) == 0 {
						delete(h.rooms, sub.room)
					}
				}
			}
			h.mu.Unlock()

		case msg := <-h.broadcast:
			h.mu.RLock()
			if clients, ok := h.rooms[msg.room]; ok {
				for client := range clients {
					select {
					case client.send <- msg.payload:
					default:
						// 发送缓冲区满，关闭慢客户端
						go h.Unregister(msg.room, client)
					}
				}
			}
			h.mu.RUnlock()
		}
	}
}

// Register 将客户端注册到指定 room
func (h *Hub) Register(room string, client *Client) {
	h.register <- &subscription{room: room, client: client}
}

// Unregister 将客户端从指定 room 移除
func (h *Hub) Unregister(room string, client *Client) {
	h.unregister <- &subscription{room: room, client: client}
}

// Broadcast 向指定 room 的所有客户端广播消息
func (h *Hub) Broadcast(room string, payload []byte) {
	h.broadcast <- &message{room: room, payload: payload}
}
