package server

import (
	"sync"

	"github.com/victorvbello/gomcp/mcp/types"
)

//muxloggingLevelBySessionID
type muxloggingLevelBySessionID struct {
	mu sync.RWMutex
	m  map[string]types.LoggingLevel
}

func newMuxloggingLevelBySessionID() *muxloggingLevelBySessionID {
	return &muxloggingLevelBySessionID{
		m: make(map[string]types.LoggingLevel),
	}
}

func (xm *muxloggingLevelBySessionID) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]types.LoggingLevel)
	xm.mu.Unlock()
}

func (xm *muxloggingLevelBySessionID) Get(key string) (types.LoggingLevel, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxloggingLevelBySessionID) Set(key string, value types.LoggingLevel) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxloggingLevelBySessionID) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}
