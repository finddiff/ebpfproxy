package webapi

import (
	"fmt"
	"strings"
	"sync"

	"github.com/sirupsen/logrus"
)

type LogrusHook struct {
	buffer *LogBuffer
	server *Server
	mu     sync.Mutex
	levels []logrus.Level
}

func NewLogrusHook(server *Server, levels []logrus.Level) *LogrusHook {
	return &LogrusHook{
		buffer: server.logBuffer,
		server: server,
		levels: levels,
	}
}

func (h *LogrusHook) Levels() []logrus.Level {
	return h.levels
}

func (h *LogrusHook) Fire(entry *logrus.Entry) error {
	h.mu.Lock()
	defer h.mu.Unlock()

	msg := entry.Message
	if len(entry.Data) > 0 {
		fields := make([]string, 0, len(entry.Data))
		for k, v := range entry.Data {
			fields = append(fields, fmt.Sprintf("%s=%v", k, v))
		}
		msg = msg + "  " + strings.Join(fields, " ")
	}

	level := strings.ToUpper(entry.Level.String()[:4])
	logText := fmt.Sprintf("%s %s", level, msg)
	h.buffer.Append(logText)

	if h.server.hub != nil {
		h.server.BroadcastJSON("log_entry", map[string]interface{}{
			"time": entry.Time.Format("15:04:05"),
			"text": logText,
		})
	}

	return nil
}
