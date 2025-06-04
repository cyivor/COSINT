package handlers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func sendRequest(url, snusKey string, body interface{}) (map[string]interface{}, error) {
	client := &http.Client{}

	var req *http.Request
	var err error

	if body != nil {
		jsonData, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %v", err)
		}
		req, err = http.NewRequest("POST", "https://api.snusbase.com/"+url, bytes.NewBuffer(jsonData))
		if err != nil {
			return nil, fmt.Errorf("failed to create post request: %v", err)
		}
	} else {
		req, err = http.NewRequest("GET", "https://api.snusbase.com/"+url, nil)
		if err != nil {
			return nil, fmt.Errorf("failed to create get request: %v", err)
		}
	}

	req.Header.Set("Auth", snusKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	var result map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %v", err)
	}

	return result, nil
}
