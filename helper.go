// Package helper contains helper functions
package helper

import (
	"fmt"
	"strings"
	"time"
)

// ChangeDateFormat changes the date format to YYYY-MM-DD
func ChangeDateFormat(i int, j int, x string) string {
	value := strings.Replace(x, "\\", "", -1)
	test, err := time.Parse("02-Jan-06", value)
	if err != nil {
		fmt.Println(err, "Row#:", i, "Column#:", j)
	}
	return test.Format("2006-01-02")
}

// StringInSlice checks if a slice contains a certain string value
func StringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

// IntInSlice checks if a slice contains a certain int value
func IntInSlice(i int, list []int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}
