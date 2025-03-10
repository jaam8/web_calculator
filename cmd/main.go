package main

import (
	"github.com/jaam8/web_calculator/internal/agent"
	"github.com/jaam8/web_calculator/internal/api"
	"github.com/jaam8/web_calculator/internal/logger"
)

func main() {
	logger.InitLogger()
	go api.RunServer()
	agent.Run()
}
