package models

import (
	"encoding/json"
	"time"
)

// TODO godoc

type Connection struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	IPVersion     int     `json:"ipVersion"`
	IPAddress     string  `json:"ipAddress"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	CountryName   string  `json:"countryName"`
	CountryCode   string  `json:"countryCode"`
	TimeZone      string  `json:"timeZone"`
	ZipCode       string  `json:"zipCode"`
	CityName      string  `json:"cityName"`
	RegionName    string  `json:"regionName"`
	IsProxy       bool    `json:"isProxy"`
	Continent     string  `json:"continent"`
	ContinentCode string  `json:"continentCode"`

	User      string    `json:"user"`
	Password  string    `json:"password"`
	Timestamp time.Time `json:"timestamp"`
}

func NewConnection(jsonData string, user string, password string, timestamp time.Time) (*Connection, error) {
	var connection Connection
	err := json.Unmarshal([]byte(jsonData), &connection)
	if err != nil {
		return nil, err
	}
	connection.User = user
	connection.Password = password
	connection.Timestamp = timestamp
	return &connection, nil
}
