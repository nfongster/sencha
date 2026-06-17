package store

import (
	"sync"

	"sencha/backend/internal/session"
)

var global = newMemoryStore()

type memoryStore struct {
	mu       sync.RWMutex
	sessions map[string]*session.Session
}

func newMemoryStore() *memoryStore {
	return &memoryStore{
		sessions: make(map[string]*session.Session),
	}
}

func Get(id string) (*session.Session, bool) {
	global.mu.RLock()
	defer global.mu.RUnlock()
	sess, ok := global.sessions[id]
	return sess, ok
}

func Set(id string, sess *session.Session) {
	global.mu.Lock()
	defer global.mu.Unlock()
	global.sessions[id] = sess
}

func Reset() {
	global = newMemoryStore()
}
