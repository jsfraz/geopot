package models

import (
	"encoding/json"
	"time"
)

// SSH connection
type Connection struct {
	ID uint64 `json:"id" gorm:"primaryKey;uniqueIndex:idx_id_timestamp,priority:1" validate:"required"`

	IPVersion     int     `json:"ipVersion" validate:"required"`
	IPAddress     string  `json:"ipAddress" validate:"required"`
	Latitude      float64 `json:"latitude" validate:"required"`
	Longitude     float64 `json:"longitude" validate:"required"`
	CountryName   string  `json:"countryName" validate:"required"`
	CountryCode   string  `json:"countryCode" validate:"required"`
	TimeZone      string  `json:"timeZone" validate:"required"`
	ZipCode       string  `json:"zipCode" validate:"required"`
	CityName      string  `json:"cityName" validate:"required"`
	RegionName    string  `json:"regionName" validate:"required"`
	IsProxy       bool    `json:"isProxy" validate:"required"`
	Continent     string  `json:"continent" validate:"required"`
	ContinentCode string  `json:"continentCode" validate:"required"`

	User      string    `json:"user" validate:"required"`
	Password  string    `json:"password" validate:"required"`
	Timestamp time.Time `json:"timestamp" gorm:"type:timestamptz;index:idx_timestamp;uniqueIndex:idx_id_timestamp,priority:2;not null" validate:"required"`

	// Number of failed geolocation attempts. Only used while the record waits
	// in Valkey, it is never persisted nor sent to WebSocket clients.
	Attempts int `json:"attempts,omitempty" gorm:"-"`
}

// geoDetails mirrors the subset of the https://docs.freeipapi.com/ response
// consumed by SetConnectionDetails. It is kept separate from Connection
// because the upstream API reports time zones as an array and may send a null
// zip code, which Connection (a database model) must not encode.
type geoDetails struct {
	IPVersion     int      `json:"ipVersion"`
	Latitude      float64  `json:"latitude"`
	Longitude     float64  `json:"longitude"`
	CountryName   string   `json:"countryName"`
	CountryCode   string   `json:"countryCode"`
	TimeZone      string   `json:"timeZone"`
	TimeZones     []string `json:"timeZones"`
	ZipCode       *string  `json:"zipCode"`
	CityName      string   `json:"cityName"`
	RegionName    string   `json:"regionName"`
	IsProxy       bool     `json:"isProxy"`
	Continent     string   `json:"continent"`
	ContinentCode string   `json:"continentCode"`
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
	var details geoDetails
	err := json.Unmarshal([]byte(jsonData), &details)
	if err != nil {
		return err
	}

	// The API returns "timeZones" (array); older responses used "timeZone".
	timeZone := details.TimeZone
	if timeZone == "" && len(details.TimeZones) > 0 {
		timeZone = details.TimeZones[0]
	}
	// "zipCode" is null for many IPs, so it is decoded as a pointer.
	zipCode := ""
	if details.ZipCode != nil {
		zipCode = *details.ZipCode
	}

	c.IPVersion = details.IPVersion
	c.Latitude = details.Latitude
	c.Longitude = details.Longitude
	c.CountryName = details.CountryName
	c.CountryCode = details.CountryCode
	c.TimeZone = timeZone
	c.ZipCode = zipCode
	c.CityName = details.CityName
	c.RegionName = details.RegionName
	c.IsProxy = details.IsProxy
	c.Continent = details.Continent
	c.ContinentCode = details.ContinentCode
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

// Marshal Connection to binary (JSON).
//
//	@return []byte
//	@return error
func (c Connection) MarshalBinary() ([]byte, error) {
	return json.Marshal(c)
}
