package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func main() {
	searchQuery := "kolnb"
	if len(os.Args) > 1 {
		searchQuery = os.Args[1]
	}

	params := url.Values{}
	params.Add("query", searchQuery)
	params.Add("page", "1")
	params.Add("limit", "10")
	params.Add("filter", "all")

	endpoints := []string{
		fmt.Sprintf("http://localhost:8083/api/v1/search/users?%s", params.Encode()),
		fmt.Sprintf("http://localhost:8083/api/v1/users/search?%s", params.Encode()),
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\nTrying endpoint: %s\n", endpoint)

		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		req.Header.Set("Content-Type", "application/json")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}

		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))

		if resp.StatusCode == 200 {
			break
		}
	}
}
