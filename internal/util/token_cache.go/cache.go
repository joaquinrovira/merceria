package tokencache

import (
	"context"
	"log"
	"merceria/internal/util/cache"
)

type Cache = *cache.Cache[string, *AutorefreshingToken]

func New(ctx context.Context) *cache.Cache[string, *AutorefreshingToken] {
	googleUserTokenCache := cache.New[string, *AutorefreshingToken]()
	go googleUserTokenCache.Start()
	go func() {
		<-ctx.Done()
		googleUserTokenCache.Stop()
	}()

	googleUserTokenCache.OnEviction(func(_ context.Context, reason cache.EvictionReason, item *cache.Item[string, *AutorefreshingToken]) {
		cached := item.Value()
		if cached == nil {
			return
		}
		cached.Cancel()
		log.Printf("Token for user %s evicted", item.Key())
	})

	return googleUserTokenCache
}
