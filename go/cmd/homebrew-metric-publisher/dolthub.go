package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
)

const dolthubMergeUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s"
const dolthubWriteUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s?q=%s"
const dolthubReadUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/%s?q=%s"

var dolthubAuthToken string

// RunQueryOnBranch executes the specified query on a new toBranch created from fromBranch on the specified database.
func RunQueryOnBranch(owner, repo, fromBranch, toBranch, query string) {
	doltHubWriteUrl := fmt.Sprintf(dolthubWriteUrlFormat, owner, repo, fromBranch, toBranch, url.QueryEscape(query))
	sendDoltHubRequest(doltHubWriteUrl)
}

// Merge attempts to Merge the specified toBranch to fromBranch on the specified database.
func Merge(owner, repo, toBranch, fromBranch string) {
	doltHubMergeUrl := fmt.Sprintf(dolthubMergeUrlFormat, owner, repo, toBranch, fromBranch)
	sendDoltHubRequest(doltHubMergeUrl)
}

func sendDoltHubRequest(url string) {
	headers := map[string]string{"authorization": dolthubAuthToken}

	// load the auth token from AWS SSM Parameter Store if it isn't loaded yet
	if len(dolthubAuthToken) > 0 {
		authTokenPointer, err := LoadParameter(os.Getenv(dolthubAuthTokenParameterNameEnv))
		if err != nil {
			panic(err)
		}
		dolthubAuthToken = *authTokenPointer
	}

	doltHubResponse, statusCode := get(url, headers)
	if statusCode == http.StatusOK {
		jsonPrettyPrint(doltHubResponse)
	} else {
		fmt.Println(string(doltHubResponse))
		panic("unable to store results in DoltHub")
	}
}

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
