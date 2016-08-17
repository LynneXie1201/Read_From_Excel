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

// CheckDateFormat checks the date format and returns a date string with the format YYYY-MM-DD, and an int indicator:
// indicator equals 0 means the original date is parsed to the format YYYY-MM-DD correctly;
// indicator equals 1 means the original date is missing some parts and now been fixed;
// indicator equals 2 means the original date is empty;
// indicator equals 3 means the original date has an invalid format that cannot be parsed to YYYY-MM-DD.
func CheckDateFormat(e *log.Logger, path string, sheet int, row int, column string, s string) (string, int) {
	//if date is empty, just return empty string and indicator equals 2
	if s == "" {
		return s, 2
	}
	// if date is not empty, first get rid of "\\", ";" and "@" that the date strings contain
	value := strings.Replace(s, "\\", "", -1)
	value = strings.Replace(value, ";", "", -1)
	value = strings.Replace(value, "@", "", -1)

	// original date with format YYYY-MM-DD
	matched1, err := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])$", value)
	CheckErr(e, err)

	// original date with format DD-MMM-YY
	matched2, err := regexp.MatchString("^(0?[1-9]|[12][0-9]|3[01])-(Jan|Feb|Mar|Apr|May|Jun|Jul|Aug|Sep|Oct|Nov|Dec)-[0-9]{2}$", value)
	CheckErr(e, err)

	// original date with format MM-DD-YY
	matched3, err := regexp.MatchString("^(0?[1-9]|1[012])-(0?[1-9]|[12][0-9]|3[01])-[0-9]{2}$", value)
	CheckErr(e, err)

	// original date with format YYYY/MM/DD
	matched4, err := regexp.MatchString("^[0-9]{4}/(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])$", value)
	CheckErr(e, err)

	// original date with format M/DD/YY HH:MM
	matched5, err := regexp.MatchString("^(0?[1-9]|1[012])/(0?[1-9]|[12][0-9]|3[01])/[0-9]{2} ([0-9]|0[0-9]|1[0-9]|2[0-3]):[0-5][0-9]$", value)
	CheckErr(e, err)

	// original date with format YYYY-M
	matched6, err := regexp.MatchString("^[0-9]{4}-(0?[1-9]|1[012])$", value)
	CheckErr(e, err)

	// original date with format YYYY
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
	// return the original date and indicator equals 3
	return value, 3
}

// StringInSlice checks if a string in a slice matches a certain string pattern.
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
	// indicator is not 0, then v from the slice is the standard string pattern
	for _, v := range list {
		matched, _ := regexp.MatchString("^"+v, str)
		if matched {
			return true
		}
	}
	return false
}

// IntInSlice checks if a slice contains a certain int value.
func IntInSlice(i int, list []int) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

// FloatInSlice checks if a slice contains a certain float value.
func FloatInSlice(i float64, list []float64) bool {
	for _, v := range list {
		if v == i {
			return true
		}
	}
	return false
}

// CheckErr checks errors, prints error messages to the errorlog and then terminate.
func CheckErr(e *log.Logger, err error) {
	if err != nil {
		e.Println(err)             // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}

// AssignStatus returns false and assigns the non-empty status value
// to the empty one when a file has two columns of status and one of them is empty;
// returns true if the two statuses have different values and none of them is empty.
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

// AssignPTID returns false and assigns the non-empty PTID value
// to the empty one when a file has two columns of PTIDs and one of them is empty;
// returns true if the two PTIDs have different values and none of them is empty.
func AssignPTID(d1 *string, d2 *string) bool {
	if *d1 == *d2 {
		return false
	} else if *d1 != "" && *d2 != "" {
		return true

	} else if *d1 == "" {
		*d1 = *d2
		return false
	}
	return false
}

// CheckHeaderRow checks if the header rows can be read.
// If not, open the file and save it, and return true and nil;
// else return false and the excel file.
func CheckHeaderRow(e *log.Logger, excelFilePath string) (bool, *xlsx.File) {
	// Open an excel file
	File, err := xlsx.OpenFile(excelFilePath)
	// if error exists, write to errlog and terminates
	CheckErr(e, err)
	// assign the A1 cell of the first sheet to v
	v, _ := File.Sheets[0].Cell(0, 0).String()
	// Check if the header row is empty.
	// If the header rows cannot be read, open the file and then save it,
	// and return true and nil
	if v == "" {
		option := excel.Option{"Visible": false, "DisplayAlerts": true}
		// open the file
		xl, err := excel.Open(excelFilePath, option)
		CheckErr(e, err)
		defer xl.Quit()
		// save the file
		xl.Save()
		xl.Quit()
		return true, nil
	}
	// return false and the excel file if the file is readable
	return false, File

}

// CheckFollowups checks if the excel sheet is an empty sheet,
// a follow_up sheet or a sheet that should be ignored.
// Return true and a header row if the sheet is a follow_up sheet;
// else return false and nil.
func CheckFollowups(e *log.Logger, path string, j int, sheet *xlsx.Sheet) (bool, []string) {

	// assign the string value of A1 cell to v
	v, _ := sheet.Cell(0, 0).String()

	// if v equals empty string, write to errlog and return false, nil;
	// if v equals "IGNORE", it means that this sheet should be skipped
	if v == "" {
		e.Println(path, "Sheet #:", j+1, "THIS SHEET DOES NOT HAVE HEADER ROW!")
		return false, nil
		// ignore files if A1 is "IGNORE"
	} else if v == "IGNORE" {
		return false, nil
	}
	// if cell A1 is neither empty nor "IGNORE",
	// use the slice keys to collect header row
	keys := []string{}
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			value, _ := cell.String()
			keys = append(keys, value)
		}
		break
	}
	// Check if the sheet is a follow up sheet by checking if header row contains "FU_D", "DIED" and "DTH_D"
	// if not, returns false and nil;
	// else returns true and keys
	if StringInSlice(0, "FU_D", keys) && StringInSlice(0, "DIED", keys) && StringInSlice(0, "DTH_D", keys) {
		return true, keys
	}
	return false, nil
}

// Close is a function that closes a file
func Close(e *log.Logger, filePath string) {
	file, err := os.Open(filePath)
	CheckErr(e, err)
	file.Close()
}

// WriteTOFile writes JSON objects to json files
func WriteTOFile(jsonFile *os.File, o interface{}) {
	j, err := json.Marshal(o)
	if err != nil {
		fmt.Println(err)
	}
	jsonFile.Write(j)

}

// CheckIntValue returns true and assigns -9 to value1 if value2 is empty;
// returns true and assigns value2 to value1 if value2 if a number;
// else assign -9 to value1 and return false.
func CheckIntValue(value1 *int, value2 string, list []int) bool {

	matched, _ := regexp.MatchString("^([-]?[0-9]+[.]?5?)$", value2)

	if value2 == "" {
		*value1 = -9
		return true
	} else if matched {
		*value1, _ = strconv.Atoi(value2)
		if IntInSlice(*value1, list) {
			return true
		}
		return false
	}
	*value1 = -9
	return false
}

// CheckFloatValue returns true and assigns -9 to value1 if value2 is empty;
// returns true and assigns value2 to value1 if value2 if a number;
// else assign -9 to value1 and return false.
func CheckFloatValue(value1 *float64, value2 string, list []float64) bool {

	matched, _ := regexp.MatchString("^([-]?[0-9]+[.]?5?)$", value2)

	if value2 == "" {
		*value1 = -9
		return true
	} else if matched {
		*value1, _ = strconv.ParseFloat(value2, 64)
		if FloatInSlice(*value1, list) {
			return true
		}
		return false
	}
	*value1 = -9
	return false
}

// CheckStringValue returns true if a value only contains characters.
func CheckStringValue(value string) bool {

	matched, _ := regexp.MatchString("^[A-Za-z]*$", value)
	if matched {
		return true
	}
	return false
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

// CheckStatusColumns checks the number of status columns, and returns the column
// names of status, assuming each file would have at most two status columns.
// Parameters including:
// e - a logger that records the error messages;
// path - the path of the excel file;
// j - the index of the sheets in that excel file;
// keys - a slice that contains the header row
func CheckStatusColumns(e *log.Logger, path string, j int, keys []string) (string, string) {
	// create a slice that holds the column names of status that matches a certain pattern
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
// if not, write to the errorlog and return false;
// else return true.
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

// CheckColumnNames checks if the columns are the expected ones;
// if not, print to the errorlog
func CheckColumnNames(file string, e *log.Logger, keys []string, path string, j int) {
	// read from the columns file
	columns, err := ReadLines(file)
	CheckErr(e, err)
	// keys are from the header row of the excel file
	for _, k := range keys {
		if !StringInSlice(1, k, columns) {
			e.Println(path, "Sheet #:", j+1, "INFO: Unexpected Column:", k)
		}
	}
}

// ReadLines reads a whole file from path into memory,
// and returns a slice of strings of its lines.
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

// ReadUserInput reads and returns the string value of the user input.
func ReadUserInput() string {
	// get the file path of standard column names from user input
	reader := bufio.NewReader(os.Stdin)
	// should enter the file path (/excel/helper/ignore.txt)
	// to the columns names that are valid
	fmt.Print("Enter path to the columns file: ")
	file, _ := reader.ReadString('\n')
	file = strings.TrimSpace(file)
	return file
}

// SubPath returns a sub path string from the full length path string
// by detect the index of keyword.
func SubPath(path string, keyword string) string {
	i := strings.Index(path, keyword)
	sub := path[i+len(keyword):]
	return sub
}

// OperationNotes returns a full-text meaning operation notes according to the code book
func OperationNotes(reason string, survival string, notes string, surgery string, nonvalve string) string {
	var s string
	if nonvalve == "" {
		s = "Reason: '" + reason + "', Surgeries: '" + surgery + "', Notes: '" + notes + "', Survival: '" + survival + "'"
		return s
	} else if reason == "" && survival == "" && notes == "" && surgery == "" {
		s = "Nonvalve re-op: '" + nonvalve + "'"
		return s
	}
	s = "Reason: '" + reason + "', Surgeries: '" + surgery + "', Notes: '" + notes +
		"', Survival: '" + survival + "', Nonvalve re-op: '" + nonvalve + "'"
	return s

}

// DateLaterThan returns true if date1 is later than date2.
func DateLaterThan(date1 string, date2 string) bool {

	d1, _ := time.Parse("2006-01-02", date1)
	d2, _ := time.Parse("2006-01-02", date2)

	return d1.After(d2)
}

// CompareDates returns the result of comparing date1 and date2.
// Return 1 if date1 is within 30 days of date2;
// else return 2.
func CompareDates(e *log.Logger, date1 string, date2 string) int {
	// parse string to date
	d1, err := time.Parse("2006-01-02", date1)
	CheckErr(e, err)
	d2, err := time.Parse("2006-01-02", date2)
	CheckErr(e, err)

	// diff equals d1 minus d2
	diff := d1.Sub(d2)
	// transfer to days
	days := int(diff.Hours() / 24)

	if days >= 0 && days <= 30 {
		return 1
	}
	return 2
}

// FollowupNotes returns a full-text meaning followup notes according to the code book
func FollowupNotes(S1 string, fuNotes string, notes string,
	reason string, plat int, coag int, poNyha float64) string {
	var s, platText, coagText, nyhaText, status string
	// Status
	if S1 == "A" {
		status = "Alive"
	} else if S1 == "D" {
		status = "Died (late death)"
	} else if S1 == "N" {
		status = "Non-survivor of hospital stay (early death)"
	} else if S1 == "O" {
		status = "Other than usual follow-up methods required (see STATUS=O REASON)"
	} else if S1 == "R" {
		status = "Re-operation"
	} else {
		status = S1
	}

	// PLAT
	if plat == 0 {
		platText = "No"
	} else if plat == 1 {
		platText = "Yes"
	} else if plat == -9 {
		platText = "not applicable"
	} else {
		platText = strconv.Itoa(plat)
	}
	// COAG
	if coag == 0 {
		coagText = "No"
	} else if coag == 1 {
		coagText = "Yes"
	} else if coag == -9 {
		coagText = "not applicable"
	} else {
		coagText = strconv.Itoa(coag)
	}

	// PO_NYHA
	if poNyha == 1 {
		nyhaText = "No limitations"
	} else if poNyha == 2 {
		nyhaText = "Symptoms with extreme exertion or heavy physical activity"
	} else if poNyha == 3 {
		nyhaText = "Symptoms with light to moderate activity or with normal daily activity"
	} else if poNyha == 4 {
		nyhaText = "Symptoms at rest"
	} else if poNyha == -9 {
		nyhaText = "not applicable"
	} else {
		nyhaText = strconv.FormatFloat(poNyha, 'f', 1, 64)
	}

	if fuNotes == "" && notes == "" && reason == "" {
		s = "Status: '" + status + "'" + ", Plat: '" + platText + "'" +
			", COAG: '" + coagText + "'" + ", PO_NYHA: '" + nyhaText +
			"'"
	} else {
		s = "Status: '" + status + "'" + ", Plat: '" + platText + "'" +
			", COAG: '" + coagText + "'" + ", PO_NYHA: '" + nyhaText +
			"'" + ", Notes: '" + strings.Replace(strings.TrimSpace(fuNotes+" "+notes+" "+reason), " ", ", ", -1) + "'"
	}

	return s

}

// DeathNotes returns a full-text meaning of death notes according to the code book
func DeathNotes(prm string, reason string, operative string) string {

	var s, prmText, opText string
	// PRM_DTH
	if prm == "0" || prm == "" || prm == "9" || prm == "-9" {
		prmText = "Not applicable"
	} else if prm == "1" {
		prmText = "Valve-related cause"
	} else if prm == "2" {
		prmText = "Cardiac, non valve-related cause"
	} else if prm == "3" {
		prmText = "Non-cardiac cause"
	} else if prm == "4" {
		prmText = "Dissection (* Used only for David op FU, otherwise PRM_DTH=3)"
	} else {
		prmText = prm
	}
	//Operative
	if operative == "1" {
		opText = "Yes"
	} else {
		opText = "No"
	}

	s = ", here is the death info: Primary cause of death: '" +
		prmText + "', Reason of death: '" + reason + "', Operative: '" +
		opText + ", please indicate if death was operative'"

	return s

}

// TeNotes returns a full-text meaning TE notes according to the code book
func TeNotes(outcome string, anti string) string {
	var s, outText, antiText string
	// outcome
	if outcome == "0" || outcome == "" || outcome == "9" || outcome == "-9" {
		outText = "Not applicable"
	} else if outcome == "1" {
		outText = "Death"
	} else if outcome == "2" {
		outText = "Permanent deficit (symptoms lasting 3 weeks or longer)"
	} else if outcome == "3" {
		outText = "Transient deficit (symptoms lasting less than 3 weeks)"
	} else {
		outText = outcome
	}
	// anti_agents
	if anti == "0" {
		antiText = "No"
	} else if anti == "1" {
		antiText = "Yes, anticoagulants"
	} else if anti == "2" {
		antiText = "Yes, anti-platelet agents"
	} else if anti == "3" {
		antiText = "Yes, both"
	} else if anti == "" || anti == "9" || anti == "-9" || anti == "8" {
		antiText = "Not applicable"
	} else {
		antiText = anti
	}

	s = "outcome: '" + outText + "', agents: '" + antiText + "'"
	return s
}

// ArhCode returns a full-text meaning ARH codes according to the code book
func ArhCode(code string) string {
	var codeText, s string

	if code == "0" {
		codeText = "No"
	} else if code == "1" {
		codeText = "Yes, no treatment required."
	} else if code == "2" {
		codeText = "Yes, requiring hospitalization."
	} else if code == "3" {
		codeText = "Yes, requiring blood transfusion."
	} else if code == "4" {
		codeText = "= Yes, resulting in stroke."
	} else if code == "5" {
		codeText = "Yes, resulting in death "
	} else if code == "" || code == "9" || code == "-9" {
		codeText = "Not applicable"
	} else {
		codeText = code
	}

	s = "code: '" + codeText + "'"
	return s
}
