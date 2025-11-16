package repository

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sync"
	"test/internal/app"
	"test/internal/models"

	"github.com/jung-kurt/gofpdf"
)

type LinkRepostiory struct {
	mu       *sync.Mutex
	app      *app.App
	linkPath string
	pdfPath  string
}

func NewLinkRepostiory(app *app.App, mu *sync.Mutex, lPath string, pPath string) *LinkRepostiory {
	return &LinkRepostiory{
		app:      app,
		mu:       mu,
		linkPath: lPath,
		pdfPath:  pPath,
	}
}

// Функция получения линков по id набора
// Возвращает маппу {"link1":"available","link2":"not available"}
func (lr *LinkRepostiory) GetLinksByID(id string) (map[string]string, error) {
	lr.mu.Lock()
	defer lr.mu.Unlock()

	data, err := lr.ReadLinkJson()
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

// Функция сохранения линков
func (lr *LinkRepostiory) SaveLinks(data map[string]string, idSet string) bool {
	lr.mu.Lock()
	defer lr.mu.Unlock()
	var data_formed models.LinkJson

	existingData, err := lr.ReadLinkJson()
	if err != nil {
		fmt.Printf("Ошибка при чтении файла:%v\n", err)
		return false
	}

	//Сериализуем данные под Json
	data_formed.ID = idSet

	data_formed.LinksData = data

	existingData = append(existingData, data_formed)

	err = lr.WriteLinkJSON(&existingData)
	if err != nil {
		fmt.Printf("Ошибка при записи файла:%v\n", err)
		return false
	}

	return true
}

// Функция чтения линков
func (lr *LinkRepostiory) ReadLinkJson() ([]models.LinkJson, error) {

	filePath := filepath.Join(lr.linkPath, "links.json")

	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	byteValue, err := ioutil.ReadAll(file)
	if err != nil {
		return nil, err
	}

	if len(byteValue) == 0 {
		return []models.LinkJson{}, nil
	}

	var data []models.LinkJson
	if err := json.Unmarshal(byteValue, &data); err != nil {
		return nil, err
	}

	return data, nil
}

// Функция сохранения линков
func (lr *LinkRepostiory) WriteLinkJSON(data *[]models.LinkJson) error {

	filePath := filepath.Join(lr.linkPath, "links.json")

	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(data, "", "    ")
	if err != nil {
		return err
	}

	_, err = file.Write(jsonData)
	return err
}

// Функция сохранения PDF
func (lr *LinkRepostiory) SavePDF(links []string, fileName string) error {

	filePath := filepath.Join(lr.pdfPath, fileName)

	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.AddPage()
	pdf.SetFont("Arial", "", 16)

	for _, line := range links {
		pdf.Cell(0, 10, line)
		pdf.Ln(12) // перенос строки
	}

	err := pdf.OutputFileAndClose(filePath)
	if err != nil {
		return err
	}
	return nil
}
