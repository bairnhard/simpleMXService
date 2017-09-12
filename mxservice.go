package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	geoip2 "github.com/oschwald/geoip2-golang"
)

func strmx(mxs []*net.MX) string {
	var buf bytes.Buffer
	sep := ""

	for _, mx := range mxs {
		fmt.Fprintf(&buf, "%s%s:%d", sep, mx.Host, mx.Pref)
		sep = ";"
	}

	return buf.String()
}

func getProvider(hostname *gin.Context) {

	// Work in progress - currently just returns an IP Address

	hoststr := hostname.Params.ByName("hostname")

	// IP V4 Lookup
	ipv4, err := net.LookupIP(hoststr)
	if err != nil {
		hostname.JSON(404, "no such host")
		//panic(err)
		log.Fatal(err)
	}

	hostname.IndentedJSON(200, ipv4)

}

//------------------------------------------------------------------------------------------

func getMXResults(domain *gin.Context) {

	domainstr := domain.Params.ByName("domain")

	//get the MX records
	MXResult, err := net.LookupMX(domainstr)
	if err != nil {
		log.Fatal(err)
		//panic(err)
	}

	domain.IndentedJSON(200, MXResult)

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

	type GeoipShort struct {
		CountryCode string  `json:"country_code"`
		Latitude    float64 `json:"latitude"`
		Longitude   float64 `json:"longitude"`
	}

	url := fmt.Sprintf("http://freegeoip.net/JSON/%s", hoststr)

	resp, err := http.Get(url)
	if err != nil {
		// panic(err)
		log.Fatal(err)

	}

	defer resp.Body.Close()

	// var record GeoipResult
	var record GeoipShort

	// Use json.Decode for reading streams of JSON data
	if err := json.NewDecoder(resp.Body).Decode(&record); err != nil {
		//panic(err)
		log.Fatal(err)
	}

	hostname.IndentedJSON(200, record)

}

func getlocalIP(ip *gin.Context) { //GeoIP

	ipg := ip.Params.ByName("ip")
	ip2 := net.ParseIP(ipg)

	//MaxMind GeoIP City Lite DB
	// download from https://dev.maxmind.com/geoip/geoip2/geolite2/

	db, err := geoip2.Open("GeoLite2-City.mmdb")
	if err != nil {
		log.Fatalln(" Database error: ", err)
		//panic(err)
	}

	record, err := db.City(ip2)
	if err != nil {
		log.Fatal(err)
	}

	ip.IndentedJSON(200, record)

	defer db.Close()

}
func downloadGeoDB(url string) string {
	tokens := strings.Split(url, "/")
	fileName := tokens[len(tokens)-1]
	fmt.Println("Downloading", url, "to", fileName)

	// TODO: check file existence first with io.IsExist

	if _, err := os.Stat(fileName); !os.IsNotExist(err) {
		// path/to/whatever exists
		os.Remove(fileName)
	}

	output, err := os.Create(fileName)

	if err != nil {
		log.Fatal("Error while creating", fileName, "-", err)
		//fmt.Println("Error while creating", fileName, "-", err)
		return "error"
	}
	defer output.Close()

	response, err := http.Get(url)
	if err != nil {
		//fmt.Println("Error while downloading", url, "-", err)
		log.Fatalln("Error while downloading", url, "-", err)
		return "error"
	}
	defer response.Body.Close()

	n, err := io.Copy(output, response.Body)
	if err != nil {
		log.Fatalln("Error while downloading", url, "-", err)
		return "error"
	}

	log.Println(n, "bytes downloaded.")

	return fileName

}

func processFile(srcFile string) {

	f, err := os.Open(srcFile)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	defer f.Close()

	gzf, err := gzip.NewReader(f)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	tarReader := tar.NewReader(gzf)
	// defer io.Copy(os.Stdout, tarReader)

	for true {
		header, err := tarReader.Next()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println(err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir: // = directory
			fmt.Println("Directory:", name)
		//	os.Mkdir(name, 0755)
		case tar.TypeReg: // = regular file
			tokens := strings.Split(name, "/")
			fileName := tokens[len(tokens)-1]
			fmt.Println("Regular file:", fileName)
			data := make([]byte, header.Size)
			_, err := tarReader.Read(data)
			if err != nil {
				panic("Error reading file!!! PANIC!!!!!!")
			}

			ioutil.WriteFile(fileName, data, 0755)
		default:
			fmt.Printf("%s : %c %s %s\n",
				"Yikes! Unable to figure out type",
				header.Typeflag,
				"in file",
				name,
			)
		}
	}
}

func md5verify(dbname string, md5name string) bool {

	hashvalue, err := ioutil.ReadFile(md5name) // md5 hash
	if err != nil {
		log.Fatalln("error reading hash file: ", err)
	}

	fmt.Println(hashvalue) // print the content as 'bytes'

	hashstr := string(hashvalue) // convert content to a 'string'
	fmt.Println(hashstr)

	file, err := os.Open(dbname)

	if err != nil {
		log.Fatalln("error reading db file: ", err)
	}

	defer file.Close()

	hash := md5.New()
	_, err = io.Copy(hash, file)

	if err != nil {
		log.Fatalln("error generating MD5 Hash: ", err)
	}

	hr := hash.Sum(nil)

	n := -1
	for i, b := range hr {
		if b == 0 {
			break
		}
		n = i
	}
	hresultstr := string(hr[:n+1])

	fmt.Printf("%s MD5 checksum is %x \n", file.Name(), hash.Sum(nil))

	fmt.Println("hresultstr: ", hresultstr)
	fmt.Println("hashfile: ", hashstr)

	if hresultstr == hashstr {
		fmt.Println("Hashes match")
		return true
	}
	fmt.Println("Hash mismatch")
	return false

}

//------------------------------------------------------------------------------------------

func main() {

	router := gin.Default()

	// Usage:
	// http://localhost:8080/getDomain/google.com
	// http://localhost:8080/getProvider/aspmx.l.google.com
	// http://localhost:8080/getGeoIp/aspmx.l.google.com    <- uses freegeoip.net - max 15k requsts per hour
	// http://localhost:8080/getlocalIp/123.123.123.123 <- uses MaxMind GeoIP databases locally - need regular update at least daily

	//TODO:
	// Implement config file
	// Implement logging
	// Implement DB Updates
	// Implement provider table lookup

	fmt.Println("Initializing download...")
	dbName := downloadGeoDB("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz")

	// unpack downloaded file
	fmt.Println("unpacking...", dbName)
	processFile(dbName)

	fmt.Println("getting MD5 HASH...", dbName)
	mdName := downloadGeoDB("http://geolite.maxmind.com/download/geoip/database/GeoLite2-City.tar.gz.md5")

	md5verify(dbName, mdName)

	router.GET("/getDomain/:domain", getMXResults)
	router.GET("/getProvider/:hostname", getProvider)
	router.GET("/getGeoIp/:hostname", geoIP)
	router.GET("/getlocalIp/:ip", getlocalIP)

	router.Run()

}
