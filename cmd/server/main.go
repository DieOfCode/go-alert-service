package main

import (
	"fmt"
	"net/http"

	"github.com/DieOfCode/go-alert-service/internal/storage"
)

func main() {
	memStorage := storage.NewMemStorage()

	http.HandleFunc("/update/", memStorage.HandleUpdateMetric)

	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Println("Ошибка запуска сервера:", err)
	}
}
