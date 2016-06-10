package main

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

func main() {
	excelFileName := "test_excel.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf(err.Error())
	}
	// Print contents of the excel files
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				fmt.Print(cell.Value + " ")
			}
			fmt.Printf("\n")
		}
	}
	// Create maps for each row

	var fileSlice [][][]string
	fileSlice, _ = xlsx.FileToSlice("test_excel.xlsx")

	key := []string{"_id", "name", "age"}
	row := 10 // num of rows of excel files
	col := 3  // num of columns of excel files
	maps := make(map[int]map[string]string)

	for i := 1; i < row; i++ {
		m := map[string]string{} // create a new map for each iteration
		for j := 0; j < col; j++ {
			m[key[j]] = fileSlice[0][i][j]

		}

		fmt.Println(m) // print out each row as a map
		maps[i] = m

	}

	fmt.Println(maps) // a map of maps
}
