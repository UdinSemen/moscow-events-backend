package utils

import (
	"math/rand"
	"strconv"
	"time"
)

func GenTimeCode(length int) string {
	r := rand.New(rand.NewSource(time.Now().Unix()))

	var timeCode string
	for _, value := range r.Perm(length) {
		stringValue := strconv.Itoa(value)
		timeCode += stringValue
	}

	return timeCode
}
