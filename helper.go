// Package helper contains helper functions
package helper

import (
	"log"
	"os"
	"strings"
	"time"
)

var (
	e *log.Logger

	//ErrLog   *os.File
	jsonFile *os.File
)

func init() {
	// Open a file for error logs
	errLog, err := os.OpenFile("L:/CVDMC Students/Yilin Xie/data/errorLogs/errlog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	CheckErr(err) // check for errors
	defer errLog.Close()
	// Create a new logger
	e = log.New(errLog, "ERROR: ", 0)
	//Create a json file to store data from reading excel files
	jsonFile, err = os.OpenFile("L:/CVDMC Students/Yilin Xie/data/json/events.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	CheckErr(err) // check for errors
	defer jsonFile.Close()
}

// ChangeDateFormat changes the date format to YYYY-MM-DD
func ChangeDateFormat(x string) string {
	value := strings.Replace(x, "\\", "", -1)
	test, err := time.Parse("02-Jan-06", value)
	CheckErr(err)
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

// AssignStatus assigns a non empty value to the STATUS
func AssignStatus(path string, i int, j int, s1 string, s2 string) {
	if s1 != "" && s2 != "" {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different status values: ", s1, s2)
	} else if s1 == "" {
		s1 = s2
	}
}

// AssignPTID assigns a non empty value to the STATUS
func AssignPTID(path string, i int, j int, d1 string, d2 string) {
	if d1 != "" && d2 != "" {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different PTID Values: ", d1, d2)
	} else if d1 == "" {
		d1 = d2
	}

}

// CheckErr checks errors
func CheckErr(err error) {
	if err != nil {
		e.Println(err)             // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}
