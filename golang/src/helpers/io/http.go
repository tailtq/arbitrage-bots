package io

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Get ... Get request
func Get(url string, data *map[string]interface{}) (*map[string]interface{}, error) {
	res, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(resBody, &data); err != nil {
		return nil, err
	}

	return data, err
}

// Post ... Post request
func Post(url string, body []byte) (*map[string]interface{}, error) {
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return nil, err
	}

	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var data map[string]interface{}
	if err := json.Unmarshal(resBody, &data); err != nil {
		return nil, err
	}

	return &data, err
}
