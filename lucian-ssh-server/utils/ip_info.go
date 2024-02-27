package utils

import (
	"fmt"
	"io"
	"net/http"
)

// TODO godoc

// https://docs.freeipapi.com/
func GetIpInfo(ipAddress string) (*string, error) {
	url := fmt.Sprintf("https://freeipapi.com/api/json/%s", ipAddress)

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
