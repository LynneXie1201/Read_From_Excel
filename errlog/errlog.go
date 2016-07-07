// Package errlog contains functions that write various styles of error logs
package errlog

import "log"

// ErrorLog generates error messages to a file.
// Parameters including:
// e - a logger that records the error messages;
// path - the path of the excel file;
// j - the index of the sheets in that excel file;
// id - the PTID of the record;
// row - the row number where the error occured;
// t - type of the event;
// field - the column name of the cell where error occured;
// invalid - invalid value.
func ErrorLog(e *log.Logger, path string, j int, id string, row int, t string, field string, invalid string) {

	e.Println(path, "Sheet#:", j, "PTID:", id, "Row #:", row+2, "Type:", t, "Info: Invalid", field, "Value:", invalid)

}

// PrintErr prints the error messages of err using the logger e.
func PrintErr(e *log.Logger, err error) {
	e.Println(err)
}

// Invalid generates different error messages using logger e.
// Parameters including:
// indicator - to decide which style of error messages get written;
// path - the path of the excel file;
// j - the index of the sheets in that excel file.
func Invalid(e *log.Logger, indicator int, path string, j int) {

	if indicator == 0 {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID columns!")
	} else if indicator == 1 {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of Status columns!")
	} else if indicator == 2 {
		e.Println(path, "Sheet #:", j, "THIS SHEET DOES NOT HAVE HEADER ROW!")
	}

}

// Differ generates different error messages using logger e.
// Parameters including:
// indicator - to decide which style of error messages get written;
// path - the path of the excel file;
// j - the index of the sheets in that excel file.
func Differ(e *log.Logger, indicator int, path string, j int, i int, s1 string, s2 string) {

	if indicator == 0 {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different status values: ", s1, s2)
	} else if indicator == 1 {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different PTID Values: ", s1, s2)
	} else if indicator == 2 {
		e.Println(path, "Sheet#", j, "PTID:", s1, "row #", i+2, "Invalid Format of PTID!")
	} else if indicator == 3 {
		e.Println(path, "Sheet#", j, "Row #: ", i+2, "INFO: Invaid PTID Value:", s1)
	} else if indicator == 4 {
		e.Println(path, "Sheet#", j, "PTID:", s1, "row #:", i+2, "INFO: iNCORRECT INFO OF REOPERATION!")
	} else if indicator == 5 {
		e.Println("Path:", path, "Sheet:", j, "Row#:", i+2, "Column:", s1, s2)
	}

}
