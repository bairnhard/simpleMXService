package main

import (
	"bytes"
	"fmt"
	"net"

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

	// TXT Lookup
	ipresult, err := net.LookupTXT(hoststr)
	if err != nil {
		hostname.JSON(404, "no such host")
		//panic(err)
	}

	//IP V4 Lookup
	ipv4, err := net.LookupIP(hoststr)
	if err != nil {
		hostname.JSON(404, "no such host")
		//panic(err)
	}

	hostname.JSON(200, ipv4)

	hostname.JSON(200, ipresult)

}

//------------------------------------------------------------------------------------------

func getMXResults(domain *gin.Context) {

	domainstr := domain.Params.ByName("domain")

	//hier kommt die komplette mx lookup arie
	MXResult, err := net.LookupMX(domainstr)
	if err != nil {
		panic(err)
	}

	MXTemp := strmx(MXResult)

	domain.JSON(200, MXTemp)

}

//------------------------------------------------------------------------------------------

func main() {

	router := gin.Default()

	// Usage:
	// http://localhost:8080/getDomain/google.com
	// http://localhost:8080/getProvider/aspmx.l.google.com

	router.GET("/getDomain/:domain", getMXResults)
	router.GET("/getProvider/:hostname", getProvider)

	router.Run()

}
