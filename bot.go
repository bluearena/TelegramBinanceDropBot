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
	exclude = make(map[string]struct{})

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

	checkedTime := time.Now()
	for{
		time.Sleep(10*time.Second)
		update := false
		if time.Now().Sub(checkedTime).Minutes() >= configuration.Period{
			update = true
			checkedTime = time.Now()
			for k := range exclude {
				delete(exclude, k)
			}
			log.Print("Updated all")
		}
		startObserving(update)
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

func startObserving(update bool){
	log.Println("Starting checking")

	var err error
	prices, err := client.NewListPricesService().Do(context.Background())
	if err != nil {
		log.Println(err)
	}

	if update{
		prevPrices = prices
		log.Print("Updated prices")
	}

	for i, p := range prices {
		if _, ok := exclude[p.Symbol]; ok{
			log.Print(p.Symbol + " is exluded. Continue.")
			continue
		}
		//log.Print(p.Symbol + " isn`t excluded.")
		currPrice, _ := strconv.ParseFloat(p.Price, 64)
		prevPrice, _ := strconv.ParseFloat(prevPrices[i].Price, 64)
		perc := currPrice * float64(100) / prevPrice
		if 100-perc >= configuration.Percentage{
			var name,base string
			switch p.Symbol[len(p.Symbol)-1] {
			case 'C':
				name = p.Symbol[:len(p.Symbol)-3]
				base = p.Symbol[len(name):]
			case 'H':
				name = p.Symbol[:len(p.Symbol)-3]
				base = p.Symbol[len(name):]
			case 'B':
				continue
			case 'T':
				name = p.Symbol[:len(p.Symbol)-4]
				base = p.Symbol[len(name):]
			}
			
			
			data, _ := client.NewPriceChangeStatsService().Symbol(p.Symbol).Do(context.Background())
			volumeFloat, _ := strconv.ParseFloat(data.Volume, 64)
			volume := Spacef(volumeFloat*currPrice)
			msg := name + "/" + base + " fell by " +
				strconv.FormatFloat(100-perc, 'f', 2, 64) + "% in the last " +
				strconv.FormatFloat(configuration.Period, 'f', 0, 64) +
				" minutes.\nCurrently @ " +
				strconv.FormatFloat(currPrice, 'f', -1, 64) + " - " +
				time.Now().Format("15:04:05") + "\n24hr volume on Binance: " +
				volume + " " + base
			bot.Send(tgbotapi.NewMessage(configuration.YourID, msg))
			//bot.Send(tgbotapi.NewMessage(393525533, msg))
			log.Print(msg)
			exclude[p.Symbol] = struct{}{}
		}
	}

}