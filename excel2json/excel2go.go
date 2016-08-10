package excel2json

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/tealeg/xlsx"
)

// LoopAllFiles recursively loops all files in a folder, and tracks all excel files,
// opens a errorlog and a json file to store error messages and json objects, and
// for each excel file, calls another function to read data from the file.
func LoopAllFiles(e *log.Logger, dirPath string, jsonFile *os.File) {
	fileList := []string{}
	filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
		} else if !f.IsDir() && strings.Contains(f.Name(), "xlsx") {
			fileList = append(fileList, path)
		}
		return nil
	})
	// get the valid column names
	columnsChecker := helper.ReadUserInput()
	// Loop through all excel files
	for _, file := range fileList {
		ReadExcelData(e, file, jsonFile, columnsChecker)
	}

	// write to JSON file
	WriteToJSON(jsonFile, allARH, allDVT, allDths, allFUMI, allFUPACE, allFix, allFollowUps,
		allLKA, allOperation, allPVL, allSBE, allSVD,
		allStroke, allTHRM, allTIA, alllHEML, allLostFollowups)
}

// ExcelToSlice returns a slice of slices of maps for one excel file.
// (Assume a excel file may contain multiple sheets)
// Each row of a sheet is restructed to a map, then appended to a slice,
// and each sheet is restructed to a slice containing list of maps.
func ExcelToSlice(e *log.Logger, excelFilePath string, columnsChecker string) ([][]map[string]string, [][]string) {

	// Check if the file has a header row that cannot be read due to some reasons
	unreadable, xlFile := helper.CheckHeaderRow(e, excelFilePath)
	// if the excel file has a header row that cannot be read
	if unreadable {
		xlFile, _ = xlsx.OpenFile(excelFilePath)
	}
	slices := [][]map[string]string{}
	keyList := [][]string{}
	// s is the index of Sheets
	for s, sheet := range xlFile.Sheets {
		// check to see if a sheet is a followup sheet
		isFu, keys := helper.CheckFollowups(e, excelFilePath, s, sheet)

		// if the sheet is a followup sheet
		if isFu {
			// check if columnn names are the expected ones
			helper.CheckColumnNames(columnsChecker, e, keys, excelFilePath, s)

			keyList = append(keyList, keys)
			slice := []map[string]string{} // a sheet is a slice
			for _, row := range sheet.Rows {
				m := map[string]string{} // a row is a map
				for j, cell := range row.Cells {
					value, _ := cell.String()
					// change all number 9 to -9
					if value == "9" {
						value = "-9"
					}
					m[keys[j]] = value
				}
				slice = append(slice, m)
			}
			slices = append(slices, slice[1:])
			// else if the sheet is not a followup sheet
		} else {
			slices = append(slices, nil)
			keyList = append(keyList, nil)
		}
	}
	return slices, keyList
}
