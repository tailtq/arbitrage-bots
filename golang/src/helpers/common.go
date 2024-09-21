package helpers

import (
	"fmt"
	"os"
)

func Panic(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicBatch(errs ...error) {
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

func VerboseLog(verbose bool, message string) {
	if verbose {
		fmt.Println(message)
	}
}

func Batch[T any](arr []T, batchSize int) [][]T {
	var batches [][]T

	for batchSize < len(arr) {
		batches = append(batches, arr[:batchSize])
		arr = arr[batchSize:]
	}

	return batches
}

func GetEnv(key string) string {
	return os.Getenv(key)
}
