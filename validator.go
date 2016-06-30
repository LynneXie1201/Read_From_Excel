// Package validator contains validate functions related to excel files
package validator

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"regexp"

	"strings"

	"github.com/tealeg/xlsx"
)

var (
	// ID1 US
	ID1 string
	// ID2 IS
	ID2 string
	// S1 IS
	S1 string
	// S2 IS
	S2 string
	e  *log.Logger
)

func init() {
	// Open a file for error logs
	errLog, err := os.OpenFile("L:/CVDMC Students/Yilin Xie/data/errorLogs/errlog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	helper.CheckErr(err) // check for errors
	defer errLog.Close()
	// Create a new logger
	e = log.New(errLog, "ERROR: ", 0)
	//Create a json file to store data from reading excel files
	jsonFile, err := os.OpenFile("L:/CVDMC Students/Yilin Xie/data/json/events.json", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	helper.CheckErr(err) // check for errors
	defer jsonFile.Close()
}

// CheckFollowups checks if the excel sheet is a follow_up sheet and return boolean value the header row
func CheckFollowups(path string, j int, sheet *xlsx.Sheet) (bool, []string) {
	// Check if the header row is empty
	v, _ := sheet.Cell(0, 0).String()
	if v == "" {
		e.Println(path, "Sheet #:", j, "THIS SHEET DOES NOT HAVE HEADER ROW!")
	} else {
		keys := []string{}
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				value, _ := cell.String()
				//fmt.Println(value)
				keys = append(keys, value)
			}
			break
		}
		if helper.StringInSlice("FU_D", keys) && helper.StringInSlice("DIED", keys) && helper.StringInSlice("DTH_D", keys) {
			return true, keys
		}
		//fmt.Println(keys)
		return false, nil
	}
	return false, nil
}

/*ExcelToSlice is a function that returns a slice of slices of maps for one excel file.
( One file may contain multiple sheets)
Each row of a sheet is restructed to a map, then appended to a slice.
Each sheet is restructed to a slice containing list of maps.
*/
func ExcelToSlice(excelFilePath string) ([][]map[string]string, [][]string) {

	xlFile, err := xlsx.OpenFile(excelFilePath)
	helper.CheckErr(err) // check for errors
	//fmt.Println(xlFile.ToSlice())
	slices := [][]map[string]string{}
	keyList := [][]string{}
	for j, sheet := range xlFile.Sheets {

		isFu, keys := CheckFollowups(excelFilePath, j, sheet) // check for each sheet inside the excel file
		if isFu != false {
			keyList = append(keyList, keys)
			slice := []map[string]string{} // a sheet is a slice
			for _, row := range sheet.Rows {
				m := map[string]string{} // a row is a map
				for j, cell := range row.Cells {
					value, _ := cell.String()
					if strings.Contains(value, "\\") {
						value = helper.ChangeDateFormat(value)
					}
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

// CheckStatus checks STATUS
func CheckStatus(path string, j int, keys []string) {
	status := []string{}
	for _, k := range keys {
		matched, err := regexp.MatchString("^.*STATUS$", k) // check status's pattern
		helper.CheckErr(err)
		if matched {
			status = append(status, k)
		}
	}
	if len(status) == 2 {
		S1, S2 = status[0], status[1]
	} else if len(status) == 1 {
		S1, S2 = status[0], status[0]
	} else {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of STATUS!")
		fmt.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of STATUS!")
		os.Exit(1)
	}
}

// CheckPTID checks PTIDs
func CheckPTID(path string, j int, keys []string) {
	id := []string{}
	for _, k := range keys {
		if strings.Contains(k, "PTID") {
			id = append(id, k)
		}
	}
	if len(id) == 2 {
		ID1, ID2 = id[0], id[1]
	} else if len(id) == 1 {
		ID1, ID2 = id[0], id[0]
	} else {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID!")
		fmt.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID!")
		os.Exit(1) // exit if it has invaid columns of PTID
	}

}
