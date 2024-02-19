package main

import (
	"github.com/DieOfCode/go-alert-service/internal/application"
	_ "github.com/jackc/pgx/v5/stdlib"
)

func main() {
	application.Run()
}
