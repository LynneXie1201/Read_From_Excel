// Package helper contains helper functions
package helper

import (
	"encoding/json"
	"excel/errlog"
	"os"

	"fmt"
	"log"
	"regexp"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

// CheckDateFormat changes the date format to YYYY-MM-DD
func CheckDateFormat(e *log.Logger, path string, sheet int, row int, column string, s string) string {
	value := strings.Replace(s, "\\", "", -1)
	value = strings.Replace(value, ";", "", -1)
	value = strings.Replace(value, "@", "", -1)
	matched1, _ := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])$", value)
	matched2, _ := regexp.MatchString("^(0?[1-9]|[12][0-9]|3[01])-(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-[0-9]{2}$", value)
	matched3, _ := regexp.MatchString("^(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])-[0-9]{2}$", value)
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

// CheckErr checks errors, and print error messages to error logs and screen
func CheckErr(e *log.Logger, err error) {
	if err != nil {
		errlog.PrintErr(e, err)    // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}

// AssignStatus assigns a non empty Status value to the the other one
// if a file has two columns of Status and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignStatus(e *log.Logger, path string, i int, j int, s1 string, s2 string) {
	if s1 != "" && s2 != "" {
		errlog.Differ(e, 0, path, j, i, s1, s2)
	} else if s1 == "" {
		s1 = s2
	}
}

// AssignPTID assigns a non empty PTID value to the the other one
// if a file has two columns of Status and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignPTID(e *log.Logger, path string, i int, j int, d1 string, d2 string) {
	if d1 != "" && d2 != "" {
		errlog.Differ(e, 1, path, j, i, d1, d2)
	} else if d1 == "" {
		d1 = d2
	}

}

// CheckFollowups checks if the excel sheet is a follow_up sheet;
// returns true and a header row if the sheet is a follow_up sheet,
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

/*ExcelToSlice returns a slice of slices of maps for one excel file.
( Assume a excel file may contain multiple sheets)
Each row of a sheet is restructed to a map, then appended to a slice.
Each sheet is restructed to a slice containing list of maps.
*/
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

// Close files
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
