package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
)

//------------------------------------------------------------------------------------------

func strmx(mxs []*net.MX) string {
	var buf bytes.Buffer
	sep := ""
	//fmt.Fprintf(&buf, "[")

	for _, mx := range mxs {
		fmt.Fprintf(&buf, "%s%s:%d", sep, mx.Host, mx.Pref)
		sep = ";"
	}
	//fmt.Fprintf(&buf, "]")
	return buf.String()
}

//------------------------------------------------------------------------------------------

func getProvider(hostname *gin.Context) {

	// Work in progress - currently just returns an IP Address

	hoststr := hostname.Params.ByName("hostname")

	// IP V4 Lookup
	ipv4, err := net.LookupIP(hoststr)
	if err != nil {
		hostname.JSON(404, "no such host")
		panic(err)
	}

	hostname.JSON(200, ipv4)

}

//------------------------------------------------------------------------------------------

func getMXResults(domain *gin.Context) {

	domainstr := domain.Params.ByName("domain")

	//get the MX records
	MXResult, err := net.LookupMX(domainstr)
	if err != nil {
		panic(err)
	}

	MXTemp := strmx(MXResult)

	domain.JSON(200, MXTemp)

}

//------------------------------------------------------------------------------------------

//GeoIP - here we will get the country of origin...
// using JSON to GO to convert json to go type struct...

func geoIP(hostname *gin.Context) { //GeoIP

	hoststr := hostname.Params.ByName("hostname")

	type GeoipResult struct {
		IP          string  `json:"ip"`
		CountryCode string  `json:"country_code"`
		CountryName string  `json:"country_name"`
		RegionCode  string  `json:"region_code"`
		RegionName  string  `json:"region_name"`
		City        string  `json:"city"`
		ZipCode     string  `json:"zip_code"`
		TimeZone    string  `json:"time_zone"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
		MetroCode   int     `json:"metro_code"`
	}

	url := fmt.Sprintf("http://freegeoip.net/JSON/%s", hoststr)

	// Build the request
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)

	}

	// For control over HTTP client headers,
	// redirect policy, and other settings,
	// create a Client
	// A Client is an HTTP client
	client := &http.Client{}

	// Send the request via a client
	// Do sends an HTTP request and
	// returns an HTTP response
	resp, err := client.Do(req)
	if err != nil {
		panic(err)

	}

	// Callers should close resp.Body
	// when done reading from it
	// Defer the closing of the body
	defer resp.Body.Close()

	// Fill the record with the data from the JSON
	var record GeoipResult

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		panic(err)
	}

	hostname.JSON(200, record)

}

//------------------------------------------------------------------------------------------

func main() {

	router := gin.Default()

	// Usage:
	// http://localhost:8080/getDomain/google.com
	// http://localhost:8080/getProvider/aspmx.l.google.com
	// http://localhost:8080/getGeoIp/aspmx.l.google.com

	router.GET("/getDomain/:domain", getMXResults)
	router.GET("/getProvider/:hostname", getProvider)
	router.GET("/getGeoIp/:hostname", geoIP)

	router.Run()

}
