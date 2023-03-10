package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/mymmrac/telego"
	th "github.com/mymmrac/telego/telegohandler"
	tu "github.com/mymmrac/telego/telegoutil"
)

var months = map[string]string{
	"01": "января",
	"02": "февраля",
	"03": "марта",
	"04": "апреля",
	"05": "мая",
	"06": "июня",
	"07": "июля",
	"08": "августа",
	"09": "сентября",
	"10": "октября",
	"11": "ноября",
	"12": "декабря",
}

func main() {
	bot, err := telego.NewBot(os.Getenv("RACETG_BOT"), telego.WithDefaultDebugLogger())
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	botUser, err := bot.GetMe()
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	log.Printf("Work with account %v\n", botUser)

	updates, _ := bot.UpdatesViaLongPolling(nil)

	bh, err := th.NewBotHandler(bot, updates)
	if err != nil {
		log.Panic(err)
		os.Exit(1)
	}

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {

		var messageToUser string

		resp, err := http.Get("http://ergast.com/api/f1/2023.json")
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var races Object
		json.Unmarshal([]byte(body), &races)

		messageToUser = fmt.Sprintf("Календарь F1 сезона 2023 :\n%s", racesToString(races.MRData.RaceTable.Races))
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(message.Chat.ID),
			messageToUser,
		))

	}, th.CommandEqual("showraces"))

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {

		var messageToUser string

		resp, err := http.Get("http://ergast.com/api/f1/2023.json")
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var races Object
		json.Unmarshal([]byte(body), &races)

		userTime := message.Date
		//fmt.Println(userTime)
		isAfter := checkCurrToLastTime(userTime, races.MRData.RaceTable.Races[len(races.MRData.RaceTable.Races)-1])

		if isAfter {
			messageToUser = "Сезон закончился!"
		} else {
			nextRace := findNextRace(userTime, races.MRData.RaceTable.Races)
			messageToUser = fmt.Sprintf("Cледующий гран-при :\n%s", raceToString(formatDateTime(nextRace)))
		}
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(message.Chat.ID),
			messageToUser,
		))

	}, th.CommandEqual("nextrace"))

	defer bh.Stop()
	defer bot.StopLongPolling()

	bh.Start()
}

func racesToString(races []Race) string {

	var countRaces int = len(races)
	racesList := make([]string, countRaces)

	for num, race := range races {
		races[num] = formatDateTime(race)
	}

	for _, race := range races {
		racesList = append(racesList, raceToString(race))
	}

	return strings.Join(racesList, "")
}

func raceToString(race Race) string {
	return fmt.Sprintf("Номер этапа: %s,\nНазвание этапа: %s,\nДата этапа: %s,\nВремя этапа: %s.\n\n",
		race.Round, race.RaceName, race.Date, race.Time)
}

func formatDateTime(race Race) Race {

	tzone, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalln(err)
	}

	tempDateTime, err := time.Parse("2006-01-02 15:04:05Z", fmt.Sprintf("%s %s", race.Date, race.Time))
	if err != nil {
		log.Fatalln(err)
	}

	race.Date = ruMonth(tempDateTime.Format("2006-01-02"))
	race.Time = tempDateTime.In(tzone).Format("15:04")

	return race
}

func ruMonth(date string) string {

	partsDate := strings.Split(date, "-")
	for key, value := range months {
		if key == partsDate[1] {
			partsDate[1] = value
		}
	}

	return strings.Join([]string{partsDate[2], partsDate[1], partsDate[0]}, " ")
}

func checkCurrToLastTime(messageDate int64, race Race) bool {
	lastRace, err := time.Parse("2006-01-02 15:04:05Z", fmt.Sprintf("%s %s", race.Date, race.Time))
	if err != nil {
		log.Fatalln(err)
	}

	if messageDate >= int64(lastRace.Unix()) {
		return true
	} else {
		return false
	}
}

func findNextRace(messageDate int64, races []Race) Race {

	userDate := time.Unix(messageDate, 0)
	var numRace int

	for num, race := range races {

		tempDateTime, err := time.Parse("2006-01-02 15:04:05Z", fmt.Sprintf("%s %s", race.Date, race.Time))
		if err != nil {
			log.Fatalln(err)
		}

		if tempDateTime.After(userDate) {
			numRace = num
			break
		}
	}

	return races[numRace]
}
