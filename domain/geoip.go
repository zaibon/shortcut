package domain

type IPLocation struct {
	IP          string
	Country     string
	CountryCode string
	City        string
	State       string
	Zipcode     string
	Latitude    float64
	Longitude   float64
	Timezone    string
	Localtime   string
}
