package main

import (
	"fmt"

	"github.com/tealeg/xlsx"
)

func excelToGo(excelFileName string) {

	//excelFileName := "test_excel.xlsx"
	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf(err.Error())
	}

	var fileSlice [][][]string
	fileSlice, _ = xlsx.FileToSlice(excelFileName) // Create a file slice
	col := xlFile.Sheets[0].MaxCol                 // get the colume number
	keys := []string{}
	for k := 0; k < col; k++ {
		keys = append(keys, fileSlice[0][0][k])

	}

	maps := make(map[int]map[string]string)
	slices := []map[string]string{}

	for _, sheet := range xlFile.Sheets {
		for i, row := range sheet.Rows {
			m := map[string]string{}
			for j, cell := range row.Cells {
				m[keys[j]] = cell.Value
				fmt.Print(cell.Value + " ")

			}
			fmt.Printf("\n")
			maps[i] = m
			slices = append(slices, m)

		}
		fmt.Printf("\n")
	}
	delete(maps, 0) // delete the header line

	fmt.Println(maps)
	fmt.Println(slices[1:])

}

func main() {
	excelToGo("test_excel.xlsx")

}
