package helper

import (
	"fmt"
	"log"
	"os"
	"strings"
	"testing"
)

var (
	e      *log.Logger
	path   = "PATH"
	sheet  = 1
	row    = 2
	column = "DATE"
	date   string
)

func init() {
	errLog, err := os.OpenFile("./helper_test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer errLog.Close()
	e = log.New(errLog, "ERROR: ", 0)
}

// TestCheckDateFormatOne
func TestCheckDateFormatOne(t *testing.T) {
	t.Log("Test for date format YYYY-MM-DD")
	date = "1980-09-27"
	d := CheckDateFormat(e, path, sheet, row, column, date)
	matched := strings.Compare("1980-09-27", d)
	if matched != 0 {
		t.Error("Expected:", "1980-09-27", "got:", d)
	}
}

// TestCheckDateFormatTwo
func TestCheckDateFormatTwo(t *testing.T) {
	t.Log("Test for date format DD-MMM-YY")
	date = "27-Sep-80"
	d := CheckDateFormat(e, path, sheet, row, column, date)
	matched := strings.Compare("1980-09-27", d)
	if matched != 0 {
		t.Error("Expected:", "1980-09-27", "got:", d)
	}
}

// TestCheckDateFormatThree
func TestCheckDateFormatThree(t *testing.T) {
	t.Log("Test for date format MM-DD-YY")
	date = "09-27-80"
	d := CheckDateFormat(e, path, sheet, row, column, date)
	matched := strings.Compare("1980-09-27", d)
	if matched != 0 {
		t.Error("Expected:", "1980-09-27", "got:", d)
	}
}

// TestCheckDateFormatFour
func TestCheckDateFormatFour(t *testing.T) {
	t.Log("Test for date format YYYY/MM/DD")
	date = "1980/09/27@@@@@@@@\\\\\\\\;;;;;;;"
	d := CheckDateFormat(e, path, sheet, row, column, date)
	matched := strings.Compare("1980-09-27", d)
	if matched != 0 {
		t.Error("Expected:", "1980-09-27", "got:", d)
	}
}

// TestCheckDateFormatFive
func TestCheckDateFormatFive(t *testing.T) {
	t.Log("Test for invalid date format")
	date = "This is wrong!"
	d := CheckDateFormat(e, path, sheet, row, column, date)
	matched := strings.Compare("This is wrong!", d)
	if matched != 0 {
		t.Error("Expected:", "This is wrong!", "got:", d)
	}
}

// TestStringInSliceOne
func TestStringInSliceOne(t *testing.T) {
	t.Log("Test for StringInSlice")
	s := "a"
	slice := []string{"a", "b", "c"}
	bool := StringInSlice(s, slice)
	if !bool {
		t.Error("Something goes wrong: the slice should contain the string!")
	}
}

// TestStringInSliceTwo
func TestStringInSliceTwo(t *testing.T) {
	t.Log("Test for StringInSlice")
	s := "a"
	slice := []string{":", "b", "c"}
	bool := StringInSlice(s, slice)
	if bool {
		t.Error("Something goes wrong: the slice should not contain the string!")
	}
}

// TestIntInSliceOne
func TestIntInSliceOne(t *testing.T) {
	t.Log("Test for IntInSlice")
	s := 1
	slice := []int{1, 2, 4}
	bool := IntInSlice(s, slice)
	if !bool {
		t.Error("Something goes wrong: the slice should contain the int!")
	}
}

// TestIntInSliceTwo
func TestIntInSliceTwo(t *testing.T) {
	t.Log("Test for IntInSlice")
	s := 1
	slice := []int{2, 4}
	bool := IntInSlice(s, slice)
	if bool {
		t.Error("Something goes wrong: the slice should not contain the int!")
	}
}

// TestAssignStatusOne
func TestAssignStatusOne(t *testing.T) {
	t.Log("Test for AssignStatus - 2 different non empty values")
	s1, s2 := "aa", "bb"
	err := AssignStatus(e, path, row, sheet, &s1, &s2)
	if err == nil {
		t.Error("Something goes wrong here!")
	}

}

func TestAssignStatusTwo(t *testing.T) {
	t.Log("Test for AssignStatus - s1 is empty while s2 is not")
	s1, s2 := "", "bb"
	err := AssignStatus(e, path, row, sheet, &s1, &s2)
	if s1 != "bb" || err != nil {
		t.Error("Expected:", s2, "got:", s1)
	}

}

func TestAssignStatusThree(t *testing.T) {
	t.Log("Test for AssignStatus - s2 is empty while s1 is not")
	s2, s1 := "", "bb"
	err := AssignStatus(e, path, row, sheet, &s1, &s2)
	if s1 != "bb" || err != nil {
		t.Error("Expected:", s1, "got:", s1)
	}
}

func TestAssignPTIDOne(t *testing.T) {
	t.Log("Test for AssignStatus - 2 different non empty values")
	s1, s2 := "aa", "bb"
	err := AssignPTID(e, path, row, sheet, &s1, &s2)
	if err == nil {
		t.Error("Something goes wrong here!")
	}
}

func TestAssignPTIDTwo(t *testing.T) {
	t.Log("Test for AssignStatus - s1 is empty while s2 is not")
	s1, s2 := "", "bb"
	err := AssignPTID(e, path, row, sheet, &s1, &s2)
	if s1 != "bb" || err != nil {
		t.Error("Expected:", s2, "got:", s1)
	}

}

func TestAssignPTIDThree(t *testing.T) {
	t.Log("Test for AssignStatus - s2 is empty while s1 is not")
	s2, s1 := "", "bb"
	err := AssignPTID(e, path, row, sheet, &s1, &s2)
	if s1 != "bb" || err != nil {
		t.Error("Expected:", s1, "got:", s1)
	}
}

func TestCheckEmptyOne(t *testing.T) {
	t.Log("Test for CheckEmpty - value2 is empty, value1 should be -9.")
	value1, value2 := 1, ""
	CheckEmpty(&value1, value2)
	if value1 != -9 {
		t.Error("Expected:", -9, "got:", value1)
	}
}

func TestCheckEmptyTwo(t *testing.T) {
	t.Log("Test for CheckEmpty - value2 is not empty, value1 should be value2'.")
	value1, value2 := 1, "4"
	CheckEmpty(&value1, value2)
	if value1 != 4 {
		t.Error("Expected:", 4, "got:", value1)
	}

}
