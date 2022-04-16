package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"time"
)

const homebrewUrlFormat = "https://formulae.brew.sh/api/formula/%s.json"

var (
	owner                         = os.Getenv("OWNER")
	database                      = os.Getenv("DATABASE")
	sourceBranch                  = os.Getenv("SOURCE_BRANCH")
	homebrewFomula                = os.Getenv("HOMEBREW_FORMULA")
	dolthubAuthTokenParameterName = os.Getenv("AUTH_TOKEN_PARAMETER")
)

func main() {
	homebrewUrl := fmt.Sprintf(homebrewUrlFormat, homebrewFomula)
	response, statusCode := get(homebrewUrl, nil)
	logJsonResponseBody(response, statusCode)

	// pull out the 30d installs
	var document map[string]interface{}
	json.Unmarshal(response, &document)
	installs := unmarshall30dInstalls(document)

	// Update on a branch on DoltHub
	RunQueryOnNewBranch(owner, database, sourceBranch, "homebrew/publish",
		fmt.Sprintf("insert into homebrew_metrics values(NOW(), %d);", installs))

	// Merge DoltHub Change
	pause() // TODO: Switch to polling
	Merge(owner, database, "homebrew/publish", sourceBranch)
}

func unmarshall30dInstalls(result map[string]interface{}) int {
	// Data we're processing looks like:
	//	"analytics": {
	//		"install": {
	//		  "30d": {
	//			"dolt": 420
	analytics := result["analytics"].(map[string]interface{})
	install := analytics["install"].(map[string]interface{})
	thirtyDays := install["30d"].(map[string]interface{})
	installsIn30days := int(thirtyDays[homebrewFomula].(float64))

	fmt.Println("Total Homebrew Installs in Past 30 Days: " + strconv.Itoa(installsIn30days))

	return installsIn30days
}

func logJsonResponseBody(response []byte, statusCode int) {
	if statusCode == http.StatusOK {
		jsonPrettyPrint(response)
	} else {
		fmt.Println(string(response))
		panic("Error executing request")
	}
}

func pause() {
	duration, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	time.Sleep(duration)
}

func jsonPrettyPrint(response []byte) {
	var out bytes.Buffer
	json.Indent(&out, response, "", "   ")
	out.WriteTo(os.Stdout)
}
