package ports

import (
	"context"

	"pingpong/domain"
)

type MatchService interface {
	SaveMatch(ctx context.Context, match domain.Match) error
	GetMatchByID(ctx context.Context, id int) (domain.Match, error)
	GetLastMatch(ctx context.Context) (domain.Match, error)
	TestConnection(ctx context.Context) error
}