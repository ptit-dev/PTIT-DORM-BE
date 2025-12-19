package service

import (
	"context"
	"time"

	"github.com/hpcloud/tail"
)

// ctx: context để dừng tail log khi client disconnect
func StartTailLogFile(ctx context.Context, logPath string, onLine func(string)) {
	done := make(chan struct{})
	go func() {
		defer close(done)
		for {
			t, err := tail.TailFile(logPath, tail.Config{Follow: true, ReOpen: true, Poll: true})
			if err != nil {
				time.Sleep(2 * time.Second)
				continue
			}
			for {
				select {
				case <-ctx.Done():
					t.Stop()
					return
				case line, ok := <-t.Lines:
					if !ok {
						return
					}
					if line != nil {
						onLine(line.Text)
					}
				}
			}
		}
	}()
}
