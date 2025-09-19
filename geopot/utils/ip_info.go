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

// Generates a random public IP address avoiding reserved IP ranges.
//
//	@return string
func RandomPublicIP() string {
	// Choose random valid first octet (avoiding reserved ranges)
	validFirstOctets := []int{1, 2, 3, 4, 5}
	/*
		validFirstOctets := []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 11, 12, 13, 14, 15, 16, 17, 18, 19, 20,
			21, 22, 23, 24, 25, 26, 27, 28, 29, 30, 31, 32, 33, 34, 35, 36, 37, 38, 39, 40, 41, 42,
			43, 44, 45, 46, 47, 48, 49, 50, 51, 52, 53, 54, 55, 56, 57, 58, 59, 60, 61, 62, 63, 64,
			65, 66, 67, 68, 69, 70, 71, 72, 73, 74, 75, 76, 77, 78, 79, 80, 81, 82, 83, 84, 85, 86,
			87, 88, 89, 90, 91, 92, 93, 94, 95, 96, 97, 98, 99, 100, 101, 102, 103, 104, 105, 106,
			107, 108, 109, 110, 111, 112, 113, 114, 115, 116, 117, 118, 119, 120, 121, 122, 123, 124,
			125, 126, 128, 129, 130, 131, 132, 133, 134, 135, 136, 137, 138, 139, 140, 141, 142, 143,
			144, 145, 146, 147, 148, 149, 150, 151, 152, 153, 154, 155, 156, 157, 158, 159, 160, 161,
			162, 163, 164, 165, 166, 167, 168, 169, 170, 171, 172, 173, 174, 175, 176, 177, 178, 179,
			180, 181, 182, 183, 184, 185, 186, 187, 188, 189, 190, 191, 192, 193, 194, 195, 196, 197,
			198, 199, 200, 201, 202, 203, 204, 205, 206, 207, 208, 209, 210, 211, 212, 213, 214, 215,
			216, 217, 218, 219, 220, 221, 222, 223}
	*/

	firstOctet := validFirstOctets[rand.IntN(len(validFirstOctets))]
	secondOctet := rand.IntN(256)
	thirdOctet := rand.IntN(256)
	fourthOctet := rand.IntN(256)
	/*
		secondOctet := rand.IntN(5)
		thirdOctet := rand.IntN(5)
		fourthOctet := rand.IntN(5)
	*/

	return fmt.Sprintf("%d.%d.%d.%d", firstOctet, secondOctet, thirdOctet, fourthOctet)
}

// Generating random coordinates for testing purposes
func RandomCoordinate(min, max float64) float64 {
	return min + (max-min)*(min+rand.Float64()*(max-min))
}
