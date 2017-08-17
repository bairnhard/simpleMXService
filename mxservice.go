package main

import (
	"bytes"
	"fmt"
	"net"
	"os"

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

func getMXResults(domain *gin.Context) {

	domainstr := domain.Params.ByName("domain")

	//hier kommt die komplette mx lookup arie
	MXResult, err := net.LookupMX(domainstr)
	if err != nil {
		panic(err)
	}

	MXTemp := strmx(MXResult)

	//hier geben wir was zurück
	domain.JSON(200, MXTemp)

}

//------------------------------------------------------------------------------------------

func main() {

	router := gin.Default()

	//get domain
	router.GET("/:domain", getMXResults)

	//PortToUse := ":" + os.Args[1]
	err := os.Setenv("PORT", os.Args[1])
	if err != nil {
		panic(err)
	}
	router.Run(:9090)

}
