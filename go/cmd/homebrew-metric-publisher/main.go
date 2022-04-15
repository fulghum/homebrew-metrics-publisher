package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"io"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"time"
)

const dolthubWriteUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s?q=%s"
const dolthubMergeUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s"
const dolthubReadUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/%s?q=%s"

const homebrewUrlFormat = "https://formulae.brew.sh/api/formula/%s.json"
const homebrewPackage = "dolt"

var authToken string

func get(url string, headers map[string]string) ([]byte, int) {
	client := &http.Client{}

	req, err := http.NewRequest("GET", url, nil)
	for key, value := range headers {
		req.Header.Add(key, value)
	}
	response, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()

	fmt.Println("Downloaded: ", url)

	reader := io.Reader(response.Body)
	data, err := io.ReadAll(reader)
	if err != nil {
		panic(err)
	}

	return data, response.StatusCode
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
	installsIn30days := int(thirtyDays[homebrewPackage].(float64))

	fmt.Println("Total Homebrew Installs in Past 30 Days: " + strconv.Itoa(installsIn30days))

	return installsIn30days
}

// loadFromParameterStore loads an encrypted parameter from AWS SSM Parameter Store.
// We use Parameter Store instead of Secrets Manager since each secret costs $.40 a month to store.
func loadFromParameterStore(parameterName string) (*string, error) {
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		// Optionally use a shared config profile for local development
		//config.WithSharedConfigProfile("jason+test@dolthub.com"),
		// TODO: Make region configurable from infrastructure code
		config.WithRegion("us-west-2"))
	if err != nil {
		return nil, err
	}

	// Create an SSM client
	client := ssm.NewFromConfig(cfg)

	output, err := client.GetParameter(context.TODO(), &ssm.GetParameterInput{
		Name:           &parameterName,
		WithDecryption: true,
	})
	if err != nil {
		return nil, err
	}

	return output.Parameter.Value, nil
}

func main() {
	homebrewUrl := fmt.Sprintf(homebrewUrlFormat, homebrewPackage)
	response, statusCode := get(homebrewUrl, nil)

	// print response body
	if statusCode == http.StatusOK {
		jsonPrettyPrint(response)
	} else {
		fmt.Println(string(response))
		panic("unable to download Homebrew metrics")
	}

	// use json.Unmarshall to load into a map
	var document map[string]interface{}
	json.Unmarshal(response, &document)

	// pull out the 30d installs
	installs := unmarshall30dInstalls(document)

	// Write to DoltHub API
	// TODO: Make parameter name configurable from infra code
	authTokenPointer, err := loadFromParameterStore("dolthub-auth-token")
	if err != nil {
		panic(err)
	}
	authToken = *authTokenPointer

	// TODO: Make repository configurable from infra construct
	owner, repo, fromBranch, toBranch := "jfulghum", "test", "main", url.QueryEscape("homebrew/publish")
	query := fmt.Sprintf("insert into homebrew_metrics values(NOW(), %d);", installs)
	doltHubWriteUrl := fmt.Sprintf(dolthubWriteUrlFormat, owner, repo, fromBranch, toBranch, url.QueryEscape(query))
	sendDoltHubRequest(doltHubWriteUrl)

	// TODO: Poll for write status
	duration, err := time.ParseDuration("1s")
	if err != nil {
		panic(err)
	}
	time.Sleep(duration)

	// Merge DoltHub Change
	doltHubMergeUrl := fmt.Sprintf(dolthubMergeUrlFormat, owner, repo, toBranch, fromBranch)
	sendDoltHubRequest(doltHubMergeUrl)
}

func sendDoltHubRequest(url string) {
	headers := map[string]string{"authorization": authToken}

	doltHubResponse, statusCode := get(url, headers)
	if statusCode == http.StatusOK {
		jsonPrettyPrint(doltHubResponse)
	} else {
		fmt.Println(string(doltHubResponse))
		panic("unable to store results in DoltHub")
	}
}

func jsonPrettyPrint(response []byte) {
	var out bytes.Buffer
	json.Indent(&out, response, "", "   ")
	out.WriteTo(os.Stdout)
}
