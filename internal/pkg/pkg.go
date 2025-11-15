package pkg

import (
	"crypto/sha1"
	"fmt"
	"net/http"
	"sort"
	"strings"
	"time"
)

// Функция для генерации имени из id наборов линков
// Работает на основе хэшеирования sha1
func PDFNameFromIDs(ids []string) string {
	sort.Strings(ids)

	joined := strings.Join(ids, ",")
	hash := sha1.Sum([]byte(joined))

	return fmt.Sprintf("%x.pdf", hash[:8])
}

// Функция отправки запроса для проверки доступности конкретного URL
func SendRequest(URL string) string {

	client := &http.Client{
		Timeout: 2 * time.Second,
	}
	resp, err := client.Head(URL)

	if (err != nil) || (resp.Status != "200 OK") {

		fmt.Printf("SEND REQUEST ERROR:%v; STATUS:%s", err, resp.Status)

		return "not available"
	}

	return "available"
}
