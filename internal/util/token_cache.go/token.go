package tokencache

import (
	"context"

	"golang.org/x/oauth2"
)

type AutorefreshingToken struct {
	oauth2.TokenSource
	cancelFunc context.CancelFunc
}

func NewToken(ctx context.Context, token *oauth2.Token, oauthConfig *oauth2.Config) *AutorefreshingToken {
	ctx, cancel := context.WithCancel(ctx)
	return &AutorefreshingToken{
		TokenSource: oauthConfig.TokenSource(ctx, token),
		cancelFunc:  cancel,
	}
}

func (ts *AutorefreshingToken) Cancel() {
	ts.cancelFunc()
}
