package whatsapp

import (
	"sync"
	"time"
)

var (
	sentMessageCache   = make(map[string]time.Time)
	sentMessageCacheMu sync.Mutex
)

// MarkMessageAsSent marks a message ID as recently sent to prevent duplicate webhooks from echos.
func MarkMessageAsSent(messageID string) {
	sentMessageCacheMu.Lock()
	defer sentMessageCacheMu.Unlock()
	sentMessageCache[messageID] = time.Now()
}

// IsMessageRecentlySent checks if a message was recently sent via API.
func IsMessageRecentlySent(messageID string) bool {
	sentMessageCacheMu.Lock()
	defer sentMessageCacheMu.Unlock()
	t, ok := sentMessageCache[messageID]
	if !ok {
		return false
	}
	// If the message was sent more than 1 minute ago, ignore it (cleanup)
	if time.Since(t) > 1*time.Minute {
		delete(sentMessageCache, messageID)
		return false
	}
	return true
}

func init() {
	// Background cleanup of the cache
	go func() {
		for {
			time.Sleep(5 * time.Minute)
			sentMessageCacheMu.Lock()
			now := time.Now()
			for id, t := range sentMessageCache {
				if now.Sub(t) > 10*time.Minute {
					delete(sentMessageCache, id)
				}
			}
			sentMessageCacheMu.Unlock()
		}
	}()
}
