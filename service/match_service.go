package service

import (
	"context"
	"log"

	"pingpong/domain"
	"pingpong/ports"
)

type matchService struct {
	repo ports.MatchRepository
}

func NewMatchService(repo ports.MatchRepository) ports.MatchService {
	return &matchService{
		repo: repo,
	}
}

func (s *matchService) SaveMatch(ctx context.Context, match domain.Match) error {
	log.Printf("Service: Saving match #%d", match.MatchNumber)
	return s.repo.SaveMatch(ctx, match)
}

func (s *matchService) GetMatchByID(ctx context.Context, id int) (domain.Match, error) {
	log.Printf("Service: Getting match with ID %d", id)
	return s.repo.GetMatchByID(ctx, id)
}

func (s *matchService) GetLastMatch(ctx context.Context) (domain.Match, error) {
	log.Println("Service: Getting last match")
	return s.repo.GetLastMatch(ctx)
}

func (s *matchService) TestConnection(ctx context.Context) error {
	log.Println("Service: Testing database connection")
	return s.repo.TestConnection(ctx)
}