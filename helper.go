package helper

// Generate error messages to a file
func newError(path string, id string, row int, t string, field string, invalid string) {

	e.Println(path, "PTID:", id, "Row #:", row+2, "Type:", t, "Info: Invalid", field, "Value:", invalid)

}

// Change the date Format to YYYY-MM-DD
func changeDateFormat(x string) string {
	value := strings.Replace(x, "\\", "", -1)
	test, _ := time.Parse("02-Jan-06", value)
	return test.Format("2006-01-02")
}

// a function that writes to json files
func writeTOFile(o interface{}) {

	j, _ := json.Marshal(o)
	jsonFile.Write(j)

}

// Check if a slice contains a certain string value
func stringInSlice(str string, list []string) bool {
	for _, v := range list {
		if v == str {
			return true
		}
	}
	return false
}

func intInSlice(i int, list []int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

// Check if the excel file is a follow_up file and return the header row
func checkFollowups(file *xlsx.File) (bool, []string) {

	keys := []string{}
	for _, sheet := range file.Sheets {
		for _, row := range sheet.Rows {
			for _, cell := range row.Cells {
				value, _ := cell.String()
				keys = append(keys, value)
			}
			break
		}
	}
	if stringInSlice("FU_D", keys) && stringInSlice("DIED", keys) && stringInSlice("DTH_D", keys) {
		return true, keys
	}
	return false, nil
}

func excelToSlice(excelFilePath string) ([]map[string]string, []string) {

	xlFile, err := xlsx.OpenFile(excelFilePath)
	if err != nil {
		fmt.Printf(err.Error())
	}
	// Check if the excel file is a followup file
	value, keys := checkFollowups(xlFile)
	if value == true {
		slices := []map[string]string{} // each row is a map
		for _, sheet := range xlFile.Sheets {
			for _, row := range sheet.Rows {
				m := map[string]string{}
				for j, cell := range row.Cells {
					value, _ := cell.String()
					if strings.Contains(value, "\\") {
						value = changeDateFormat(value)
					}
					if value == "9" {
						value = "-9"
					}
					m[keys[j]] = value
				}
				slices = append(slices, m)
			}
		}
		return slices[1:], keys // return the followup excel file as a slice
	}
	return nil, keys // return nil if the excel file isn't a follow up file
}


func checkPTID(path string, keys []string) {
	id := []string{}
	for _, k := range keys {
		if strings.Contains(k, "PTID") {
			id = append(id, k)
		}
	}
	if len(id) == 2 {
		id1, id2= id[0], id[1]

	} else if len(id) == 1 {
		id1, id2 = id[0],  id[0]

	} else {
		e.Println(path, "INFO: This file has invalid numbers of PTID!")
		os.Exit(1) // exit if it has invaid columns of PTID
	}

}

func checkStatus(path string, keys []string) {
	status := []string{}
	for _, k := range keys {
		 matched, err := MatchString(".STATUS", k) // check status's pattern
		 if matched {
			status = append(status, k)
		}
	}
	// Ends here!
	if len(id) == 2 {
		s1, s2 = status[0], status[1]

	} else if len(id) == 1 {
		s1, s2 = status[0], status[0]

	} else {
		e.Println(path, "INFO: This file has invalid numbers of STATUS!")
		os.Exit(1)
	}
}

func assignNonEmptyStatus(path string, i int, s1 string, s2 string){
	if s1 != "" && s2 != ""{
		e.Println(path, "row #: ",i,"INFO: different status values!")
	} else if s1 == ""{
		s1 = s2
	}

}
