package io

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

// Get ... Get request
func Get(url string, responseData *map[string]interface{}) error {
	res, err := http.Get(url)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if err = json.Unmarshal(resBody, &responseData); err != nil {
		return err
	}

	return err
}

// Post ... Post request
func Post(url string, body []byte, responseData interface{}) error {
	res, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}

	defer res.Body.Close()
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return err
	}

	return json.Unmarshal(resBody, &responseData)
}
