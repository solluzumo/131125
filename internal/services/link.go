package services

import (
	"test/internal/repository"
)

type LinkService struct {
	LinkRepo *repository.LinkRepostiory
}

func NewLinkService(lRepo *repository.LinkRepostiory) *LinkService {
	return &LinkService{
		LinkRepo: lRepo,
	}
}
