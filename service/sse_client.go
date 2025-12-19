package service

import "sync"

// LogHub quản lý các client SSE (singleton)
type LogHub struct {
	clients map[chan string]struct{}
	mu      sync.Mutex
}

var logHubInstance *LogHub
var once sync.Once

func GetLogHub() *LogHub {
	once.Do(func() {
		logHubInstance = &LogHub{
			clients: make(map[chan string]struct{}),
		}
	})
	return logHubInstance
}

func (h *LogHub) Broadcast(line string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	for ch := range h.clients {
		select {
		case ch <- line:
		default:
			// Nếu client không đọc kịp, bỏ qua để tránh block
		}
	}
}

func (h *LogHub) AddClient(ch chan string) {
	h.mu.Lock()
	h.clients[ch] = struct{}{}
	h.mu.Unlock()
}

func (h *LogHub) RemoveClient(ch chan string) {
	h.mu.Lock()
	delete(h.clients, ch)
	h.mu.Unlock()
	close(ch)
}