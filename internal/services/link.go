package services

import (
	"fmt"
	"test/internal/app"
	"test/internal/dto"
	"test/internal/models"
	"test/internal/repository"

	"github.com/google/uuid"
)

type LinkService struct {
	LinkRepo *repository.LinkRepostiory
	App      *app.App
}

func NewLinkService(lRepo *repository.LinkRepostiory, app *app.App) *LinkService {
	return &LinkService{
		LinkRepo: lRepo,
		App:      app,
	}
}

// Создаём линки и сохраняем их в памяти
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
		fmt.Println(link_obj)
		links = append(links, link_obj)
	}

	linkList.LinksData = links

	ls.LinkRepo.SaveLinks(linkList)

	return linkList
}
