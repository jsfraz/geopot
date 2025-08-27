package models

import (
	"encoding/json"
	"time"
)

// SSH connection
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

// Create new SSH connection info.
//
//	@param host
//	@param user
//	@param password
//	@param timestamp
//	@return *Connection
func NewConnection(host string, user string, password string, timestamp time.Time) *Connection {
	var connection Connection
	connection.IPAddress = host
	connection.User = user
	connection.Password = password
	connection.Timestamp = timestamp
	return &connection
}

// Set details about SSH connection.
//
//	@receiver c
//	@param jsonData
//	@return error
func (c *Connection) SetConnectionDetails(jsonData string) error {
	var connection Connection
	err := json.Unmarshal([]byte(jsonData), &connection)
	if err != nil {
		return err
	}
	c.IPVersion = connection.IPVersion
	c.Latitude = connection.Latitude
	c.Longitude = connection.Longitude
	c.CountryName = connection.CountryName
	c.CountryCode = connection.CountryCode
	c.TimeZone = connection.TimeZone
	c.ZipCode = connection.ZipCode
	c.CityName = connection.CityName
	c.RegionName = connection.RegionName
	c.IsProxy = connection.IsProxy
	c.Continent = connection.Continent
	c.ContinentCode = connection.ContinentCode
	return nil
}

// Create Connection instance from JSON.
//
//	@param jsonData
//	@return *Connection
//	@return error
func ConnectionFromJson(jsonData string) (*Connection, error) {
	var c Connection
	err := json.Unmarshal([]byte(jsonData), &c)
	if err != nil {
		return nil, err
	}
	return &c, nil
}

func (c Connection) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
