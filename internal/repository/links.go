package repository

import (
	"fmt"
	"path/filepath"
	"test/internal/app"
	"test/internal/models"
)

type LinkRepostiory struct {
	App *app.App
}

func NewLinkRepostiory(app *app.App) *LinkRepostiory {
	return &LinkRepostiory{
		App: app,
	}
}

func (lr *LinkRepostiory) SaveLinks(data *models.LinkList) bool {
	var data_formed models.LinkJson
	path := filepath.Join(lr.App.Config.DataDir, "links.json")

	existingData, err := ReadJson[models.LinkJson](path)
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
		return false
	}

	//Сериализуем данные под Json
	data_formed.ID = data.ID
	data_formed.LinksData = make(map[string]string)
	for _, el := range data.LinksData {
		fmt.Println("fasdf    ", el)
		obj := *el
		data_formed.LinksData[obj.URL] = obj.Status
	}

	existingData = append(existingData, data_formed)

	err = WriteJSON(path, existingData)
	if err != nil {
		fmt.Printf("Ошибка при записи файла:%v\n", err)
		return false
	}

	return true
}
