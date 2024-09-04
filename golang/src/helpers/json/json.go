package json

import (
	"encoding/json"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"os"
	"strings"
)

func Unmarshal(data []byte, v any) error {
	return json.Unmarshal(data, v)
}

func Marshal(v any) (string, error) {
	data, err := json.Marshal(v)
	return string(data), err
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

func ReadJSONABIFile(filePath string) (abi.ABI, error) {
	jsonData, err := os.ReadFile(filePath)

	if err == nil {
		abiResult, err := abi.JSON(strings.NewReader(string(jsonData)))
		return abiResult, err
	}

	return abi.ABI{}, err
}
