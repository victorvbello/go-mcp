package client

import (
	"sync"

	"github.com/victorvbello/gomcp/mcp/types"
)

//Validate the structuredContent of the tools
type clientToolOutputValidator func(data map[string]interface{}) (bool, error)

func compileOutputToolSchema(outSchema *types.ToolOutputSchema) (clientToolOutputValidator, error) {
	val := func(data map[string]interface{}) (bool, error) {
		return true, nil
	}
	return val, nil
}

//muxCachedToolOutputValidators
type muxCachedToolOutputValidators struct {
	mu sync.RWMutex
	m  map[string]clientToolOutputValidator
}

func newMuxCachedToolOutputValidators() *muxCachedToolOutputValidators {
	return &muxCachedToolOutputValidators{
		m: make(map[string]clientToolOutputValidator),
	}
}

func (xm *muxCachedToolOutputValidators) Clear() {
	xm.mu.Lock()
	xm.m = make(map[string]clientToolOutputValidator)
	xm.mu.Unlock()
}

func (xm *muxCachedToolOutputValidators) Get(key string) (clientToolOutputValidator, bool) {
	xm.mu.RLock()
	val, ok := xm.m[key]
	xm.mu.RUnlock()
	return val, ok
}

func (xm *muxCachedToolOutputValidators) Set(key string, value clientToolOutputValidator) {
	xm.mu.Lock()
	xm.m[key] = value
	xm.mu.Unlock()
}

func (xm *muxCachedToolOutputValidators) Delete(key string) {
	xm.mu.Lock()
	delete(xm.m, key)
	xm.mu.Unlock()
}
