package once

import (
	"net/http"
	"testing"
	"testing/synctest"
	"time"
)

func TestOnce_SingleExecution(t *testing.T) {
	ttl := 10 * time.Hour
	once := NewUnmanaged(ttl)

	callCount := 0
	fn := func() Result {
		callCount++
		return func(w http.ResponseWriter, r *http.Request) error { return nil }
	}

	const (
		nonce1 = "nonce1"
		nonce2 = "nonce2"
	)

	// First call (should execute)
	_ = once.Resolve(nonce1, fn)
	if callCount != 1 {
		t.Errorf("Expected call count 1 after first call, got %d", callCount)
	}

	// Second call (should return cached result, not execute fn again)
	_ = once.Resolve(nonce1, fn)
	if callCount != 1 {
		t.Errorf("Expected call count 1 after second call, got %d", callCount)
	}

	// Test with a different nonce
	callCount = 0
	fn2 := func() Result {
		callCount++
		return func(w http.ResponseWriter, r *http.Request) error { return nil }
	}
	_ = once.Resolve(nonce2, fn2)
	if callCount != 1 {
		t.Errorf("Expected call count 1 for nonce2, got %d", callCount)
	}
}

// Test if auto TTL cleanup properly deletes entries from the cache
func TestOnce_TTL_Cleanup(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ctx := t.Context()
		ttl := 1 * time.Hour
		once := New(ctx, ttl)

		callCount := 0
		fn := func() Result {
			callCount++
			return func(w http.ResponseWriter, r *http.Request) error { return nil }
		}

		const nonce1 = "nonce1"

		// First call (should execute)
		once.Resolve("nonce1", fn)
		if callCount != 1 {
			t.Fatalf("Expected nonce1 to execute once, got %d", callCount)
		}

		// Wait for TTL to pass
		synctest.Wait()

		// Second call (should execute, as it should have been cleaned up)
		callCount = 0
		_ = once.Resolve("nonce1", fn)
		if callCount != 1 {
			t.Errorf("Expected nonce1 to execute again after TTL expiry, got %d", callCount)
		}
	})
}

// Test if Tick properly deletes entries from the cache
func TestOnce_Cleanup_Tick(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		ttl := 1 * time.Hour
		once := NewUnmanaged(ttl)

		// Execute nonce1
		once.Resolve("nonce1", func() Result { return func(w http.ResponseWriter, r *http.Request) error { return nil } })

		// Manually run cleanup logic (simulating the background process)
		synctest.Wait()
		once.Tick()

		// We cannot easily check internal state (maps) in a unit test, but we can test the outcome:
		// Attempting to resolve the nonce again should trigger execution.
		callCount := 0
		fn := func() Result {
			callCount++
			return func(w http.ResponseWriter, r *http.Request) error { return nil }
		}
		_ = once.Resolve("nonce1", fn)
		if callCount != 1 {
			t.Errorf("Tick failed to clean up nonce1, expected execution, got %d", callCount)
		}
	})
}
