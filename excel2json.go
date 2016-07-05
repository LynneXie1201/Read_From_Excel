package excel2json

import (
	"excel/errlog"
	"excel/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
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
	id1          string        // value of the first PTID column
	id2          string        // value of the second PTID column
	s1           string        // value of the first Status column
	s2           string        // value of the second Status column
	
)

// Create different ypes of events
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

// Initialize before the main function called
func init() {
	nums = []int{0, -9, 1, 2, 3, 4, 5}                 // list of numbers for validate codes
	codes = []string{"A", "D", "L", "N", "O", "R", ""} // correct codes for STATUS

}

// CheckPtidColumns checks the number of PTID columns,
// assume each file would have at most two PTID columns.
func CheckPtidColumns(e *log.Logger, path string, j int, keys []string) {
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
		errlog.Invalid(e, 0, path, j)
		os.Exit(1) // exit if it has invaid columns of PTID
	}

}

// CheckStatusColumns checks the number of STATUS columns,
// assume each file would have at most two STATUS columns.
func CheckStatusColumns(e *log.Logger, path string, j int, keys []string) {
	status := []string{}
	for _, k := range keys {
		matched, err := regexp.MatchString("^.*STATUS$", k) // check status's pattern
		helper.CheckErr(e, err)
		if matched {
			status = append(status, k)
		}
	}
	if len(status) == 2 {
		s1, s2 = status[0], status[1]
	} else if len(status) == 1 {
		s1, s2 = status[0], status[0]
	} else {
		errlog.Invalid(e, 1, path, j)
		os.Exit(1)
	}
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

	slices, keyList := helper.ExcelToSlice(e, path) // slices is a slice of slices of maps
	// j is the index of sheets
	for j, s := range slices {
		if s == nil {
			// s is not a follow_up sheet
			fmt.Println("oops! this is not a follow_up sheet: ", path, "sheet #:", j)
		} else {
			// s is a follow_up excel sheet
			fmt.Println("Bingo! this is a follow_up sheet: ", path, "sheet #:", j)
			keys := keyList[j]
			// check PTID and STATUS
			CheckPtidColumns(e, path, j, keys)
			CheckStatusColumns(e, path, j, keys)
			// i is the index of rows
			for i, m := range s {
				// check PTID
				if m[id1] != m[id2] {
					helper.AssignPTID(e, path, i, j, m[id1], m[id2])
				}
				if len(m[id1]) != 10 {
					errlog.Differ(e, 2, path, j, i, m[id1], "")
				} else {
					matched, err := regexp.MatchString("(0?[1-9]|1[012])(0?[1-9]|[12][0-9]|3[01])", m[id1][4:8]) // validate MM and DD part of a PTID
					helper.CheckErr(e, err)
					if !matched {
						errlog.Differ(e, 3, path, j, i, m[id1], "")
					}
				}
				// Check STATUS
				if m[s1] != m[s2] {
					helper.AssignStatus(e, path, i, j, m[s1], m[s2])
				}
				// Event follow_up
				if m["FU_D"] != "" {
					fU := followUp{
						PTID:          m[id1],
						Date:          helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"]),
						Type:          "followup",
						Status:        m[s1],
						NoneValveReop: m["NONVALVE REOP"],
						FuNotes:       m["FU NOTES"],
						Notes:         m["NOTES"],
						LostOnDate:    m["STATUS=L DATE"],
						OtherNote:     m["STATUS=O REASON"]}
					// check if these 3 columns are empty or not
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
					helper.WriteTOFile(jsonFile, fU)        // write this object to the json file
					allFollowUps = append(allFollowUps, fU) // also store data in a slice
				}
				// Event LAST KNOWN ALIVE DATE
				if m["LKA_D"] != "" {
					l := lkaDate{
						PTID: m[id1],
						Type: "LKA_D",
						Date: helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])}
					helper.WriteTOFile(jsonFile, l)
					allLKA = append(allLKA, l)
				}
				// Event Death
				if m["DTH_D"] != "" {
					d := death{
						PTID:   m[id1],
						Type:   "death",
						Date:   helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"]),
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
					if !helper.IntInSlice(d.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "DIED", m["DIED"])
					}
					if !helper.IntInSlice(d.PrmDth, nums[:6]) {
						errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}
					helper.WriteTOFile(jsonFile, d)
					allDths = append(allDths, d)
				}

				// Event FUREOP
				if m["FUREOP_D"] != "" {
					re := reOperation{
						PTID:    m[id1],
						Type:    "FUREOP",
						Date:    helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"]),
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
					if !helper.IntInSlice(re.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, re.PTID, i, re.Type, "FUREOP", m["FUREOP"])
					}
					if m[s1] == "R" && (re.Code != 1 || m["FUREOP_D"] == "") {
						errlog.Differ(e, 4, path, j, i, m[id1], "")
					}
					helper.WriteTOFile(jsonFile, re)
					allReOper = append(allReOper, re)
				}

				// Event TE
				if m["TE1_D"] != "" {
					te1 := te{
						PTID: m[id1],
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE1_Date", m["TE1_D"])}
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
					if !helper.IntInSlice(te1.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1", m["TE1"])
					}
					if !helper.IntInSlice(te1.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
					}
					if !helper.IntInSlice(te1.Anti, nums[:5]) && (te1.Anti != 8) {
						errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
					}
					helper.WriteTOFile(jsonFile, te1)
					allTE = append(allTE, te1)
				}
				if m["TE2_D"] != "" {
					te2 := te{
						PTID: m[id1],
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])}
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
					if !helper.IntInSlice(te2.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "TE2", m["TE2"])
					}
					if !helper.IntInSlice(te2.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "TE2_OUT", m["TE2_OUT"])
					}
					if !helper.IntInSlice(te2.Anti, nums[:5]) && (te2.Anti != 8) {
						errlog.ErrorLog(e, path, j, te2.PTID, i, te2.Type, "ANTI_TE2", m["ANTI_TE2"])
					}
					helper.WriteTOFile(jsonFile, te2)
					allTE = append(allTE, te2)
				}
				if m["TE3_D"] != "" {
					te3 := te{
						PTID: m[id1],
						Type: "TE",
						Date: helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])}
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
					if !helper.IntInSlice(te3.Code, nums[:5]) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "TE3", m["TE3"])
					}
					if !helper.IntInSlice(te3.Outcome, nums[:5]) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "TE3_OUT", m["TE3_OUT"])
					}
					if !helper.IntInSlice(te3.Anti, nums[:5]) && (te3.Anti != 8) {
						errlog.ErrorLog(e, path, j, te3.PTID, i, te3.Type, "ANTI_TE3", m["ANTI_TE3"])
					}
					helper.WriteTOFile(jsonFile, te3)
					allTE = append(allTE, te3)
				}

				// Event FUMI
				if m["FUMI_D"] != "" {
					f1 := general{
						PTID: m[id1],
						Type: "FUMI",
						Date: helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])}
					if m["FUMI"] != "" {
						f1.Code, _ = strconv.Atoi(m["FUMI"])
					} else {
						f1.Code = -9
					}

					helper.WriteTOFile(jsonFile, f1)
					events = append(events, f1)
				}

				// Event FUPACE
				if m["FUPACE_D"] != "" {
					f2 := general{
						PTID: m[id1],
						Type: "FUPACE",
						Date: helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])}
					if m["FUPACE"] != "" {
						f2.Code, _ = strconv.Atoi(m["FUPACE"])
					} else {
						f2.Code = -9
					}

					helper.WriteTOFile(jsonFile, f2)
					events = append(events, f2)
				}

				// Event SBE
				if m["SBE1_D"] != "" {
					sbe1 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE1_Date", m["SBE1_D"]),
						Organism: m["SBE1 ORGANISM"]}
					if m["SBE1"] != "" {
						sbe1.Code, _ = strconv.Atoi(m["SBE1"])
					} else {
						sbe1.Code = -9
					}

					// Generate Error Messages
					if !helper.IntInSlice(sbe1.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe1.PTID, i, sbe1.Type, "SBE1", m["SBE1"])
					}
					helper.WriteTOFile(jsonFile, sbe1)
					allSBE = append(allSBE, sbe1)
				}

				if m["SBE2_D"] != "" {
					sbe2 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE2_Date", m["SBE2_D"]),
						Organism: m["SBE2 ORGANISM"]}
					if m["SBE2"] != "" {
						sbe2.Code, _ = strconv.Atoi(m["SBE2"])
					} else {
						sbe2.Code = -9
					}

					// Generate Error Messages
					if !helper.IntInSlice(sbe2.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe2.PTID, i, sbe2.Type, "SBE2", m["SBE2"])
					}
					helper.WriteTOFile(jsonFile, sbe2)
					allSBE = append(allSBE, sbe2)
				}

				if m["SBE3_D"] != "" {
					sbe3 := sbe{
						PTID:     m[id1],
						Type:     "SBE",
						Date:     helper.CheckDateFormat(e, path, j, i, "SBE3_Date", m["SBE3_D"]),
						Organism: m["SBE3 ORGANISM"]}
					if m["SBE3"] != "" {
						sbe3.Code, _ = strconv.Atoi(m["SBE3"])
					} else {
						sbe3.Code = -9
					}

					// Generate Error Messages
					if !helper.IntInSlice(sbe3.Code, nums[:3]) {
						errlog.ErrorLog(e, path, j, sbe3.PTID, i, sbe3.Type, "SBE3", m["SBE3"])
					}
					helper.WriteTOFile(jsonFile, sbe3)
					allSBE = append(allSBE, sbe3)
				}

				// Event SVD
				if m["SVD_D"] != "" {
					s4 := general{
						PTID: m[id1],
						Type: "SVD",
						Date: helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])}
					if m["SVD"] != "" {
						s4.Code, _ = strconv.Atoi(m["SVD"])
					} else {
						s4.Code = -9
					}

					helper.WriteTOFile(jsonFile, s4)
					events = append(events, s4)
				}
				// Event PVL
				if m["PVL1_D"] != "" {
					pvl1 := general{
						PTID: m[id1],
						Type: "PVL",
						Date: helper.CheckDateFormat(e, path, j, i, "PVL1_Date", m["PVL1_D"])}
					if m["PVL1"] != "" {
						pvl1.Code, _ = strconv.Atoi(m["PVL1"])
					} else {
						pvl1.Code = -9
					}

					helper.WriteTOFile(jsonFile, pvl1)
					events = append(events, pvl1)
				}

				if m["PVL2_D"] != "" {
					pvl2 := general{
						PTID: m[id1],
						Type: "PVL",
						Date: helper.CheckDateFormat(e, path, j, i, "PVL2_Date", m["PVL2_D"])}
					if m["PVL2"] != "" {
						pvl2.Code, _ = strconv.Atoi(m["PVL2"])
					} else {
						pvl2.Code = -9
					}

					helper.WriteTOFile(jsonFile, pvl2)
					events = append(events, pvl2)
				}

				// Event DVT
				if m["DVT_D"] != "" {
					d1 := general{
						PTID: m[id1],
						Type: "DVT",
						Date: helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])}
					if m["DVT"] != "" {
						d1.Code, _ = strconv.Atoi(m["DVT"])
					} else {
						d1.Code = -9
					}

					helper.WriteTOFile(jsonFile, d1)
					events = append(events, d1)
				}
				// Event ARH
				if m["ARH1_D"] != "" {
					arh1 := general{
						PTID: m[id1],
						Type: "ARH",
						Date: helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])}
					if m["ARH1"] != "" {
						arh1.Code, _ = strconv.Atoi(m["ARH1"])
					} else {
						arh1.Code = -9
					}

					// Generate Error Messages
					if !helper.IntInSlice(arh1.Code, nums[:]) {
						errlog.ErrorLog(e, path, j, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
					}
					helper.WriteTOFile(jsonFile, arh1)
					events = append(events, arh1)
				}

				if m["ARH2_D"] != "" {
					arh2 := general{
						PTID: m[id1],
						Type: "ARH",
						Date: helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])}
					if m["ARH2"] != "" {
						arh2.Code, _ = strconv.Atoi(m["ARH2"])
					} else {
						arh2.Code = -9
					}

					// Generate Error Messages
					if !helper.IntInSlice(arh2.Code, nums[:]) {
						errlog.ErrorLog(e, path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					helper.WriteTOFile(jsonFile, arh2)
					events = append(events, arh2)
				}

				// Event THRM
				if m["THRM1_D"] != "" {
					thrm1 := general{
						PTID: m[id1],
						Type: "THRM",
						Date: helper.CheckDateFormat(e, path, j, i, "THRM1_Date", m["THRM1_D"])}
					if m["THRM1"] != "" {
						thrm1.Code, _ = strconv.Atoi(m["THRM1"])
					} else {
						thrm1.Code = -9
					}

					helper.WriteTOFile(jsonFile, thrm1)
					events = append(events, thrm1)
				}

				if m["THRM2_D"] != "" {
					thrm2 := general{
						PTID: m[id1],
						Type: "THRM",
						Date: helper.CheckDateFormat(e, path, j, i, "THRM2_Date", m["THRM2_D"])}
					if m["THRM2"] != "" {
						thrm2.Code, _ = strconv.Atoi(m["THRM2"])
					} else {
						thrm2.Code = -9
					}

					helper.WriteTOFile(jsonFile, thrm2)
					events = append(events, thrm2)
				}

				// Event HEML
				if m["HEML1_D"] != "" {
					heml1 := general{
						PTID: m[id1],
						Type: "HEML",
						Date: helper.CheckDateFormat(e, path, j, i, "HEML1_Date", m["HEML1_D"])}
					if m["HEML1"] != "" {
						heml1.Code, _ = strconv.Atoi(m["HEML1"])
					} else {
						heml1.Code = -9
					}

					helper.WriteTOFile(jsonFile, heml1)
					events = append(events, heml1)
				}

				if m["HEML2_D"] != "" {
					heml2 := general{
						PTID: m[id1],
						Type: "HEML",
						Date: helper.CheckDateFormat(e, path, j, i, "HEML2_Date", m["HEML2_D"])}
					if m["HEML2"] != "" {
						heml2.Code, _ = strconv.Atoi(m["HEML2"])
					} else {
						heml2.Code = -9
					}

					helper.WriteTOFile(jsonFile, heml2)
					events = append(events, heml2)
				}
			}
		}
	}
}
