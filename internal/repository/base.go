package repository

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"test/internal/dto"
	"test/internal/models"
)

func UpdateLink(data *dto.LinkListResponse) bool {
	return true
}

func GetLinksByID(path string, id string) (map[string]string, error) {
	data, err := ReadJson[models.LinkJson](path)
	if err != nil {
		return nil, err
	}
	for _, el := range data {
		if el.ID == id {
			return el.LinksData, nil
		}
	}
	return nil, nil
}

func ReadJson[T any](path string) ([]T, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	// если файл пустой — возвращаем пустой срез
	if len(byteValue) == 0 {
		return []T{}, nil
	}

	var data []T
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}

	return data, nil
}

func WriteJSON[T any](filename string, data []T) error {
	file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return err
	}
	defer file.Close()

	// Конвертируем структуру данных в JSON-строку
	// json.MarshalIndent используется для читаемого форматирования
	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	// Записываем JSON-строку в файл
	_, err = file.Write(jsonData)
	if err != nil {
		return err
	}

	return nil
}
