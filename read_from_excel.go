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
	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				fmt.Printf("%s ", cell.Value)
			}
			fmt.Printf("\n")
		}
	}
}
