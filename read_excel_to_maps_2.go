package main

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

func main() {

	key := []string{"_id", "name", "age"}

	maps := make(map[int]map[string]string)

	excelFileName := "test_excel.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf(err.Error())
	}

	for _, sheet := range xlFile.Sheets {
		for i, row := range sheet.Rows {
			m := map[string]string{}
			for j, cell := range row.Cells {
				m[key[j]] = cell.Value
				fmt.Print(cell.Value + " ")

			}
			fmt.Printf("\n")
			maps[i] = m
		}
		fmt.Printf("\n")
	}
	delete(maps, 0) // delete the header line
	fmt.Println(maps)

}
