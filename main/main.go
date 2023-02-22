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

		messageToUser = fmt.Sprintf("Календарь F1 сезона 2023 :\n%s", scheduledRaces(races.MRData.RaceTable.Races))
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(message.Chat.ID),
			messageToUser,
		))

	}, th.CommandEqual("showraces"))

	defer bh.Stop()
	defer bot.StopLongPolling()

	bh.Start()
}

func scheduledRaces(races []Race) string {

	var countRaces int = len(races)
	racesList := make([]string, countRaces)
	formatDateTime(races)

	for _, race := range races {
		racesList = append(racesList, fmt.Sprintf("Номер этапа: %s,\n Название этапа: %s,\n Дата этапа: %s,\n Время этапа: %s.\n\n",
			race.Round, race.RaceName, race.Date, race.Time))
	}

	return strings.Join(racesList, "")
}

func raceToString(race Race) string {
	return fmt.Sprintf("Номер этапа: %s,\nНазвание этапа: %s,\nДата этапа: %s,\nВремя этапа: %s.\n\n",
		race.Round, race.RaceName, race.Date, race.Time)
}

func formatDateTime(races []Race) {

	tzone, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalln(err)
	}

	for num, race := range races {

		var tempDateTime time.Time

		tempDateTime, err = time.Parse("2006-01-02 15:04:05Z", fmt.Sprintf("%s %s", race.Date, race.Time))
		if err != nil {
			log.Fatalln(err)
		}

		race.Date = ruMonth(tempDateTime.Format("2006-01-02"))
		race.Time = tempDateTime.In(tzone).Format("15:04")
		races[num] = race
	}
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
