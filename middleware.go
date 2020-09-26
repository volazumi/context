package context

import (
	"fmt"
	"sync"
)

var (
	registerMutex = sync.Mutex{}
	middlewares   map[string]func(*Context) error
)

//RegisterMiddleware will be used by middlewares
func RegisterMiddleware(name string, fn func(*Context) error) error {
	registerMutex.Lock()
	defer registerMutex.Unlock()
	_, exists := middlewares[name]
	if exists {
		return fmt.Errorf("middleware %s already exists", name)
	}
	middlewares[name] = fn
	return nil
}
