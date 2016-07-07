// Package helper provides helper functions
package helper

import (
	"encoding/json"
	"excel/errlog"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

// CheckDateFormat checks the date format and returns a date string with format YYYY-MM-DD
func CheckDateFormat(e *log.Logger, path string, sheet int, row int, column string, s string) string {
	// get rid of "\\",";"and "@"in the date strings
	value := strings.Replace(s, "\\", "", -1)
	value = strings.Replace(value, ";", "", -1)
	value = strings.Replace(value, "@", "", -1)
	// YYYY-MM-DD
	matched1, _ := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])$", value)
	// DD-MMM-YY
	matched2, _ := regexp.MatchString("^(0?[1-9]|[12][0-9]|3[01])-(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-[0-9]{2}$", value)
	// MM-DD-YY
	matched3, _ := regexp.MatchString("^(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])-[0-9]{2}$", value)
	// YYYY/MM/DD
	matched4, _ := regexp.MatchString("^[0-9]{4}/(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])$", value)
	if matched1 {
		return value
	} else if matched2 {
		t, err := time.Parse("02-Jan-06", value)
		if err != nil {
			errlog.Differ(e, 5, path, sheet, row, column, value)
		}
		return t.Format("2006-01-02")
	} else if matched3 {
		t, err := time.Parse("01-02-06", value)
		if err != nil {
			errlog.Differ(e, 5, path, sheet, row, column, value)
		}
		return t.Format("2006-01-02")
	} else if matched4 {
		t, err := time.Parse("2006/01/02", value)
		if err != nil {
			errlog.Differ(e, 5, path, sheet, row, column, value)
		}
		return t.Format("2006-01-02")
	}
	errlog.Differ(e, 5, path, sheet, row, column, value)
	return value
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

// CheckErr checks errors, and prints error messages to error logs and screen
func CheckErr(e *log.Logger, err error) {
	if err != nil {
		errlog.PrintErr(e, err)    // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}

// AssignStatus assigns a non empty Status value to the empty one
// if a file has two columns of Status and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignStatus(e *log.Logger, path string, i int, j int, s1 *string, s2 *string) error {
	if *s1 != "" && *s2 != "" {
		errlog.Differ(e, 0, path, j, i, *s1, *s2)
		return fmt.Errorf("The two values are different: %s , %s", *s1, *s2)
	} else if *s1 == "" {
		*s1 = *s2
		return nil
	}
	return nil
}

// AssignPTID assigns a non empty PTID value to  the empty one
// if a file has two columns of PTID and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignPTID(e *log.Logger, path string, i int, j int, d1 *string, d2 *string) error {
	if *d1 != "" && *d2 != "" {
		errlog.Differ(e, 1, path, j, i, *d1, *d2)
		return fmt.Errorf("The two values are different: %s , %s", *d1, *d2)
	} else if *d1 == "" {
		*d1 = *d2
		return nil
	}
	return nil
}

// CheckFollowups checks if the excel sheet is a follow_up sheet.
// Returns true and a header row if the sheet is a follow_up sheet;
// else returns false and nil.
func CheckFollowups(e *log.Logger, path string, j int, sheet *xlsx.Sheet) (bool, []string) {
	// Check if the header row is empty
	v, _ := sheet.Cell(0, 0).String()
	if v == "" {
		errlog.Invalid(e, 2, path, j)
	} else {
		keys := []string{}
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				value, _ := cell.String()
				keys = append(keys, value)
			}
			break
		}
		for _, k := range keys {
			matched, err := regexp.MatchString("STATUS", k)
			CheckErr(e, err)
			if matched && StringInSlice("FU_D", keys) && StringInSlice("DIED", keys) && StringInSlice("DTH_D", keys) {
				return true, keys
			}
		}
		return false, nil
	}
	return false, nil
}

// ExcelToSlice returns a slice of slices of maps for one excel file.
// (Assume a excel file may contain multiple sheets)
// Each row of a sheet is restructed to a map, then appended to a slice,
// and each sheet is restructed to a slice containing list of maps.
func ExcelToSlice(e *log.Logger, excelFilePath string) ([][]map[string]string, [][]string) {

	xlFile, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		fmt.Println(err)
	}
	slices := [][]map[string]string{}
	keyList := [][]string{}
	for s, sheet := range xlFile.Sheets {
		isFu, keys := CheckFollowups(e, excelFilePath, s, sheet) // check for each sheet inside the excel file
		if isFu != false {
			keyList = append(keyList, keys)
			slice := []map[string]string{} // a sheet is a slice
			for _, row := range sheet.Rows {
				m := map[string]string{} // a row is a map
				for j, cell := range row.Cells {
					value, _ := cell.String()
					if value == "9" {
						value = "-9"
					}
					m[keys[j]] = value
				}
				slice = append(slice, m)
			}
			slices = append(slices, slice[1:])
		} else {
			slices = append(slices, nil)
			keyList = append(keyList, nil)
		}
	}
	return slices, keyList
}

// Close is a function that closes a file
func Close(e *log.Logger, filePath string) {
	file, err := os.Open(filePath)
	CheckErr(e, err)
	file.Close()
}

// WriteTOFile writes to json files
func WriteTOFile(jsonFile *os.File, o interface{}) {
	j, err := json.Marshal(o)
	if err != nil {
		fmt.Println(err)
	}
	jsonFile.Write(j)

}

// CheckEmpty checks if value2 is empty or not.
// If value2 is not empty, parse string to int, ahd assign to value1;
// else assign -9 to value1.
func CheckEmpty(value1 *int, value2 string) {
	if value2 != "" {
		*value1, _ = strconv.Atoi(value2)
	} else {
		*value1 = -9
	}
}
