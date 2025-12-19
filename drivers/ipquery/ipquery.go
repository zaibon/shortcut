package ipquery

import (
	"encoding/json"
	"fmt"
	"net/http"
)

type Response struct {
	IP  string `json:"ip"`
	Isp struct {
		Asn string `json:"asn"`
		Org string `json:"org"`
		Isp string `json:"isp"`
	} `json:"isp"`
	Location struct {
		Country     string  `json:"country"`
		CountryCode string  `json:"country_code"`
		City        string  `json:"city"`
		State       string  `json:"state"`
		Zipcode     string  `json:"zipcode"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		Timezone    string  `json:"timezone"`
		Localtime   string  `json:"localtime"`
	} `json:"location"`
	Risk struct {
		IsMobile     bool `json:"is_mobile"`
		IsVpn        bool `json:"is_vpn"`
		IsTor        bool `json:"is_tor"`
		IsProxy      bool `json:"is_proxy"`
		IsDatacenter bool `json:"is_datacenter"`
		RiskScore    int  `json:"risk_score"`
	} `json:"risk"`
}

const baseURL = "https://api.ipquery.io"

func QueryIP(ip string) (*Response, error) {
	url := fmt.Sprintf("%s/%s?format=json", baseURL, ip)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}

	defer func() {
		_ = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("failed to query IP: %s", resp.Status)
	}

	r := &Response{}
	if err := json.NewDecoder(resp.Body).Decode(r); err != nil {
		return nil, err
	}

	return r, nil
}
