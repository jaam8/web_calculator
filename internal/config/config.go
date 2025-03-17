package config

import (
	"github.com/joho/godotenv"
	"log"
	"os"
	"strconv"
	"time"
)

var Configs = NewConfig()

type Config struct {
	Port                string
	TimeAddition        int
	TimeSubtraction     int
	TimeMultiplications int
	TimeDivisions       int
	ComputingPower      int
	RequestURL          string
	WaitTime            int
}

func NewConfig() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	ta, _ := strconv.Atoi(os.Getenv("TIME_ADDITION_MS"))
	if ta == 0 {
		ta = 100
	}
	ts, _ := strconv.Atoi(os.Getenv("TIME_SUBTRACTION_MS"))
	if ts == 0 {
		ts = 100
	}
	tm, _ := strconv.Atoi(os.Getenv("TIME_MULTIPLICATIONS_MS"))
	if tm == 0 {
		tm = 100
	}
	td, _ := strconv.Atoi(os.Getenv("TIME_DIVISIONS_MS"))
	if td == 0 {
		td = 100
	}
	cp, _ := strconv.Atoi(os.Getenv("COMPUTING_POWER"))
	if cp == 0 {
		cp = 2
	}
	requestURL := os.Getenv("REQUEST_URL")
	if requestURL == "" {
		requestURL = "http://localhost:8080"
	}
	waitTime, _ := strconv.Atoi(os.Getenv("WAIT_TIME_MS"))
	if waitTime == 0 {
		waitTime = 100
	}
	log.Println("Successfully loaded .env file")
	return &Config{
		Port:                port,
		TimeAddition:        ta,
		TimeSubtraction:     ts,
		TimeMultiplications: tm,
		TimeDivisions:       td,
		ComputingPower:      cp,
		RequestURL:          requestURL,
		WaitTime:            waitTime,
	}
}

func (c Config) GetOperationsTime(oper string) time.Duration {
	switch oper {
	case "+":
		return time.Duration(c.TimeAddition) * time.Millisecond
	case "-":
		return time.Duration(c.TimeSubtraction) * time.Millisecond
	case "*":
		return time.Duration(c.TimeMultiplications) * time.Millisecond
	case "/":
		return time.Duration(c.TimeDivisions) * time.Millisecond
	default:
		return 0
	}
}
