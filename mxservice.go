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

func getMXResults(domain *gin.Context) {

	domainstr := domain.Params.ByName("domain")

	//hier kommt die komplette mx lookup arie
	MXResult, err := net.LookupMX(domainstr)
	if err != nil {
		panic(err)
	}

	MXTemp := strmx(MXResult)

	//fmt.Println(MXTemp)

	//hier geben wir was zurück
	domain.JSON(200, MXTemp)

}

//------------------------------------------------------------------------------------------

func main() {

	//_______________________________________________________________________
	router := gin.Default()

	//get domain
	router.GET("/:domain", getMXResults)

	router.Run()

}
