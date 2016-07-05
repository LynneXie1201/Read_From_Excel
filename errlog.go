package errlog

import "log"

// ErrorLog generates error messages to a file
func ErrorLog(e *log.Logger, path string, j int, id string, row int, t string, field string, invalid string) {

	e.Println(path, "Sheet#:", j, "PTID:", id, "Row #:", row+2, "Type:", t, "Info: Invalid", field, "Value:", invalid)

}

// PrintErr IS
func PrintErr(e *log.Logger, err error) {
	e.Println(err)

}

// Invalid is
func Invalid(e *log.Logger, indicator int, path string, j int) {

	if indicator == 0 {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID columns!")
	} else if indicator == 1 {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of Status columns!")
	} else if indicator == 2 {
		e.Println(path, "Sheet #:", j, "THIS SHEET DOES NOT HAVE HEADER ROW!")
	}

}

// Differ is
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
