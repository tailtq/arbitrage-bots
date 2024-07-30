package io

import (
	"encoding/json"
	"io"
	"net/http"
)

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
