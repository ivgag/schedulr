package service

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type TimezoneService struct {
	config *GoogleConfig
}

func NewTimezoneService(config *GoogleConfig) *TimezoneService {
	return &TimezoneService{config: config}
}

// GetTimezone returns the IANA time zone identifier for the given location.
func (s *TimezoneService) GetTimezone(latitude float64, longitude float64) (*TimezoneDto, error) {
	url := "https://maps.googleapis.com/maps/api/timezone/json" +
		"?location=" + fmt.Sprintf("%f,%f", latitude, longitude) +
		"&timestamp=" + fmt.Sprintf("%d", time.Now().Unix()) +
		"&key=" + s.config.ApiKey

	response, err := http.DefaultClient.Get(
		url,
	)

	if err != nil {
		return nil, err
	}

	defer response.Body.Close()

	var timezone TimezoneDto
	err = json.NewDecoder(response.Body).Decode(&timezone)
	if err != nil {
		return nil, err
	}

	return &timezone, nil
}

type TimezoneDto struct {
	TimezoneId   string `json:"timeZoneId"`
	TimezoneName string `json:"timeZoneName"`
	RawOffset    int    `json:"rawOffset"`
	DstOffset    int    `json:"dstOffset"`
	Status       string `json:"status"`
}
