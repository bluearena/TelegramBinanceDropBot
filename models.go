package main

import "time"

type Config struct {
	BotToken  string
	YourID	  int64
	Period 	  time.Duration
	Percentage float64
}
