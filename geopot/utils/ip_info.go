package utils

import (
	"fmt"
	"io"
	"math/rand/v2"
	"net"
	"net/http"
)

// Get details about IP address from https://docs.freeipapi.com/
//
// If ipAddress is nil, it will get info about the caller's IP address
//
//	@param ipAddress
//	@return *string
//	@return error
func GetIpInfo(ipAddress *string) (*string, error) {
	url := ""
	if ipAddress == nil {
		url = "https://freeipapi.com/api/json"
	} else {
		url = fmt.Sprintf("https://freeipapi.com/api/json/%s", *ipAddress)
	}

	response, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != 200 {
		err = fmt.Errorf("status code %s", response.Status)
		return nil, err
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}

	stringResult := string(body)
	return &stringResult, nil
}

// Check if IP address is public
//
//	@param ipAddress
//	@return bool
func IsPublicIP(ipAddress string) bool {
	ip := net.ParseIP(ipAddress)
	return ip != nil && !ip.IsPrivate() && !ip.IsLoopback() && !ip.IsUnspecified()
}

// Generating random coordinates for testing purposes
func RandomCoordinate(min, max float64) float64 {
	return min + (max-min)*(min+rand.Float64()*(max-min))
}
