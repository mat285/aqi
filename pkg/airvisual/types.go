package airvisual

import "time"

// Status is the status of a request
type Status string

const (
	// StatusSuccess is the sucess status
	StatusSuccess Status = "success"
	// StatusFailed is the failed status
	StatusFailed Status = "failed"
)

const (
	BaseURL           = "https://api.airvisual.com/v2/"
	CityURL           = BaseURL + "city"
	LocationURLFormat = CityURL + "?city=%s&state=%s&country=%s"
)

// LocationRequest is a request for a location data
type LocationRequest struct {
	City    string
	State   string
	Country string
}

// Response is a response from air visual
type Response struct {
	Status Status `json:"status"`
	Data   Data   `json:"data"`
}

// Data is the payload of a response
type Data struct {
	City    string `json:"city"`
	State   string `json:"state"`
	Country string `json:"country"`
	Current Air    `json:"current"`
}

// Air is the data about the air
type Air struct {
	Pollution Pollution `json:"pollution"`
	Weather   Weather   `json:"weather"`
}

// Pollution is pollution data
type Pollution struct {
	Time time.Time `json:"ts"`
	AQI  int       `json:"aqius"`
}

// Weather is the weather data
type Weather struct {
	Time          time.Time `json:"ts"`
	Humidity      int       `json:"hu"`
	Temperature   int       `json:"tp"`
	WindDirection int       `json:"wd"`
	WindSpeed     float32   `json:"ws"`
}
