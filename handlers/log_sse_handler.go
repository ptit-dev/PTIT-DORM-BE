package handlers

// Khởi động tail log file qua service, truyền callback là LogHub.Broadcast
import (
	"Backend_Dorm_PTIT/service"
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
)

// Khởi động tail log file qua service, truyền callback là LogHub.Broadcast
func StartTailLogFile(logPath string) {
	service.StartTailLogFile(context.Background(), logPath, service.GetLogHub().Broadcast)
}

// SSE endpoint cho admin
func StreamLogSSE(c *gin.Context) {

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Flush()

	ch := make(chan string, 100)
	hub := service.GetLogHub()
	hub.AddClient(ch)
	defer hub.RemoveClient(ch)

	// Gửi log mới cho client
	for {
		select {
		case line, ok := <-ch:
			if !ok {
				return
			}
			fmt.Fprintf(c.Writer, "data: %s\n\n", line)
			c.Writer.Flush()
		case <-c.Request.Context().Done():
			return
		}
	}
}
