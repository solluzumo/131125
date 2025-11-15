package services

import (
	"test/internal/dto"
	"test/internal/models"
	"test/internal/repository"

	"github.com/google/uuid"
)

type LinkService struct {
	LinkRepo *repository.LinkRepostiory
}

func NewLinkService(lRepo *repository.LinkRepostiory) *LinkService {
	return &LinkService{
		LinkRepo: lRepo,
	}
}

// Функция создания и сохранения линки
func (ls *LinkService) SaveLinkService(data dto.LinkListRequest) *models.LinkList {
	var links []*models.Link
	linkList := &models.LinkList{
		ID: uuid.New().String(),
	}

	for _, el := range data.Links {
		link_obj := &models.Link{
			URL:    el,
			Status: "not avaliable",
		}
		links = append(links, link_obj)
	}

	linkList.LinksData = links

	ls.LinkRepo.SaveLinks(linkList)

	return linkList
}
