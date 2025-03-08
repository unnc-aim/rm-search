package common

import (
	"os"
	"strconv"
)

func IsProd() bool {
	prod := os.Getenv("PROD")
	b, _ := strconv.ParseBool(prod)
	return b
}
