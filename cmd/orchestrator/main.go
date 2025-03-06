package main

import (
	"github.com/jaam8/web_calculator/internal/api"
	"github.com/jaam8/web_calculator/internal/logger"
)

func main() {
	logger.InitLogger()
	api.RunServer()
}
