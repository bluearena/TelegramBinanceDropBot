package main

import (
	"log"
	"gopkg.in/telegram-bot-api.v4"
	"os"
	"io"
	"io/ioutil"
	"bytes"
	"encoding/json"
	"github.com/adshao/go-binance"

	"golang.org/x/net/context"
	"time"
	"strconv"

)

var (
	bot           *tgbotapi.BotAPI
	configuration Config
	prevPrices []*binance.SymbolPrice
	client *binance.Client
)

func main() {
	initLog()
	initConfig()
	initBinance()

	var err error
	bot, err = tgbotapi.NewBotAPI(configuration.BotToken)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}

	bot.Debug = false
	log.Printf("Authorized on account %s", bot.Self.UserName)

	for{
		time.Sleep(4*time.Minute)
		startObserving()
	}

}


func initLog() {
	f, err := os.OpenFile("log.txt", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
	defer f.Close()
	mw := io.MultiWriter(os.Stdout, f)
	log.SetOutput(mw)
}

func initConfig() {
	file, err := os.Open("config.json")
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}
	defer file.Close()

	body, err := ioutil.ReadAll(file)
	log.Print("First 10 bytes from config.json")
	log.Print(body[:10])
	body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
	log.Print("First 10 bytes after trim")
	reader := bytes.NewReader(body)
	log.Print(body[:10])
	decoder := json.NewDecoder(reader)

	err = decoder.Decode(&configuration)
	if err != nil {
		log.Print("ERROR: ")
		log.Panic(err)
	}

}

func initBinance(){
	client = binance.NewClient("", "")
	var err error
	prevPrices, err = client.NewListPricesService().Do(context.Background())
	if err != nil {
		log.Println(err)
	}
}

func startObserving(){
	log.Println("Starting checking")
	var err error
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		log.Println(err)
	}
	for i, p := range prices {
		currPrice, _ := strconv.ParseFloat(p.Price, 64)
		prevPrice, _ := strconv.ParseFloat(prevPrices[i].Price, 64)
		perc := currPrice * float64(100) / prevPrice
		if 100-perc >= 10{
			bot.Send(tgbotapi.NewMessage(configuration.YourID, p.Symbol + " fell by " + strconv.FormatFloat(100-perc, 'f', 2, 64) + "%"))
		}
	}
	prevPrices = prices
}