package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

func main() {
	// Define the search query
	searchQuery := "kolnb"
	if len(os.Args) > 1 {
		searchQuery = os.Args[1]
	}

	// Create the URL with parameters
	params := url.Values{}
	params.Add("query", searchQuery)
	params.Add("page", "1")
	params.Add("limit", "10")
	params.Add("filter", "all")

	// Build the URL - try both endpoints with API prefix
	endpoints := []string{
		fmt.Sprintf("http://localhost:8083/api/v1/search/users?%s", params.Encode()),
		fmt.Sprintf("http://localhost:8083/api/v1/users/search?%s", params.Encode()),
	}

	for _, endpoint := range endpoints {
		fmt.Printf("\nTrying endpoint: %s\n", endpoint)

		// Create the HTTP request
		req, err := http.NewRequest("GET", endpoint, nil)
		if err != nil {
			fmt.Printf("Error creating request: %v\n", err)
			continue
		}

		// Set headers
		req.Header.Set("Content-Type", "application/json")

		// Make the request
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("Error making request: %v\n", err)
			continue
		}

		// Read the response body
		body, err := ioutil.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			fmt.Printf("Error reading response: %v\n", err)
			continue
		}

		// Print the status code and response body
		fmt.Printf("Status Code: %d\n", resp.StatusCode)
		fmt.Printf("Response Body: %s\n", string(body))

		// If we got a successful response, no need to try other endpoints
		if resp.StatusCode == 200 {
			break
		}
	}
}
