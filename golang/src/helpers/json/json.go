package json

import (
	"encoding/json"
	"os"
)

func Unmarshal(data []byte, v any) error {
    return json.Unmarshal(data, v)
}

func ReadJSONFile(filePath string, v any) error {
    jsonData, _ := os.ReadFile(filePath)
    err := Unmarshal(jsonData, v)

    return err
}

func WriteJSONFile(filePath string, v any) error {
    jsonData, _ := json.MarshalIndent(v, "", "\t")
    err := os.WriteFile(filePath, jsonData, 0644)
    
    return err
}
