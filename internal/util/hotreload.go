package util

import (
	"context"
	"log"
	"os"
	"time"
)

func ShittyFsNotify(ctx context.Context, fs *os.Root, path string) <-chan struct{} {
	ch := make(chan struct{})
	go func() {
		prev, _ := fs.Stat(path)
		for {
			select {
			case <-ctx.Done():
				return
			case <-time.After(200 * time.Millisecond):
			}

			next, err := fs.Stat(path)
			if err != nil {
				continue
			}

			changed := false
			changed = changed || next.ModTime() != prev.ModTime()
			changed = changed || next.Size() != prev.Size()

			if changed {
				prev = next
				ch <- struct{}{}
			}
		}
	}()
	return ch
}

func ReloadingFile(ctx context.Context, fs *os.Root, path string) func() (data []byte, err error) {
	data, err := fs.ReadFile(path)
	go func() {
		for range ShittyFsNotify(ctx, fs, path) {
			log.Printf("%s changed, reloading", path)
			data, err = fs.ReadFile(path)
		}
	}()
	return func() ([]byte, error) { return data, err }
}
