package main

import (
	"fmt"
	"net/http"

	"github.com/DieOfCode/go-alert-service/internal/handler"
	"github.com/DieOfCode/go-alert-service/internal/storage"
)

func main() {
	memStorage := storage.NewMemStorage()
	handler := handler.NewHandler(memStorage)

	http.HandleFunc("/update/", handler.HandleUpdateMetric)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
