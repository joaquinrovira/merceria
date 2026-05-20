package once

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// Result defines the function signature for the operation that should be
// executed once.
type Result = func(w http.ResponseWriter, r *http.Request) error

// Once manages a set of operations, ensuring that each unique nonce is
// executed only once within the defined TTL period.
type Once struct {
	ttl time.Duration
	mtx sync.Mutex

	mtxs     map[string]*sync.Mutex
	cleanup  map[string]*time.Timer
	resolver map[string]Result
}

// New creates and returns a new Once instance that automatically handles
// background cleanup of expired nonces based on the provided TTL.
func New(ctx context.Context, ttl time.Duration) *Once {
	once := NewUnmanaged(ttl)

	go func() {
		tout := max(ttl/4, 200*time.Millisecond)
		for ctx.Err() == nil {
			once.Tick()
			select {
			case <-time.After(tout):
			case <-ctx.Done():
			}
		}
	}()

	return once
}

// NewUnmanaged creates a Once instance without starting the automatic TTL
// cleanup routine.
func NewUnmanaged(ttl time.Duration) *Once {
	return &Once{
		mtxs:     map[string]*sync.Mutex{},
		cleanup:  map[string]*time.Timer{},
		resolver: map[string]Result{},
	}
}

// Resolve ensures that the function returned by fn is executed only once for
// the given nonce. It returns the resulting Result function. If the nonce has
// been executed recently, the cached Result is returned immediately.
func (o *Once) Resolve(nonce string, fn func() Result) (r Result) {
	o.mtx.Lock()
	mtx := o.mtxs[nonce]
	if mtx == nil {
		mtx = &sync.Mutex{}
		o.mtxs[nonce] = mtx
		mtx.Lock()
		o.mtx.Unlock()
	} else {
		o.mtx.Unlock()
		mtx.Lock()
	}
	defer mtx.Unlock()

	if v, ok := o.resolver[nonce]; ok {
		return v
	}

	v := fn()
	o.resolver[nonce] = v

	o.mtx.Lock()
	defer o.mtx.Unlock()
	o.cleanup[nonce] = time.NewTimer(o.ttl)

	return v
}

// Tick is called periodically to clean up expired nonces from the cache,
// freeing up resources.
func (o *Once) Tick() {
	o.mtx.Lock()
	defer o.mtx.Unlock()

	for nonce, t := range o.cleanup {
		select {
		default:
		case <-t.C:
			delete(o.resolver, nonce)
			delete(o.cleanup, nonce)
			delete(o.mtxs, nonce)
		}
	}
}
