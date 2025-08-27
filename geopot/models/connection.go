package models

import (
	"encoding/json"
	"time"
)

// SSH connection
type Connection struct {
	ID uint64 `json:"id" gorm:"primarykey"`

	IPVersion     int     `json:"ipVersion"`
	IPAddress     string  `json:"ipAddress" gorm:"index:idx_conn_ip;index:idx_ip_timestamp,priority:1;index:idx_ip_country,priority:1"`
	Latitude      float64 `json:"latitude"`
	Longitude     float64 `json:"longitude"`
	CountryName   string  `json:"countryName" gorm:"index:idx_conn_country;index:idx_ip_country,priority:2;index:idx_country_time,priority:1"`
	CountryCode   string  `json:"countryCode" gorm:"index:idx_conn_country_code"`
	TimeZone      string  `json:"timeZone"`
	ZipCode       string  `json:"zipCode"`
	CityName      string  `json:"cityName" gorm:"index:idx_conn_city"`
	RegionName    string  `json:"regionName" gorm:"index:idx_conn_region"`
	IsProxy       bool    `json:"isProxy" gorm:"index:idx_conn_proxy;index:idx_proxy_time,priority:1"`
	Continent     string  `json:"continent" gorm:"index:idx_conn_continent"`
	ContinentCode string  `json:"continentCode"`

	User      string    `json:"user" gorm:"index:idx_conn_user;index:idx_user_time,priority:1"`
	Password  string    `json:"password"`
	Timestamp time.Time `json:"timestamp" gorm:"index:idx_conn_time;type:timestamptz;index:idx_ip_timestamp,priority:2;index:idx_country_time,priority:2;index:idx_proxy_time,priority:2;index:idx_user_time,priority:2"`
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
