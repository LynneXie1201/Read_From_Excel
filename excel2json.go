package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

// global variables to store different types of events
var (
	allFollowUps []followUp
	allLKA       []lkaDate
	allDths      []death
	allReOper    []reOperation
	allTE        []te
	allSBE       []sbe
	events       []general //events including FUMI, FUPACE, SVD, PVL, DVT, ARH, THRM, HEML
	e            *log.Logger
	jsonFile     *os.File
	codes        []string
	nums         []int
	id1          string
	id2          string
	s1           string
	s2           string
)

// Create types of events
type followUp struct {
	PTID, Type, Date, Status, NoneValveReop, FuNotes, Notes, LostOnDate, OtherNote string
	Plat, Coag, PoNYHA                                                             int
}

type lkaDate struct {
	PTID, Type, Date string
}

type death struct {
	PTID, Type, Date, Reason string
	Code, PrmDth             int
}

type reOperation struct {
	PTID, Type, Date, Reason, Surgery, Notes string
	Code, Survival                           int
}

type te struct {
	PTID, Type, Date    string
	Code, Outcome, Anti int
}

type sbe struct {
	PTID, Type, Date, Organism string
	Code                       int
}

type general struct {
	PTID, Type, Date string
	Code             int
}

// Generate error messages to a file
func errorLog(path string, j int, id string, row int, t string, field string, invalid string) {

	e.Println(path, "Sheet#:", j, "PTID:", id, "Row #:", row+2, "Type:", t, "Info: Invalid", field, "Value:", invalid)

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
func checkFollowups(sheet *xlsx.Sheet) (bool, []string) {
	//v, _ := sheet.Cell(0, 0).String()
	//fmt.Println(v)
	keys := []string{}
	for _, row := range sheet.Rows {
		for _, cell := range row.Cells {
			value, _ := cell.String()
			//fmt.Println(value)
			keys = append(keys, value)
		}
		break
	}
	if stringInSlice("FU_D", keys) && stringInSlice("DIED", keys) && stringInSlice("DTH_D", keys) {
		return true, keys
	}
	//fmt.Println(keys)
	return false, nil
}

// Recursively loop all excel files in a folder
func loopAllFiles(dirPath string) {
	fileList := []string{}
	err := filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() && strings.Contains(f.Name(), "xlsx") {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err == nil {
		// Open a file for error logs
		errLog, err := os.OpenFile("PATH_TO_ERRORLOG_FILE", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		checkErr(err) // check for errors
		defer errLog.Close()
		// Create a new logger
		e = log.New(errLog, "ERROR: ", 0)
		//Create a json file to store data from reading excel files
		jsonFile, err = os.OpenFile("PATH_TO_JSON_FILE", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		checkErr(err) // check for errors
		defer jsonFile.Close()
		// Loop through all excel files
		for _, file := range fileList {
			readExcelData(file)
		}
		// Close error log and json files
		errLog.Close()
		jsonFile.Close()
	}
}

func readExcelData(path string) {

	slices, keyList := excelToSlice(path) // s is a slice of maps
	for j, s := range slices {
		if s == nil {
			// s is not a follow_up excel file
			fmt.Println("oops! this is not a follow_up file: ", path)
		} else {
			// s is a follow_up excel file
			fmt.Println("Bingo! this is a follow_up file: ", path)
			keys := keyList[j]
			// check PTID
			checkPTID(path, j, keys)
			checkStatus(path, j, keys)
			for i, m := range s {
				// check PTID
				if m[id1] != m[id2] {
					assignNonEmptyPTID(path, i, j, m[id1], m[id2])
				}
				if len(m[id1]) != 10 {
					e.Println(path, "Sheet#", j, "PTID:", m[id1], "row #", i+2, "Invalid Format of PTID!")
				} else {
					matched, err := regexp.MatchString("(0?[1-9]|1[012])(0?[1-9]|[12][0-9]|3[01])", m[id1][4:8]) // validta MM AND dd OF A PTID
					checkErr(err)
					if !matched {
						e.Println(path, "Sheet#", j, "Row #: ", i+2, "INFO: Invaid PTID Value:", m[id1])
					}
				}
				// Check STATUS
				if m[s1] != m[s2] {
					assignNonEmptyStatus(path, i, j, m[s1], m[s2])
				}
				// Event follow_up
				if m["FU_D"] != "" {
					fU := followUp{
						PTID:          m[id1],
						Date:          m["FU_D"],
						Type:          "followup",
						Status:        m[s1],
						NoneValveReop: m["NONVALVE REOP"],
						FuNotes:       m["FU NOTES"],
						Notes:         m["NOTES"],
						LostOnDate:    m["STATUS=L DATE"],
						OtherNote:     m["STATUS=O REASON"]}
					// check if these 3 coloms empty or not
					if m["COAG"] != "" {
						fU.Coag, _ = strconv.Atoi(m["COAG"])
					} else {
						fU.Coag = -9
					}
					if m["PO_NYHA"] != "" {
						fU.PoNYHA, _ = strconv.Atoi(m["PO_NYHA"])
					} else {
						fU.PoNYHA = -9
					}
					if m["PLAT"] != "" {
						fU.Plat, _ = strconv.Atoi(m["PLAT"])
					} else {
						fU.Plat = -9
					}

					// Validate fields' values
					if !stringInSlice(fU.Status, codes) {
						errorLog(path, j, fU.PTID, i, fU.Type, "Status", fU.Status)
					}
					if !intInSlice(fU.PoNYHA, nums[1:6]) {
						errorLog(path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
					}
					if !intInSlice(fU.Coag, nums[:3]) {
						errorLog(path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
					}
					if !intInSlice(fU.Plat, nums[:3]) {
						errorLog(path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
					}
					writeTOFile(fU)                         // write this object to the json file
					allFollowUps = append(allFollowUps, fU) // also store data in a slice
				}
				// Event LAST KNOWN ALIVE DATE
				if m["LKA_D"] != "" {
					l := lkaDate{
						PTID: m[id1],
						Type: "LKA_D",
						Date: m["LKA_D"]}
					writeTOFile(l)
					allLKA = append(allLKA, l)
				}
				// Event Death
				if m["DTH_D"] != "" {
					d := death{
						PTID:   m[id1],
						Type:   "death",
						Date:   m["DTH_D"],
						Reason: m["REASDTH"]}
					if m["PRM_DTH"] != "" {
						d.PrmDth, _ = strconv.Atoi(m["PRM_DTH"])
					} else {
						d.PrmDth = -9
					}
					if m["DIED"] != "" {
						d.Code, _ = strconv.Atoi(m["DIED"])
					} else {
						d.Code = -9
					}

					// Validate fields' values
					if !intInSlice(d.Code, nums[:3]) {
						errorLog(path, j, d.PTID, i, d.Type, "DIED", m["DIED"])
					}
					if !intInSlice(d.PrmDth, nums[:6]) {
						errorLog(path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}
					writeTOFile(d)
					allDths = append(allDths, d)
				}

				// Event FUREOP
				if m["FUREOP_D"] != "" {
					re := reOperation{
						PTID:    m[id1],
						Type:    "FUREOP",
						Date:    m["FUREOP_D"],
						Reason:  m["REASREOP"],
						Surgery: m["REOPSURG"],
						Notes:   m["REOPNOTES"]}
					if m["FUREOP"] != "" {
						re.Code, _ = strconv.Atoi(m["FUREOP"])
					} else {
						re.Code = -9
					}
					if m["REOPSURVIVAL"] != "" {
						re.Survival, _ = strconv.Atoi(m["REOPSURVIVAL"])
					} else {
						re.Survival = -9
					}

					// Validate fields' values
					if !intInSlice(re.Code, nums[:3]) {
						errorLog(path, j, re.PTID, i, re.Type, "FUREOP", m["FUREOP"])
					}
					if m[s1] == "R" && (re.Code != 1 || m["FUREOP_D"] == "") {
						e.Println(path, "PTID: ", m[id1], "row #: ", i+2, "INFO: iNCORRECT INFO OF REOPERATION!")
					}
					writeTOFile(re)
					allReOper = append(allReOper, re)
				}

				// Event TE
				if m["TE1_D"] != "" {
					te1 := te{
						PTID: m[id1],
						Type: "TE",
						Date: m["TE1_D"]}
					if m["TE1"] != "" {
						te1.Code, _ = strconv.Atoi(m["TE1"])
					} else {
						te1.Code = -9
					}
					if m["TE1_OUT"] != "" {
						te1.Outcome, _ = strconv.Atoi(m["TE1_OUT"])
					} else {
						te1.Outcome = -9
					}
					if m["ANTI_TE1"] != "" {
						te1.Anti, _ = strconv.Atoi(m["ANTI_TE1"])
					} else {
						te1.Anti = -9
					}

					// Generate Error Messages
					if !intInSlice(te1.Code, nums[:5]) {
						errorLog(path, j, te1.PTID, i, te1.Type, "TE1", m["TE1"])
					}
					if !intInSlice(te1.Outcome, nums[:5]) {
						errorLog(path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
					}
					if !intInSlice(te1.Anti, nums[:5]) && (te1.Anti != 8) {
						errorLog(path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
					}
					writeTOFile(te1)
					allTE = append(allTE, te1)
				}
				if m["TE2_D"] != "" {
					te2 := te{
						PTID: m[id1],
						Type: "TE",
						Date: m["TE2_D"]}
					if m["TE2"] != "" {
						te2.Code, _ = strconv.Atoi(m["TE2"])
					} else {
						te2.Code = -9
					}
					if m["TE2_OUT"] != "" {
						te2.Outcome, _ = strconv.Atoi(m["TE2_OUT"])
					} else {
						te2.Outcome = -9
					}
					if m["ANTI_TE2"] != "" {
						te2.Anti, _ = strconv.Atoi(m["ANTI_TE2"])
					} else {
						te2.Anti = -9
					}

					// Generate Error Messages
					if !intInSlice(te2.Code, nums[:5]) {
						errorLog(path, j, te2.PTID, i, te2.Type, "TE2", m["TE2"])
					}
					if !intInSlice(te2.Outcome, nums[:5]) {
						errorLog(path, j, te2.PTID, i, te2.Type, "TE2_OUT", m["TE2_OUT"])
					}
					if !intInSlice(te2.Anti, nums[:5]) && (te2.Anti != 8) {
						errorLog(path, j, te2.PTID, i, te2.Type, "ANTI_TE2", m["ANTI_TE2"])
					}
					writeTOFile(te2)
					allTE = append(allTE, te2)
				}
				if m["TE3_D"] != "" {
					te3 := te{
						PTID: m[id1],
						Type: "TE",
						Date: m["TE3_D"]}
					if m["TE3"] != "" {
						te3.Code, _ = strconv.Atoi(m["TE3"])
					} else {
						te3.Code = -9
					}
					if m["TE3_OUT"] != "" {
						te3.Outcome, _ = strconv.Atoi(m["TE3_OUT"])
					} else {
						te3.Outcome = -9
					}
					if m["ANTI_TE3"] != "" {
						te3.Anti, _ = strconv.Atoi(m["ANTI_TE3"])
					} else {
						te3.Anti = -9
					}

					// Generate Error Messages
					if !intInSlice(te3.Code, nums[:5]) {
						errorLog(path, j, te3.PTID, i, te3.Type, "TE3", m["TE3"])
					}
					if !intInSlice(te3.Outcome, nums[:5]) {
						errorLog(path, j, te3.PTID, i, te3.Type, "TE3_OUT", m["TE3_OUT"])
					}
					if !intInSlice(te3.Anti, nums[:5]) && (te3.Anti != 8) {
						errorLog(path, j, te3.PTID, i, te3.Type, "ANTI_TE3", m["ANTI_TE3"])
					}
					writeTOFile(te3)
					allTE = append(allTE, te3)
				}

				// Event FUMI
				if m["FUMI_D"] != "" {
					f1 := general{
						PTID: m[id1],
						Type: "FUMI",
						Date: m["FUMI_D"]}
					if m["FUMI"] != "" {
						f1.Code, _ = strconv.Atoi(m["FUMI"])
					} else {
						f1.Code = -9
					}

					writeTOFile(f1)
					events = append(events, f1)
				}

				// Event FUPACE
				if m["FUPACE_D"] != "" {
					f2 := general{
						PTID: m[id1],
						Type: "FUPACE",
						Date: m["FUPACE_D"]}
					if m["FUPACE"] != "" {
						f2.Code, _ = strconv.Atoi(m["FUPACE"])
					} else {
						f2.Code = -9
					}

					writeTOFile(f2)
					events = append(events, f2)
				}

				// Event SBE
				if m["SBE1_D"] != "" {
					sbe1 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     m["SBE1_D"],
						Organism: m["SBE1 ORGANISM"]}
					if m["SBE1"] != "" {
						sbe1.Code, _ = strconv.Atoi(m["SBE1"])
					} else {
						sbe1.Code = -9
					}

					// Generate Error Messages
					if !intInSlice(sbe1.Code, nums[:3]) {
						errorLog(path, j, sbe1.PTID, i, sbe1.Type, "SBE1", m["SBE1"])
					}
					writeTOFile(sbe1)
					allSBE = append(allSBE, sbe1)
				}

				if m["SBE2_D"] != "" {
					sbe2 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     m["SBE2_D"],
						Organism: m["SBE2 ORGANISM"]}
					if m["SBE2"] != "" {
						sbe2.Code, _ = strconv.Atoi(m["SBE2"])
					} else {
						sbe2.Code = -9
					}

					// Generate Error Messages
					if !intInSlice(sbe2.Code, nums[:3]) {
						errorLog(path, j, sbe2.PTID, i, sbe2.Type, "SBE2", m["SBE2"])
					}
					writeTOFile(sbe2)
					allSBE = append(allSBE, sbe2)
				}

				if m["SBE3_D"] != "" {
					sbe3 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     m["SBE3_D"],
						Organism: m["SBE3 ORGANISM"]}
					if m["SBE3"] != "" {
						sbe3.Code, _ = strconv.Atoi(m["SBE3"])
					} else {
						sbe3.Code = -9
					}

					// Generate Error Messages
					if !intInSlice(sbe3.Code, nums[:3]) {
						errorLog(path, j, sbe3.PTID, i, sbe3.Type, "SBE3", m["SBE3"])
					}
					writeTOFile(sbe3)
					allSBE = append(allSBE, sbe3)
				}

				// Event SVD
				if m["SVD_D"] != "" {
					s4 := general{
						PTID: m[id1],
						Type: "SVD",
						Date: m["SVD_D"]}
					if m["SVD"] != "" {
						s4.Code, _ = strconv.Atoi(m["SVD"])
					} else {
						s4.Code = -9
					}

					writeTOFile(s4)
					events = append(events, s4)
				}
				// Event PVL
				if m["PVL1_D"] != "" {
					pvl1 := general{
						PTID: m[id1],
						Type: "PVL",
						Date: m["PVL1_D"]}
					if m["PVL1"] != "" {
						pvl1.Code, _ = strconv.Atoi(m["PVL1"])
					} else {
						pvl1.Code = -9
					}

					writeTOFile(pvl1)
					events = append(events, pvl1)
				}

				if m["PVL2_D"] != "" {
					pvl2 := general{
						PTID: m[id1],
						Type: "PVL",
						Date: m["PVL2_D"]}
					if m["PVL2"] != "" {
						pvl2.Code, _ = strconv.Atoi(m["PVL2"])
					} else {
						pvl2.Code = -9
					}

					writeTOFile(pvl2)
					events = append(events, pvl2)
				}

				// Event DVT
				if m["DVT_D"] != "" {
					d1 := general{
						PTID: m[id1],
						Type: "DVT",
						Date: m["DVT_D"]}
					if m["DVT"] != "" {
						d1.Code, _ = strconv.Atoi(m["DVT"])
					} else {
						d1.Code = -9
					}

					writeTOFile(d1)
					events = append(events, d1)
				}
				// Event ARH
				if m["ARH1_D"] != "" {
					arh1 := general{
						PTID: m[id1],
						Type: "ARH",
						Date: m["ARH1_D"]}
					if m["ARH1"] != "" {
						arh1.Code, _ = strconv.Atoi(m["ARH1"])
					} else {
						arh1.Code = -9
					}

					// Generate Error Messages
					if !intInSlice(arh1.Code, nums[:]) {
						errorLog(path, j, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
					}
					writeTOFile(arh1)
					events = append(events, arh1)
				}

				if m["ARH2_D"] != "" {
					arh2 := general{
						PTID: m[id1],
						Type: "ARH",
						Date: m["ARH2_D"]}
					if m["ARH2"] != "" {
						arh2.Code, _ = strconv.Atoi(m["ARH2"])
					} else {
						arh2.Code = -9
					}

					// Generate Error Messages
					if !intInSlice(arh2.Code, nums[:]) {
						errorLog(path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					writeTOFile(arh2)
					events = append(events, arh2)
				}

				// Event THRM
				if m["THRM1_D"] != "" {
					thrm1 := general{
						PTID: m[id1],
						Type: "THRM",
						Date: m["THRM1_D"]}
					if m["THRM1"] != "" {
						thrm1.Code, _ = strconv.Atoi(m["THRM1"])
					} else {
						thrm1.Code = -9
					}

					writeTOFile(thrm1)
					events = append(events, thrm1)
				}

				if m["THRM2_D"] != "" {
					thrm2 := general{
						PTID: m[id1],
						Type: "THRM",
						Date: m["THRM2_D"]}
					if m["THRM2"] != "" {
						thrm2.Code, _ = strconv.Atoi(m["THRM2"])
					} else {
						thrm2.Code = -9
					}

					writeTOFile(thrm2)
					events = append(events, thrm2)
				}

				// Event HEML
				if m["HEML1_D"] != "" {
					heml1 := general{
						PTID: m[id1],
						Type: "HEML",
						Date: m["HEML1_D"]}
					if m["HEML1"] != "" {
						heml1.Code, _ = strconv.Atoi(m["HEML1"])
					} else {
						heml1.Code = -9
					}

					writeTOFile(heml1)
					events = append(events, heml1)
				}

				if m["HEML2_D"] != "" {
					heml2 := general{
						PTID: m[id1],
						Type: "HEML",
						Date: m["HEML2_D"]}
					if m["HEML2"] != "" {
						heml2.Code, _ = strconv.Atoi(m["HEML2"])
					} else {
						heml2.Code = -9
					}

					writeTOFile(heml2)
					events = append(events, heml2)
				}
			}
		}
	}
}

// CHECK PTID
func checkPTID(path string, j int, keys []string) {
	id := []string{}
	for _, k := range keys {
		if strings.Contains(k, "PTID") {
			id = append(id, k)
		}
	}
	if len(id) == 2 {
		id1, id2 = id[0], id[1]
	} else if len(id) == 1 {
		id1, id2 = id[0], id[0]
	} else {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID!")
		fmt.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of PTID!")
		os.Exit(1) // exit if it has invaid columns of PTID
	}

}

// CHECK STATUS
func checkStatus(path string, j int, keys []string) {
	status := []string{}
	for _, k := range keys {
		matched, err := regexp.MatchString("^.*STATUS$", k) // check status's pattern
		checkErr(err)
		if matched {
			status = append(status, k)
		}
	}
	if len(status) == 2 {
		s1, s2 = status[0], status[1]
	} else if len(status) == 1 {
		s1, s2 = status[0], status[0]
	} else {
		e.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of STATUS!")
		fmt.Println(path, "Sheet #:", j, "INFO: This file has invalid numbers of STATUS!")
		os.Exit(1)
	}
}

func assignNonEmptyStatus(path string, i int, j int, s1 string, s2 string) {
	if s1 != "" && s2 != "" {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different status values: ", s1, s2)
	} else if s1 == "" {
		s1 = s2
	}

}

func assignNonEmptyPTID(path string, i int, j int, d1 string, d2 string) {
	if d1 != "" && d2 != "" {
		e.Println(path, "Sheet#", j, "Row #:", i+2, "INFO: Different PTID Values: ", d1, d2)
	} else if d1 == "" {
		d1 = d2
	}

}

// handling errors
func checkErr(err error) {
	if err != nil {
		e.Println(err)             // print to error log
		log.Fatalln("ERROR:", err) // print to terminal and then terminate
	}
}

// a function returns a slice of maps for one excel file
func excelToSlice(excelFilePath string) ([][]map[string]string, [][]string) {

	xlFile, err := xlsx.OpenFile(excelFilePath)
	//fmt.Println(xlFile.ToSlice())
	checkErr(err) // check for errors

	slices := [][]map[string]string{}
	keyList := [][]string{}
	for _, sheet := range xlFile.Sheets {

		value, keys := checkFollowups(sheet) // check for each sheet inside the excel file
		if value != false {
			keyList = append(keyList, keys)
			slice := []map[string]string{} // a sheet is a slice
			for _, row := range sheet.Rows {
				m := map[string]string{} // a row is a map
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
				slice = append(slice, m)
			}
			slices = append(slices, slice[1:])
		}
		if value == false {
			slices = append(slices, nil)
			keyList = append(keyList, nil)
		}
	}
	return slices, keyList
}

func main() {
	nums = []int{0, -9, 1, 2, 3, 4, 5}                 // list of numbers for validate codes
	codes = []string{"A", "D", "L", "N", "O", "R", ""} // correct codes for STATUS
	loopAllFiles("PATH_TO_YOUR_FOLDER")
}
