package excel2json

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"testing"
)

var (
	e     *log.Logger
	path  = "PATH"
	sheet = 1
	row   = 2

	keys []string
)

func init() {

	errLog, err := os.OpenFile("./excel2json_test.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer errLog.Close()
	e = log.New(errLog, "ERROR: ", 0)

}

// TestCheckPtidColumnsOne
func TestCheckPtidColumnsOne(t *testing.T) {
	t.Log("Test for no PTID column")
	keys = []string{"Name", "AddressPT"}
	if os.Getenv("BE_CRASHER") == "1" {
		CheckPtidColumns(e, path, sheet, keys)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckPtidColumnsOne")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestCheckPtidColumnsTwo(t *testing.T) {
	t.Log("Test for only one PTID column")
	keys = []string{"Name", "AddressPT", "PTID111"}
	id1, id2 := CheckPtidColumns(e, path, sheet, keys)
	if id1 != "PTID111" || id1 != id2 {
		t.Errorf("Expected PTID111, got %s", id1)
	}
}

func TestCheckPtidColumnsThree(t *testing.T) {
	t.Log("Test for two PTID columns")
	keys = []string{"Name", "AddressPT", "PTID111", "aaaaa", "PTID2"}
	id1, id2 := CheckPtidColumns(e, path, sheet, keys)
	if id1 != "PTID111" || id2 != "PTID2" {
		t.Errorf("Expected PTID111, got %s. Expected PTID2, got %s", id1, id2)
	}
}

// TestCheckPtidColumnsFour
func TestCheckPtidColumnsFour(t *testing.T) {
	t.Log("Test for more than 2 PTID columns")
	keys = []string{"Name", "AddressPT", "PTID1", "Second_PTID", "PTIDDDD"}
	if os.Getenv("BE_CRASHER") == "1" {
		CheckPtidColumns(e, path, sheet, keys)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckPtidColumnsFour")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}

// TestCheckStatusColumnsOne
func TestCheckStatusColumnsOne(t *testing.T) {
	t.Log("Test for no Status column")
	keys = []string{"Name", "AddressPT"}
	if os.Getenv("BE_CRASHER") == "1" {
		CheckStatusColumns(e, path, sheet, keys)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckStatusColumnsOne")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)

}

func TestCheckStatusColumnsTwo(t *testing.T) {
	t.Log("Test for only one Status column")
	keys = []string{"FU_STATUS", "Name", "AddressPT", "PTID111"}
	s1, s2 := CheckStatusColumns(e, path, sheet, keys)
	if s1 != "FU_STATUS" || s1 != s2 {
		t.Errorf("Expected FU_STATUS, got %s", s1)
	}
}

func TestCheckStatusColumnsThree(t *testing.T) {
	t.Log("Test for two Status columns")
	keys = []string{"Name", "STATUS_1", "follow_STATUS", "STATUS"}
	s1, s2 := CheckStatusColumns(e, path, sheet, keys)
	if s1 != "follow_STATUS" || s2 != "STATUS" {
		t.Errorf("Expected follow_STATUS, got %s. Expected STATUS, got %s", s1, s2)
	}
}

// TestCheckStatusColumnsFour
func TestCheckStatusColumnsFour(t *testing.T) {
	t.Log("Test for more than 2 Status columns")
	keys = []string{"Name", "AddressSTATUS", "PTID_STATUS", "Second_STATUS"}
	if os.Getenv("BE_CRASHER") == "1" {
		CheckStatusColumns(e, path, sheet, keys)
		return
	}
	cmd := exec.Command(os.Args[0], "-test.run=TestCheckStatusColumnsFour")
	cmd.Env = append(os.Environ(), "BE_CRASHER=1")
	err := cmd.Run()
	if e, ok := err.(*exec.ExitError); ok && !e.Success() {
		return
	}
	t.Fatalf("process ran with err %v, want exit status 1", err)
}
