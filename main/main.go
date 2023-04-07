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
		isAfter := checkCurrToLastTime(userTime, races.MRData.RaceTable.Races[len(races.MRData.RaceTable.Races)-1])

		if isAfter {
			messageToUser = "Сезон закончился!"
		} else {
			nextRace := findNextRace(userTime, races.MRData.RaceTable.Races)
			messageToUser = fmt.Sprintf("Cледующий гран-при :\n%s", raceFullInfoToString(formatDateTime(nextRace)))
		}
		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(message.Chat.ID),
			messageToUser,
		))

	}, th.CommandEqual("nextrace"))

	bh.HandleMessage(func(bot *telego.Bot, message telego.Message) {

		var messageToUser string

		resp, err := http.Get("http://ergast.com/api/f1/2023/driverStandings.json")
		if err != nil {
			log.Fatalln(err)
		}
		defer resp.Body.Close()

		body, err := io.ReadAll(resp.Body)
		if err != nil {
			log.Fatalln(err)
		}

		var object Object
		json.Unmarshal([]byte(body), &object)
		driversTable := object.MRData.StandingsTable.StandingsLists[0].DriverStandings

		messageToUser = fmt.Sprintf("Личный зачёт F1, сезон 2023: \n%s", driversToString(driversTable))

		_, _ = bot.SendMessage(tu.Messagef(
			tu.ID(message.Chat.ID),
			messageToUser,
		))

	}, th.CommandEqual("driverstandings"))

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

func raceFullInfoToString(race Race) string {
	if len(race.Sprint.Date) > 0 {
		return fmt.Sprintf("Номер этапа: %s,\nНазвание этапа: %s,\nВремя гонки: %s,\n\nПервая практика: %s,\nВторая практика: %s, \nКвалификация: %s,\nСпринт: %s.\n\n",
			race.Round, race.RaceName, race.Date+" "+race.Time, race.FirstPractice.Date+" "+race.FirstPractice.Time, race.SecondPractice.Date+" "+race.SecondPractice.Time, race.Qualifying.Date+" "+race.Qualifying.Time, race.Sprint.Date+" "+race.Sprint.Time)
	} else {
		return fmt.Sprintf("Номер этапа: %s,\nНазвание этапа: %s,\nВремя гонки: %s,\n\nПервая практика: %s,\nВторая практика: %s, \nТретья практика: %s,\nКвалификация: %s.\n",
			race.Round, race.RaceName, race.Date+" "+race.Time, race.FirstPractice.Date+" "+race.FirstPractice.Time, race.SecondPractice.Date+" "+race.SecondPractice.Time, race.ThirdPractice.Date+" "+race.ThirdPractice.Time, race.Qualifying.Date+" "+race.Qualifying.Time)
	}
}

func formatDateTime(race Race) Race {

	tzone, err := time.LoadLocation("Europe/Moscow")
	if err != nil {
		log.Fatalln(err)
	}

	raceDate := parseStringToTime(race.Date, race.Time)
	fPracticeDate := parseStringToTime(race.FirstPractice.Date, race.FirstPractice.Time)
	sPracticeDate := parseStringToTime(race.SecondPractice.Date, race.SecondPractice.Time)
	qualDate := parseStringToTime(race.Qualifying.Date, race.Qualifying.Time)

	race.Date = ruMonth(raceDate.Format("2006-01-02"))
	race.Time = raceDate.In(tzone).Format("15:04")

	race.FirstPractice.Date = ruMonth(fPracticeDate.Format("2006-01-02"))
	race.FirstPractice.Time = fPracticeDate.In(tzone).Format("15:04")

	race.SecondPractice.Date = ruMonth(sPracticeDate.Format("2006-01-02"))
	race.SecondPractice.Time = sPracticeDate.In(tzone).Format("15:04")

	race.Qualifying.Date = ruMonth(qualDate.Format("2006-01-02"))
	race.Qualifying.Time = qualDate.In(tzone).Format("15:04")

	if len(race.Sprint.Date) > 0 {
		sprDate := parseStringToTime(race.Sprint.Date, race.Sprint.Time)
		race.Sprint.Date = ruMonth(sprDate.Format("2006-01-02"))
		race.Sprint.Time = sprDate.In(tzone).Format("15:04")
	} else {
		tPracticeDate := parseStringToTime(race.ThirdPractice.Date, race.ThirdPractice.Time)
		race.ThirdPractice.Date = ruMonth(tPracticeDate.Format("2006-01-02"))
		race.ThirdPractice.Time = tPracticeDate.In(tzone).Format("15:04")
	}

	return race
}

func parseStringToTime(dateRace string, timeRace string) time.Time {
	tempDateTime, err := time.Parse("2006-01-02 15:04:05Z", fmt.Sprintf("%s %s", dateRace, timeRace))
	if err != nil {
		log.Fatalln(err)
	}
	return tempDateTime
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

func driversToString(drivers []DriverStandingsItem) string {
	var countDrivers int = len(drivers)
	driversList := make([]string, countDrivers)

	for _, driver := range drivers {
		driversList = append(driversList, driverToString(driver))
	}

	return strings.Join(driversList, "")
}

func driverToString(driver DriverStandingsItem) string {
	//return fmt.Sprintf("%2s | %3s... | %s \n", driver.PositionText, driver.Driver.Code, driver.Points)

	data := fmt.Sprintln(`` + driver.PositionText + ` | ` + driver.Driver.Code + ` | ` + driver.Points + ``)
	return data
}
