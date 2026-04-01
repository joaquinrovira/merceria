package spreadsheets

import (
	"context"
	"fmt"
	"log"
	v0 "merceria/internal/spreadsheets/v0"
	"merceria/internal/util/cache"
	"time"

	"golang.org/x/oauth2/jwt"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

type Service struct {
	ctx       context.Context
	srv       *sheets.Service
	operators *cache.Cache[string, Operator] // SpreadsheetId -> Operator
}

func New(ctx context.Context, jwt *jwt.Config) (*Service, error) {
	srv, err := sheets.NewService(ctx, option.WithHTTPClient(jwt.Client(ctx)))
	if err != nil {
		return nil, fmt.Errorf("initializing: %w", err)
	}

	cache := cache.New(
		cache.WithTTL[string, Operator](5*time.Minute),
		cache.WithCapacity[string, Operator](100),
	)
	go cache.Start()
	go func() {
		<-ctx.Done()
		cache.Stop()
	}()

	return &Service{
		ctx:       ctx,
		srv:       srv,
		operators: cache,
	}, nil
}

//	type Operator interface {
//		Insert(ctx context.Context, rows []model.Row) error
//	}
type Operator = *v0.Operator

func (s *Service) GetOperator(sid string) (*v0.Operator, error) {
	item := s.operators.Get(sid)
	if item != nil {
		return item.Value(), nil
	}

	ctx, cancel := context.WithCancel(s.ctx)
	operator, err := v0.New(ctx, s.srv.Spreadsheets, sid)
	if err != nil {
		cancel()
		return nil, fmt.Errorf("creating operator for spreadsheet '%s': %w", sid, err)
	}

	unsub := s.operators.OnEviction(func(ctx context.Context, er cache.EvictionReason, i *cache.Item[string, Operator]) {
		if i.Key() == sid {
			cancel()
			log.Printf("Spreadsheet %s operator evicted", sid)
		}
	})

	go func() {
		defer unsub()
		<-ctx.Done()
	}()

	s.operators.Set(sid, operator, cache.DefaultTTL)
	return operator, nil
}
