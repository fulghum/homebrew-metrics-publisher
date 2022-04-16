package main

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

const dolthubMergeUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s"
const dolthubWriteUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/write/%s/%s?q=%s"
const dolthubReadUrlFormat = "https://www.dolthub.com/api/v1alpha1/%s/%s/%s?q=%s"

// dolthubAuthToken holds the value loaded from AWS SSM Parameter Store for the AUTH_TOKEN_PARAMETER environment variable
var dolthubAuthToken string

// RunQueryOnNewBranch executes the specified query on a new branch created from a source branch on the specified database.
func RunQueryOnNewBranch(owner, repo, sourceBranch, newBranch, query string) {
	doltHubWriteUrl := fmt.Sprintf(dolthubWriteUrlFormat,
		owner, repo, url.QueryEscape(sourceBranch), url.QueryEscape(newBranch), url.QueryEscape(query))
	sendDoltHubRequest(doltHubWriteUrl)
}

// Merge attempts to merge the tip of fromBranch into toBranch on the specified database.
func Merge(owner, repo, fromBranch, toBranch string) {
	doltHubMergeUrl := fmt.Sprintf(dolthubMergeUrlFormat,
		owner, repo, url.QueryEscape(fromBranch), url.QueryEscape(toBranch))
	sendDoltHubRequest(doltHubMergeUrl)
}

func sendDoltHubRequest(url string) {
	// load the auth token from AWS SSM Parameter Store if it isn't loaded yet
	if len(dolthubAuthToken) == 0 {
		authTokenPointer, err := LoadParameter(dolthubAuthTokenParameterName)
		if err != nil {
			panic(err)
		}
		dolthubAuthToken = *authTokenPointer
	}

	headers := map[string]string{"authorization": dolthubAuthToken}
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
