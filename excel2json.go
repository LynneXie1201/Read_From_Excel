package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

// global variables
var (
	allFollowUps []followUp
	allLKA       []lkaDate
	allDths      []death
	allReOper    []reOperation
	allTE        []te
	allSBE       []sbe
	events       []general //events including FUMI, FUPACE, SVD, PVL, DVT, ARH, THRM, HEML
	e            *log.Logger
	codes        []string
	nums         []int
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

// Helper functions

// Generate error messages to file and terminal
func newError(path string, id string, row int, t string, field string, invalid string) {
	errLog, err := os.OpenFile("errlog.txt", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Printf(err.Error())
	}
	multi := io.MultiWriter(errLog, os.Stdout)
	e = log.New(multi,
		"ERROR: ",
		log.Ldate|log.Ltime|log.Lshortfile)
	e.Println("File Path:", path, "PTID:", id, "Row #:", row+2, "Type:", t, "Info: Invalid", field, "Value:", invalid)
}

// Change the date Format to YYYY-MM-DD
func changeDateFormat(x string) string {
	value := strings.Replace(x, "\\", "", -1)
	test, _ := time.Parse("02-Jan-06", value)
	return test.Format("2006-01-02")
}

// a function that writes to json files
func writeTOFile(s interface{}, name string) {
	jsonFile, err := os.Create("./" + name + ".json")
	if err != nil {
		fmt.Println(err)
	}
	defer jsonFile.Close()

	p, _ := json.Marshal(s)
	jsonFile.Write(p)
	jsonFile.Close()
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
func checkFollowups(excelFileName string) (bool, []string) {
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
	// Check for the key fields
	if stringInSlice("FU_D", keys) && stringInSlice("Followup_STATUS", keys) {
		return true, keys
	}
	return false, nil
}

// Recursively for loop all files in the folder
func loopAllFiles(dirPath string) {
	fileList := []string{}
	err := filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if !f.IsDir() {
			fileList = append(fileList, path)
		}
		return nil
	})
	if err == nil {
		for _, file := range fileList {
			readExcelData(file)
		}
	}
}

// a function returns a slice of maps for one excel file
func excelToSlice(excelFileName string) []map[string]string {

	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf(err.Error())
	}

	value, keys := checkFollowups(excelFileName)
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
					m[keys[j]] = value
				}
				slices = append(slices, m)
			}
		}
		return slices[1:]
	}
	return nil
}

func readExcelData(path string) {
	// Returns a slice of maps from excel files
	s := excelToSlice(path)
	if s == nil {
		fmt.Println("no! this is not a follow_up file: ", path)
	} else {
		fmt.Println("yes! this is a follow_up file: ", path)
		for i, m := range s {
			// Event follow_up
			if m["FU_D"] != "" {
				fU := followUp{
					PTID:          m["PTID"],
					Date:          m["FU_D"],
					Type:          "followup",
					Status:        m["Followup_STATUS"],
					NoneValveReop: m["NONVALVE REOP"],
					FuNotes:       m["FU NOTES"],
					Notes:         m["NOTES"],
					LostOnDate:    m["STATUS=L DATE"],
					OtherNote:     m["STATUS=O REASON"]}
				fU.Coag, _ = strconv.Atoi(m["COAG"])
				fU.PoNYHA, _ = strconv.Atoi(m["PO_NYHA"])
				fU.Plat, _ = strconv.Atoi(m["PLAT"])
				// Validate fields' values
				if !stringInSlice(fU.Status, codes) {
					newError(path, fU.PTID, i, fU.Type, "Status", fU.Status)
				}
				if !intInSlice(fU.PoNYHA, nums[1:5]) {
					newError(path, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
				}
				if !intInSlice(fU.Coag, nums[:2]) {
					newError(path, fU.PTID, i, fU.Type, "COAG", m["COAG"])
				}
				if !intInSlice(fU.Plat, nums[0:2]) {
					newError(path, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
				}
				allFollowUps = append(allFollowUps, fU)
			}
			// Event LAST KNOWN ALIVE DATE
			if m["LKA_D"] != "" {
				l := lkaDate{
					PTID: m["PTID"],
					Type: "LKA_D",
					Date: m["LKA_D"]}
				allLKA = append(allLKA, l)
			}
			// Event Death
			if m["DTH_D"] != "" {
				d := death{
					PTID:   m["PTID"],
					Type:   "death",
					Date:   m["DTH_D"],
					Reason: m["REASDTH"]}
				d.PrmDth, _ = strconv.Atoi(m["PRM_DTH"])
				d.Code, _ = strconv.Atoi(m["DIED"])
				// Validate fields' values
				if !intInSlice(d.Code, nums[:2]) {
					newError(path, d.PTID, i, d.Type, "DIED", m["DIED"])
				}
				if !intInSlice(d.PrmDth, nums[:5]) {
					newError(path, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
				}
				allDths = append(allDths, d)
			}

			// Event FUREOP
			if m["FUREOP_D"] != "" {
				re := reOperation{
					PTID:    m["PTID"],
					Type:    "FUREOP",
					Date:    m["FUREOP_D"],
					Reason:  m["REASREOP"],
					Surgery: m["REOPSURG"],
					Notes:   m["REOPNOTES"]}
				re.Code, _ = strconv.Atoi(m["FUREOP"])
				re.Survival, _ = strconv.Atoi(m["REOPSURVIVAL"])
				// Validate fields' values
				if !intInSlice(re.Code, nums[:2]) {
					newError(path, re.PTID, i, re.Type, "FUREOP", m["FUREOP"])
				}
				allReOper = append(allReOper, re)
			}

			// Event TE
			if m["TE1_D"] != "" {
				te1 := te{
					PTID: m["PTID"],
					Type: "TE",
					Date: m["TE1_D"]}
				te1.Code, _ = strconv.Atoi(m["TE1"])
				te1.Outcome, _ = strconv.Atoi(m["TE1_OUT"])
				te1.Anti, _ = strconv.Atoi(m["ANTI_TE1"])
				if !intInSlice(te1.Code, nums[:4]) {
					newError(path, te1.PTID, i, te1.Type, "TE1", m["TE1"])
				}
				if !intInSlice(te1.Outcome, nums[:4]) {
					newError(path, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
				}
				if !intInSlice(te1.Anti, nums[:4]) && (te1.Anti != 8) {
					newError(path, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
				}
				allTE = append(allTE, te1)
			}
			if m["TE2_D"] != "" {
				te2 := te{
					PTID: m["PTID"],
					Type: "TE",
					Date: m["TE2_D"]}
				te2.Code, _ = strconv.Atoi(m["TE2"])
				te2.Outcome, _ = strconv.Atoi(m["TE2_OUT"])
				te2.Anti, _ = strconv.Atoi(m["ANTI_TE2"])
				if !intInSlice(te2.Code, nums[:4]) {
					newError(path, te2.PTID, i, te2.Type, "TE2", m["TE2"])
				}
				if !intInSlice(te2.Outcome, nums[:4]) {
					newError(path, te2.PTID, i, te2.Type, "TE2_OUT", m["TE2_OUT"])
				}
				if !intInSlice(te2.Anti, nums[:4]) && (te2.Anti != 8) {
					newError(path, te2.PTID, i, te2.Type, "ANTI_TE2", m["ANTI_TE2"])
				}
				allTE = append(allTE, te2)
			}
			if m["TE3_D"] != "" {
				te3 := te{
					PTID: m["PTID"],
					Type: "TE",
					Date: m["TE3_D"]}
				te3.Code, _ = strconv.Atoi(m["TE3"])
				te3.Outcome, _ = strconv.Atoi(m["TE3_OUT"])
				te3.Anti, _ = strconv.Atoi(m["ANTI_TE3"])
				if !intInSlice(te3.Code, nums[:4]) {
					newError(path, te3.PTID, i, te3.Type, "TE3", m["TE3"])
				}
				if !intInSlice(te3.Outcome, nums[:4]) {
					newError(path, te3.PTID, i, te3.Type, "TE3_OUT", m["TE3_OUT"])
				}
				if !intInSlice(te3.Anti, nums[:4]) && (te3.Anti != 8) {
					newError(path, te3.PTID, i, te3.Type, "ANTI_TE3", m["ANTI_TE3"])
				}
				allTE = append(allTE, te3)
			}

			// Event FUMI
			if m["FUMI_D"] != "" {
				f1 := general{
					PTID: m["PTID"],
					Type: "FUMI",
					Date: m["FUMI_D"]}
				f1.Code, _ = strconv.Atoi(m["FUMI"])
				events = append(events, f1)
			}

			// Event FUPACE
			if m["FUPACE_D"] != "" {
				f2 := general{
					PTID: m["PTID"],
					Type: "FUPACE",
					Date: m["FUPACE_D"]}
				f2.Code, _ = strconv.Atoi(m["FUPACE"])
				events = append(events, f2)
			}

			// Event SBE
			if m["SBE1_D"] != "" {
				sbe1 := sbe{
					PTID:     m["PTID"],
					Type:     "SBE",
					Date:     m["SBE1_D"],
					Organism: m["SBE1 ORGANISM"]}
				sbe1.Code, _ = strconv.Atoi(m["SBE1"])
				if !intInSlice(sbe1.Code, nums[:2]) {
					newError(path, sbe1.PTID, i, sbe1.Type, "SBE1", m["SBE1"])
				}
				allSBE = append(allSBE, sbe1)
			}

			if m["SBE2_D"] != "" {
				sbe2 := sbe{
					PTID:     m["PTID"],
					Type:     "SBE",
					Date:     m["SBE2_D"],
					Organism: m["SBE2 ORGANISM"]}
				sbe2.Code, _ = strconv.Atoi(m["SBE2"])
				if !intInSlice(sbe2.Code, nums[:2]) {
					newError(path, sbe2.PTID, i, sbe2.Type, "SBE2", m["SBE2"])
				}
				allSBE = append(allSBE, sbe2)
			}

			if m["SBE3_D"] != "" {
				sbe3 := sbe{
					PTID:     m["PTID"],
					Type:     "SBE",
					Date:     m["SBE3_D"],
					Organism: m["SBE3 ORGANISM"]}
				sbe3.Code, _ = strconv.Atoi(m["SBE3"])
				if !intInSlice(sbe3.Code, nums[:2]) {
					newError(path, sbe3.PTID, i, sbe3.Type, "SBE3", m["SBE3"])
				}
				allSBE = append(allSBE, sbe3)
			}

			// Event SVD
			if m["SVD_D"] != "" {
				s4 := general{
					PTID: m["PTID"],
					Type: "SVD",
					Date: m["SVD_D"]}
				s4.Code, _ = strconv.Atoi(m["SVD"])
				events = append(events, s4)
			}
			// Event PVL
			if m["PVL1_D"] != "" {
				pvl1 := general{
					PTID: m["PTID"],
					Type: "PVL",
					Date: m["PVL1_D"]}
				pvl1.Code, _ = strconv.Atoi(m["PVL1"])
				events = append(events, pvl1)
			}

			if m["PVL2_D"] != "" {
				pvl2 := general{
					PTID: m["PTID"],
					Type: "PVL",
					Date: m["PVL2_D"]}
				pvl2.Code, _ = strconv.Atoi(m["PVL2"])
				events = append(events, pvl2)
			}

			// Event DVT
			if m["DVT_D"] != "" {
				d1 := general{
					PTID: m["PTID"],
					Type: "DVT",
					Date: m["DVT_D"]}
				d1.Code, _ = strconv.Atoi(m["DVT"])
				events = append(events, d1)
			}
			// Event ARH
			if m["ARH1_D"] != "" {
				arh1 := general{
					PTID: m["PTID"],
					Type: "ARH",
					Date: m["ARH1_D"]}
				arh1.Code, _ = strconv.Atoi(m["ARH1"])
				if !intInSlice(arh1.Code, nums[:]) {
					newError(path, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
				}
				events = append(events, arh1)
			}

			if m["ARH2_D"] != "" {
				arh2 := general{
					PTID: m["PTID"],
					Type: "ARH",
					Date: m["ARH2_D"]}
				arh2.Code, _ = strconv.Atoi(m["ARH2"])
				if !intInSlice(arh2.Code, nums[:]) {
					newError(path, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
				}
				events = append(events, arh2)
			}

			// Event THRM
			if m["THRM1_D"] != "" {
				thrm1 := general{
					PTID: m["PTID"],
					Type: "THRM",
					Date: m["THRM1_D"]}
				thrm1.Code, _ = strconv.Atoi(m["THRM1"])
				events = append(events, thrm1)
			}

			if m["THRM2_D"] != "" {
				thrm2 := general{
					PTID: m["PTID"],
					Type: "THRM",
					Date: m["THRM2_D"]}
				thrm2.Code, _ = strconv.Atoi(m["THRM2"])
				events = append(events, thrm2)
			}

			// Event HEML
			if m["HEML1_D"] != "" {
				heml1 := general{
					PTID: m["PTID"],
					Type: "HEML",
					Date: m["HEML1_D"]}
				heml1.Code, _ = strconv.Atoi(m["HEML1"])
				events = append(events, heml1)
			}

			if m["HEML2_D"] != "" {
				heml2 := general{
					PTID: m["PTID"],
					Type: "HEML",
					Date: m["HEML2_D"]}
				heml2.Code, _ = strconv.Atoi(m["HEML2"])
				events = append(events, heml2)
			}
		}
	}
}

func main() {
	nums = []int{0, 1, 2, 3, 4, 5}
	codes = []string{"A", "D", "L", "N", "O", "R"}
	loopAllFiles("L:/CVDMC Students/Yilin Xie/data/excel")
	fmt.Println(allSBE)

}
