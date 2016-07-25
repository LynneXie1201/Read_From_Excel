// Package helper provides helper functions
package helper

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/aswjh/excel"
	"github.com/tealeg/xlsx"
)

// CheckDateFormat checks the date format and returns a date string with format YYYY-MM-DD
// est == 2 means date is empty or doesn't match any format
func CheckDateFormat(e *log.Logger, path string, sheet int, row int, column string, s string) (string, int) {
	//if date is empty, just return
	if s == "" {
		return s, 2
	}
	// get rid of "\\",";"and "@"in the date strings
	value := strings.Replace(s, "\\", "", -1)
	value = strings.Replace(value, ";", "", -1)
	value = strings.Replace(value, "@", "", -1)
	// YYYY-MM-DD
	matched1, err := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])$", value)
	CheckErr(e, err)
	// DD-MMM-YY
	matched2, err := regexp.MatchString("^(0?[1-9]|[12][0-9]|3[01])-(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-[0-9]{2}$", value)
	CheckErr(e, err)
	// MM-DD-YY
	matched3, err := regexp.MatchString("^(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])-[0-9]{2}$", value)
	CheckErr(e, err)
	// YYYY/MM/DD
	matched4, err := regexp.MatchString("^[0-9]{4}/(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])$", value)
	CheckErr(e, err)
	// M/DD/YY HH:MM
	matched5, err := regexp.MatchString("^(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])/[0-9]{2} ([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$", value)
	CheckErr(e, err)
	// YYYY-M
	matched6, err := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])$", value)
	CheckErr(e, err)
	// YYYY
	matched7, err := regexp.MatchString("^[0-9]{4}$", value)
	CheckErr(e, err)

	if matched1 {
		return value, 0
	} else if matched2 {
		t, err := time.Parse("02-Jan-06", value)
		CheckErr(e, err)
		return t.Format("2006-01-02"), 0
	} else if matched3 {
		t, err := time.Parse("01-02-06", value)
		CheckErr(e, err)
		return t.Format("2006-01-02"), 0
	} else if matched4 {
		t, err := time.Parse("2006/01/02", value)
		CheckErr(e, err)
		return t.Format("2006-01-02"), 0
	} else if matched5 {
		t, err := time.Parse("1/2/06 15:04", value)
		CheckErr(e, err)
		return t.Format("2006-01-02"), 0
	} else if matched6 {
		t, err := time.Parse("2006-1", value)
		CheckErr(e, err)
		newTime := t.Format("2006-01")
		newTime += "-15"
		return newTime, 1
	} else if matched7 {
		newTime := value + "-07-01"
		return newTime, 1
	}
	e.Println(path, "Sheet#:", sheet+1, "Row#:", row+2, "Column:", column, "INFO: Invalid Format of Date:", value)
	//errlog.Differ(e, 5, path, sheet, row, column, value)
	// return value
	return value, 3
}

// StringInSlice checks if a string in the slice matches a certain string pattern
func StringInSlice(indicator int, str string, list []string) bool {
	// if indicator is 0, str is the standard string pattern
	if indicator == 0 {
		for _, v := range list {
			matched, _ := regexp.MatchString("^"+str+"$", v)
			if matched {
				return true
			}
		}
		return false
	}
	// indicator is not 0, then v is the standard string pattern
	for _, v := range list {
		matched, _ := regexp.MatchString("^"+v, str)
		if matched {
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

// FloatInSlice checks if a slice contains a certain float value
func FloatInSlice(i float64, list []float64) bool {
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
		e.Println(err)             // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}

// AssignStatus assigns a non empty Status value to the empty one
// if a file has two columns of Status and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignStatus(s1 *string, s2 *string) bool {
	if *s1 == *s2 {
		return false
	} else if *s1 != "" && *s2 != "" {
		return true
	} else if *s1 == "" {
		*s1 = *s2
		return false
	}
	return false
}

// AssignPTID assigns a non empty PTID value to  the empty one
// if a file has two columns of PTID and one of them is empty;
// reports an error message if two columns have different values and none of them is empty.
func AssignPTID(d1 *string, d2 *string) bool {
	if *d1 == *d2 {
		return false
	} else if *d1 != "" && *d2 != "" {
		return true
		//errlog.Differ(e, 1, path, j, i, *d1, *d2)
		//return fmt.Errorf("The two values are different: %s , %s", *d1, *d2)
	} else if *d1 == "" {
		*d1 = *d2
		return false
	}
	return false
}

// CheckEmptyHeaderRow checks if the header rows cannot be read,
// then use a package to open the file and save it.
func CheckEmptyHeaderRow(e *log.Logger, excelFilePath string) (bool, *xlsx.File) {
	// Open an excel file
	File, err := xlsx.OpenFile(excelFilePath)
	CheckErr(e, err) // if error exists, write to errlog and terminates
	// assign the A1 cell to v
	v, _ := File.Sheets[0].Cell(0, 0).String()
	// Check if the header row is empty,
	// if the header rows cannot be read, use a package to open the file and then save it.
	if v == "" {
		option := excel.Option{"Visible": false, "DisplayAlerts": true}
		xl, err := excel.Open(excelFilePath, option)
		CheckErr(e, err) // if error exists, write to errlog and terminates
		defer xl.Quit()
		xl.Save()
		xl.Quit()
		return true, nil
	}
	return false, File

}

// CheckFollowups checks if the excel sheet is a follow_up sheet.
// Returns true and a header row if the sheet is a follow_up sheet;
// else returns false and nil.
func CheckFollowups(e *log.Logger, path string, j int, sheet *xlsx.Sheet) (bool, []string) {

	// Check if the header row is empty
	v, _ := sheet.Cell(0, 0).String()
	if v == "" {
		// if the header row is empty, then write to errlog
		e.Println(path, "Sheet #:", j+1, "THIS SHEET DOES NOT HAVE HEADER ROW!")
		return false, nil
		// ignore files if A1 is "IGNORE"
	} else if v == "IGNORE" {
		return false, nil
	}
	keys := []string{}
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			value, _ := cell.String()
			keys = append(keys, value)
		}
		break
	}
	// Check follow up columns
	if StringInSlice(0, "FU_D", keys) && StringInSlice(0, "DIED", keys) && StringInSlice(0, "DTH_D", keys) {
		return true, keys
	}
	return false, nil
}

// ExcelToSlice returns a slice of slices of maps for one excel file.
// (Assume a excel file may contain multiple sheets)
// Each row of a sheet is restructed to a map, then appended to a slice,
// and each sheet is restructed to a slice containing list of maps.
func ExcelToSlice(e *log.Logger, excelFilePath string, columnsChecker string) ([][]map[string]string, [][]string) {

	isEmpty, xlFile := CheckEmptyHeaderRow(e, excelFilePath)
	// if the excel file has a empty header row
	if isEmpty {
		xlFile, _ = xlsx.OpenFile(excelFilePath)
	}
	slices := [][]map[string]string{}
	keyList := [][]string{}
	// s is the index of Sheets
	for s, sheet := range xlFile.Sheets {
		isFu, keys := CheckFollowups(e, excelFilePath, s, sheet) // check for each sheet inside the excel file
		if isFu != false {
			// Check columnn names

			CheckColumnNames(columnsChecker, e, keys, excelFilePath, s)

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

// CheckIntValue checks int value:
// If value2 is empty, assign -9 to value1 and return true;
// if value2 is character, assign -9 to value1 and return false;
// else parse string value2 to int value1
func CheckIntValue(value1 *int, value2 string) bool {

	matched, _ := regexp.MatchString("^[A-Za-z]$", value2)

	if value2 == "" {
		*value1 = -9
		return true
	} else if matched {
		*value1 = -9
		return false
	}
	*value1, _ = strconv.ParseInt(value2)
	return true
}

// CheckFloatValue checks float value:
// If value2 is not empty, parse string to int, ahd assign to value1;
// else assign -9 to value1.
func CheckFloatValue(value1 *float64, value2 string) bool {

	matched, _ := regexp.MatchString("^[A-Za-z]$", value2)
	if value2 == "" {
		*value1 = -9
		return true

	} else if matched {
		*value1 = -9
		return false
	}
	*value1, _ = strconv.ParseFloat(value2, 64)
	return true
}

// CheckStringValue is
func CheckStringValue(value1 *string, value2 string) bool {

	matched, _ := regexp.MatchString("^[A-Za-z]$", value2)
	if value2 == "" {
		value1 = nil
		return true
	} else if !matched {
		return false
	}
	*value1 = value2
	return true
}

// CheckPtidColumns checks the number of PTID columns,
// and returns the column names of PTID, assuming each file would have at most two PTID columns.
// Parameters including:
// e - a logger that records the error messages;
// path - the path of the excel file;
// j - the index of the sheets in that excel file;
// keys - a slice that contains the header row
func CheckPtidColumns(e *log.Logger, path string, j int, keys []string) (string, string) {
	// create a slice that holds the column names that contains PTID
	id := []string{}
	for _, k := range keys {
		if strings.Contains(k, "PTID") {
			id = append(id, k)
		}
	}
	// if len is 2, we have 2 columns of PTID
	if len(id) == 2 {
		id1, id2 := id[0], id[1]
		return id1, id2
		// if len is 1, we have only one column of PTID
	} else if len(id) == 1 {
		id1, id2 := id[0], id[0]
		return id1, id2
	}
	// else would be invaid as we assume each file would have at most two PTID columns,
	// then an error message gets written and the program stops.
	e.Println(path, "Sheet #:", j+1, "INFO: This file has invalid numbers of PTID columns!")
	os.Exit(1) // exit if it has invaid columns of PTID
	return "", ""
}

// CheckStatusColumns checks the number of STATUS columns,
// and returns the column names of STATUS, assuming each file would have at most two STATUS columns.
// Parameters including:
// e - a logger that records the error messages;
// path - the path of the excel file;
// j - the index of the sheets in that excel file;
// keys - a slice that contains the header row
func CheckStatusColumns(e *log.Logger, path string, j int, keys []string) (string, string) {
	// create a slice that holds the column names of STATUS that matches a certain pattern
	status := []string{}
	for _, k := range keys {
		matched, err := regexp.MatchString("^.*STATUS$", k) // check status's pattern
		CheckErr(e, err)
		if matched {
			status = append(status, k)
		}
	}
	// if len is 2, we have 2 columns of STATUS
	if len(status) == 2 {
		s1, s2 := status[0], status[1]
		return s1, s2
		// if len is 1, we have only one column of STATUS
	} else if len(status) == 1 {
		s1, s2 := status[0], status[0]
		return s1, s2
	}
	// else would be invaid as we assume each file would have at most two STATUS columns,
	// then an error message gets written and the program stops.
	e.Println(path, "Sheet #:", j+1, "INFO: This file has invalid numbers of Status columns!")
	os.Exit(1)
	return "", ""
}

// CheckPtidFormat checks if the format of PTID is LLLFDDMMYY;
// if not, write to error log.
func CheckPtidFormat(id string, e *log.Logger, path string, j int, i int) bool {
	// valid PTID format: LLLFMMDDYY
	matched, err := regexp.MatchString("^.{4}(0?[1-9]|1[012])(0?[1-9]|[12][0-9]|3[01])[0-9]{2}$", id)
	CheckErr(e, err)
	if !matched {
		e.Println(path, "Sheet #:", j+1, "Row #:", i+2, "INFO: Invaid PTID Value:", id)
		return false
	}
	return true
}

// CheckColumnNames checks if the columns are expected ones
func CheckColumnNames(file string, e *log.Logger, keys []string, path string, j int) {
	// read from the columns file
	columns, err := ReadLines(file)

	CheckErr(e, err)
	for _, k := range keys {

		if !StringInSlice(1, k, columns) {
			e.Println(path, "Sheet #:", j+1, "INFO: Unexpected Column:", k)
		}
	}
}

// ReadLines reads a whole file into memory
// and returns a slice of its lines.
func ReadLines(path string) ([]string, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	return lines, scanner.Err()
}

//GetUserInput reads user input from terminal
func GetUserInput() string {
	// get standard column names file path from user input
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("Enter path to the columns file: ") // C:/Users/Lynne Xie/Documents/go/src/excel/helper/column.txt
	file, _ := reader.ReadString('\n')
	file = strings.TrimSpace(file)
	return file
}

// CompareDates is  date1 - date2
func CompareDates(e *log.Logger, date1 string, date2 string) int {
	d1, err := time.Parse("2006-01-02", date1)
	CheckErr(e, err)
	d2, err := time.Parse("2006-01-02", date2)
	CheckErr(e, err)
	// diff equals d1 - d2
	diff := d1.Sub(d2)
	days := int(diff.Hours() / 24)
	if days >= 0 && days <= 30 {
		return 1
	}
	return 2
}

// SubPath returns a sub path form sep
func SubPath(path string, sep string) string {
	i := strings.Index(path, sep)
	sub := path[i+len(sep) : len(path)]
	return sub
}

// OperationString is
func OperationString(reason string, survival string, notes string, surgery string, nonvalve string) string {
	var s string
	if nonvalve == "" {
		s = "Reason: " + reason + ", Surgeries: " + surgery + ", Notes: " + notes + ", Survival = " + survival
		return s
	} else if reason == "" && survival == "" && notes == "" && surgery == "" {
		s = "Nonvalve re-op: " + nonvalve
		return s
	}
	s = "Reason: " + reason + ", Surgeries: " + surgery + ", Notes: " + notes + ", Survival = " + survival + ", Nonvalve re-op: " + nonvalve
	return s

}
