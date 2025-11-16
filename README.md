
## Установка и тестирование

Клонируем репозиторий 
```
git clone https://github.com/solluzumo/131125.git
cd 131125/
```

Устанавливаем зависимости
```
go mod download
```

Запускаем проект
```
go run cmd/main.go
```

## Эндпоинты для тестирования

Проверка работоспособности URL
```
POST http://localhost:8080/links
{
    "links":["https://habr.com","https://google.com"]
}
```

Получение pdf
```
POST http://localhost:8080/get-pdf
{
    "links_list":[
      "id1",
      "id2",
      "id3"
    ]
}

```
