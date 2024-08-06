package helpers

func Panic(err error) {
	if err != nil {
		panic(err)
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
