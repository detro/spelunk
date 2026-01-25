package main

import (
	"github.com/detro/golang-template/internal/echo"
	"github.com/detro/golang-template/internal/logger"
)

var (
	log = logger.Default()
)

func main() {
	log.Info("Hello, World!")
	log.Info("Ping()", "result", echo.Ping())
}
