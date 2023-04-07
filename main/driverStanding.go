package main

type DriverStandingsItem struct {
	Position     int
	PositionText string
	Points       string
	Wins         string
	Driver       Driver
	Constructors []Constructors
}

type Constructors struct {
	ConstructorId string
	Url           string
	Name          string
	Nationality   string
}

type StandingsListItem struct {
	Season          string
	Round           string
	DriverStandings []DriverStandingsItem
}

type StandingsTable struct {
	Season         string
	StandingsLists []StandingsListItem
}
