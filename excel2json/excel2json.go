// Package excel2json provides functions that loop through excel files,
// read data from these files and create different events.
package excel2json

import (
	"excel/errlog"
	"excel/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

var (
	allFollowUps []followUp    // store followUp events
	allLKA       []lkaDate     // store last known alive date events
	allDths      []death       // store death events
	allReOper    []reOperation // store reoperation events
	allTE        []te          // store TE events
	allSBE       []sbe         // store SBE events
	events       []general     //store events including FUMI, FUPACE, SVD, PVL, DVT, ARH, THRM, HEML
	codes        []string      // status codes
	nums         []int         // numerical values for various codes

)

// follow up event
type followUp struct {
	PTID, Type, Date, Status, NoneValveReop, FuNotes, Notes, LostOnDate, OtherNote string
	Plat, Coag, PoNYHA                                                             int
}

// last-known-alive-date event
type lkaDate struct {
	PTID, Type, Date string
}

// death event
type death struct {
	PTID, Type, Date, Reason string
	Code, PrmDth             int
}

// re-operation event
type reOperation struct {
	PTID, Type, Date, Reason, Surgery, Notes string
	Code, Survival                           int
}

// TE event
type te struct {
	PTID, Type, Date    string
	Code, Outcome, Anti int
}

// SBE event
type sbe struct {
	PTID, Type, Date, Organism string
	Code                       int
}

// type of events that share the same variables
type general struct {
	PTID, Type, Date string
	Code             int
}

// Initialize before other functions get called
func init() {
	nums = []int{0, -9, 1, 2, 3, 4, 5}                 // list of numbers for validate codes
	codes = []string{"A", "D", "L", "N", "O", "R", ""} // correct codes for STATUS
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
	errlog.Invalid(e, 0, path, j)
	os.Exit(1) // exit if it has invaid columns of PTID
	return "", ""
}

// CheckStatusColumns checks the number of STATUS columns,
// and returns the column names of STATUS, assuming each file would have at most two STATUS columns.
// Parameters including:
// e - a logger that records the error messages;
// path - the path of the excel file;
// j - the index of the sheets in that excel file;
// keys - a slice that contains the header row
func CheckStatusColumns(e *log.Logger, path string, j int, keys []string) (string, string) {
	// create a slice that holds the column names of STATUS that matches a certain pattern
	status := []string{}
	for _, k := range keys {
		matched, err := regexp.MatchString("^.*STATUS$", k) // check status's pattern
		helper.CheckErr(e, err)
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
	errlog.Invalid(e, 1, path, j)
	os.Exit(1)
	return "", ""
}

// LoopAllFiles recursively loops all files in a folder, and tracks all excel files,
// opens a error log and a json file to store error messages and json objects, and
// for each excel file, calls another function to read data from the file.
func LoopAllFiles(e *log.Logger, dirPath string, jsonFile *os.File) {
	fileList := []string{}
	filepath.Walk(dirPath, func(path string, f os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
		} else if !f.IsDir() && strings.Contains(f.Name(), "xlsx") {
			fileList = append(fileList, path)
		}
		return nil
	})

	// Loop through all excel files
	for _, file := range fileList {
		ReadExcelData(e, file, jsonFile)
	}
}

// ReadExcelData uses the returned values of the function ExcelToSlice to
// build different types of events, and stores events to a json file.
func ReadExcelData(e *log.Logger, path string, jsonFile *os.File) {
	// slices is a slice of slices of maps, each map is a row in a excel file
	// keyList is a slice of slices of strings, each slice of strings is a header row of a excel file
	slices, keyList := helper.ExcelToSlice(e, path)
	// j is the index of sheets
	// s is a slice of maps representing the excel sheet of index j
	for j, s := range slices {
		if s == nil {
			// s is not a follow_up sheet
			fmt.Println("oops! this is not a follow_up sheet: ", path, "sheet #:", j)
		} else {
			// s is a follow_up excel sheet
			fmt.Println("Bingo! this is a follow_up sheet: ", path, "sheet #:", j)
			// keys is the header row of the excel sheet of index j
			keys := keyList[j]
			// check the number of PTID and STATUS' colomns
			p1, p2 := CheckPtidColumns(e, path, j, keys)
			st1, st2 := CheckStatusColumns(e, path, j, keys)
			// i is the index of rows
			// m is the map representing the correspnding row with the index i
			for i, m := range s {
				// check PTID
				ID1, ID2 := m[p1], m[p2]
				// have different PTID values
				if ID1 != ID2 {
					helper.AssignPTID(e, path, i, j, &ID1, &ID2)
				}
				// if format of PTID is not LLLFDDMMYY
				matched, err := regexp.MatchString("^.{4}(0?[1-9]|1[012])(0?[1-9]|[12][0-9]|3[01])[0-9]{2}$", ID1)
				helper.CheckErr(e, err)
				if !matched {
					errlog.Differ(e, 2, path, j, i, ID1, "")
				}
				// Check STATUS
				S1, S2 := m[st1], m[st2]
				// two different STATUS values
				if S1 != S2 {
					helper.AssignStatus(e, path, i, j, &S1, &S2)
				}
				// Event follow_up
				if m["FU_D"] != "" {
					fU := followUp{
						PTID:          ID1,
						Date:          helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"]),
						Type:          "followup",
						Status:        S1,
						NoneValveReop: m["NONVALVE REOP"],
						FuNotes:       m["FU NOTES"],
						Notes:         m["NOTES"],
						LostOnDate:    m["STATUS=L DATE"],
						OtherNote:     m["STATUS=O REASON"]}
					// check if these 3 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&fU.Coag, m["COAG"])
					helper.CheckEmpty(&fU.PoNYHA, m["PO_NYHA"])
					helper.CheckEmpty(&fU.Plat, m["PLAT"])

					// Validate fields' values
					if !helper.StringInSlice(fU.Status, codes) {
						errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "Status", fU.Status)
					}
					if !helper.IntInSlice(fU.PoNYHA, nums[1:6]) {
						errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
					}
					if !helper.IntInSlice(fU.Coag, nums[:3]) {
						errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
					}
					if !helper.IntInSlice(fU.Plat, nums[:3]) {
						errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, fU)
					allFollowUps = append(allFollowUps, fU)
				}
				// Event LAST KNOWN ALIVE DATE
				if m["LKA_D"] != "" {
					l := lkaDate{
						PTID: ID1,
						Type: "LKA_D",
						Date: helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, l)
					allLKA = append(allLKA, l)
				}
				// Event Death
				if m["DTH_D"] != "" {
					d := death{
						PTID:   ID1,
						Type:   "death",
						Date:   helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"]),
						Reason: m["REASDTH"]}
					// check if these two columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&d.PrmDth, m["PRM_DTH"])
					helper.CheckEmpty(&d.Code, m["DIED"])

					// Validate fields' values
					if !helper.IntInSlice(d.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "DIED", m["DIED"])
					}
					if !helper.IntInSlice(d.PrmDth, nums[:6]) {
						errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, d)
					allDths = append(allDths, d)
				}

				// Event FUREOP
				if m["FUREOP_D"] != "" {
					re := reOperation{
						PTID:    ID1,
						Type:    "FUREOP",
						Date:    helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"]),
						Reason:  m["REASREOP"],
						Surgery: m["REOPSURG"],
						Notes:   m["REOPNOTES"]}
					// check if these 2 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&re.Code, m["FUREOP"])
					helper.CheckEmpty(&re.Survival, m["REOPSURVIVAL"])

					// Validate fields' values
					if !helper.IntInSlice(re.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, re.PTID, i, re.Type, "FUREOP", m["FUREOP"])
					}
					if S1 == "R" && (re.Code != 1 || m["FUREOP_D"] == "") {
						errlog.Differ(e, 4, path, j, i, re.PTID, "")
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, re)
					allReOper = append(allReOper, re)
				}

				// Event TE
				if m["TE1_D"] != "" {
					te1 := te{
						PTID: ID1,
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE1_Date", m["TE1_D"])}
					// check if these 3 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&te1.Code, m["TE1"])
					helper.CheckEmpty(&te1.Outcome, m["TE1_OUT"])
					helper.CheckEmpty(&te1.Anti, m["ANTI_TE1"])

					// Validate fields' values
					if !helper.IntInSlice(te1.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1", m["TE1"])
					}
					if !helper.IntInSlice(te1.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
					}
					if !helper.IntInSlice(te1.Anti, nums[:5]) && (te1.Anti != 8) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, te1)
					allTE = append(allTE, te1)
				}
				if m["TE2_D"] != "" {
					te2 := te{
						PTID: ID1,
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])}
					// check if these 3 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&te2.Code, m["TE2"])
					helper.CheckEmpty(&te2.Outcome, m["TE2_OUT"])
					helper.CheckEmpty(&te2.Anti, m["ANTI_TE2"])

					// Validate fields' values
					if !helper.IntInSlice(te2.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "TE2", m["TE2"])
					}
					if !helper.IntInSlice(te2.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "TE2_OUT", m["TE2_OUT"])
					}
					if !helper.IntInSlice(te2.Anti, nums[:5]) && (te2.Anti != 8) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "ANTI_TE2", m["ANTI_TE2"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, te2)
					allTE = append(allTE, te2)
				}
				if m["TE3_D"] != "" {
					te3 := te{
						PTID: ID1,
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])}
					// check if these 3 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&te3.Code, m["TE3"])
					helper.CheckEmpty(&te3.Outcome, m["TE3_OUT"])
					helper.CheckEmpty(&te3.Anti, m["ANTI_TE3"])

					// Validate fields' values
					if !helper.IntInSlice(te3.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "TE3", m["TE3"])
					}
					if !helper.IntInSlice(te3.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "TE3_OUT", m["TE3_OUT"])
					}
					if !helper.IntInSlice(te3.Anti, nums[:5]) && (te3.Anti != 8) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "ANTI_TE3", m["ANTI_TE3"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, te3)
					allTE = append(allTE, te3)
				}

				// Event FUMI
				if m["FUMI_D"] != "" {
					f1 := general{
						PTID: ID1,
						Type: "FUMI",
						Date: helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&f1.Code, m["FUMI"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, f1)
					events = append(events, f1)
				}

				// Event FUPACE
				if m["FUPACE_D"] != "" {
					f2 := general{
						PTID: ID1,
						Type: "FUPACE",
						Date: helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&f2.Code, m["FUPACE"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, f2)
					events = append(events, f2)
				}

				// Event SBE
				if m["SBE1_D"] != "" {
					sbe1 := sbe{
						PTID:     ID1,
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE1_Date", m["SBE1_D"]),
						Organism: m["SBE1 ORGANISM"]}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&sbe1.Code, m["SBE1"])

					// Validate fields' values
					if !helper.IntInSlice(sbe1.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe1.PTID, i, sbe1.Type, "SBE1", m["SBE1"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, sbe1)
					allSBE = append(allSBE, sbe1)
				}

				if m["SBE2_D"] != "" {
					sbe2 := sbe{
						PTID:     ID1,
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE2_Date", m["SBE2_D"]),
						Organism: m["SBE2 ORGANISM"]}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&sbe2.Code, m["SBE2"])

					// Validate fields' values
					if !helper.IntInSlice(sbe2.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe2.PTID, i, sbe2.Type, "SBE2", m["SBE2"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, sbe2)
					allSBE = append(allSBE, sbe2)
				}

				if m["SBE3_D"] != "" {
					sbe3 := sbe{
						PTID:     ID1,
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE3_Date", m["SBE3_D"]),
						Organism: m["SBE3 ORGANISM"]}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&sbe3.Code, m["SBE3"])

					// Validate fields' values
					if !helper.IntInSlice(sbe3.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe3.PTID, i, sbe3.Type, "SBE3", m["SBE3"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, sbe3)
					allSBE = append(allSBE, sbe3)
				}

				// Event SVD
				if m["SVD_D"] != "" {
					s4 := general{
						PTID: ID1,
						Type: "SVD",
						Date: helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&s4.Code, m["SVD"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, s4)
					events = append(events, s4)
				}
				// Event PVL
				if m["PVL1_D"] != "" {
					pvl1 := general{
						PTID: ID1,
						Type: "PVL",
						Date: helper.CheckDateFormat(e, path, j, i, "PVL1_Date", m["PVL1_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&pvl1.Code, m["PVL1"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, pvl1)
					events = append(events, pvl1)
				}

				if m["PVL2_D"] != "" {
					pvl2 := general{
						PTID: ID1,
						Type: "PVL",
						Date: helper.CheckDateFormat(e, path, j, i, "PVL2_Date", m["PVL2_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&pvl2.Code, m["PVL2"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, pvl2)
					events = append(events, pvl2)
				}

				// Event DVT
				if m["DVT_D"] != "" {
					d1 := general{
						PTID: ID1,
						Type: "DVT",
						Date: helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&d1.Code, m["DVT"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, d1)
					events = append(events, d1)
				}
				// Event ARH
				if m["ARH1_D"] != "" {
					arh1 := general{
						PTID: ID1,
						Type: "ARH",
						Date: helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&arh1.Code, m["ARH1"])

					// Validate fields' values
					if !helper.IntInSlice(arh1.Code, nums[:]) {
						errlog.ErrorLog(e, path, j, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, arh1)
					events = append(events, arh1)
				}

				if m["ARH2_D"] != "" {
					arh2 := general{
						PTID: ID1,
						Type: "ARH",
						Date: helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&arh2.Code, m["ARH2"])

					// Validate fields' values
					if !helper.IntInSlice(arh2.Code, nums[:]) {
						errlog.ErrorLog(e, path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, arh2)
					events = append(events, arh2)
				}

				// Event THRM
				if m["THRM1_D"] != "" {
					thrm1 := general{
						PTID: ID1,
						Type: "THRM",
						Date: helper.CheckDateFormat(e, path, j, i, "THRM1_Date", m["THRM1_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&thrm1.Code, m["THRM1"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, thrm1)
					events = append(events, thrm1)
				}

				if m["THRM2_D"] != "" {
					thrm2 := general{
						PTID: ID1,
						Type: "THRM",
						Date: helper.CheckDateFormat(e, path, j, i, "THRM2_Date", m["THRM2_D"])}
					helper.CheckEmpty(&thrm2.Code, m["THRM2"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, thrm2)
					events = append(events, thrm2)
				}

				// Event HEML
				if m["HEML1_D"] != "" {
					heml1 := general{
						PTID: ID1,
						Type: "HEML",
						Date: helper.CheckDateFormat(e, path, j, i, "HEML1_Date", m["HEML1_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&heml1.Code, m["HEML1"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, heml1)
					events = append(events, heml1)
				}

				if m["HEML2_D"] != "" {
					heml2 := general{
						PTID: ID1,
						Type: "HEML",
						Date: helper.CheckDateFormat(e, path, j, i, "HEML2_Date", m["HEML2_D"])}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&heml2.Code, m["HEML2"])
					// write this object to the json file and store in a slice
					helper.WriteTOFile(jsonFile, heml2)
					events = append(events, heml2)
				}
			}
		}
	}
}
