// Package excel2json provides functions that loop through excel files,
// read data from these files and create different events.
package excel2json

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

var (
	allFollowUps []followUp // store followUp events
	allDths      []death    // store death events
	allTIA       []tia      // store TIA events
	allStroke    []stroke   // store stroke events
	allSBE       []sbe      // store SBE events
	allARH       []arh
	allOperation []operation
	allFUMI      []general
	allFUPACE    []general
	allSVD       []general
	allPVL       []general
	allDVT       []general
	allTHRM      []general
	alllHEML     []general
	allLAK       []general
	allFix       []general
	codes        []string // status codes
	nums         []int    // numerical values for various codes
	floats       []float64
)

type source struct {
	Type string   `json:"type"`
	Path []string `json:"path"`
}

type operation struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PeriopID   *string
	PTID, Date string
	DateEst    int `json:"date_est"`
	Surgeon    string
	Surgeries  []string
	Children   []string
	Parent     *int
	Notes      string
	Source     source
	Fix        []errMessage
}

type errMessage struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

// FollowUp is follow up event
type followUp struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int `json:"date_est"`
	Status     *string
	Notes      string
	LostOnDate string
	Unusual    string `json:"unusual"`
	Plat, Coag int
	PoNYHA     float64
	Source     source
	Fix        []errMessage
}

// death event
type death struct {
	Type              string
	MRN               string `json:"mrn"`
	ResearchID        string `json:"research_id"`
	PTID, Date        string
	DateEst           int `json:"date_est"`
	Reason            string
	PrmDth, Operative int
	Source            source
	Fix               []errMessage
}

// stroke event
type stroke struct {
	Type                  string
	MRN                   string `json:"mrn"`
	ResearchID            string `json:"research_id"`
	PTID, Date            string
	DateEst               int `json:"date_est"`
	Outcome, Agents, When int
	Source                source
	Fix                   []errMessage
}

// TIA event
type tia struct {
	Type            string
	MRN             string `json:"mrn"`
	ResearchID      string `json:"research_id"`
	PTID, Date      string
	DateEst         int `json:"date_est"`
	Outcome, Agents int
	Source          source
	Fix             []errMessage
}

// SBE event
type sbe struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int `json:"date_est"`
	Organism   *string
	Source     source
	Fix        []errMessage
}

// type of events that share the same variables
type general struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int    `json:"date_est"`
	Notes      string `json:"omitempty"` // some events don't have notes field
	Source     source
	Fix        []errMessage `json:"omitempty"` // Fix events don't need fix field
}

// type of events that share the same variables
type arh struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int `json:"date_est"`
	Code       int
	Source     source
	Fix        []errMessage
}

// Initialize before other functions get called
func init() {
	nums = []int{0, -9, 1, 2, 3, 4, 5}                 // list of numbers for validate codes
	codes = []string{"D", "L", "N", "O", "A", "R", ""} // correct codes for STATUS
	floats = []float64{0, -9, 1, 2, 3, 4, 5, 1.5, 2.5, 3.5, 4.5}
}

// CompareFollowUps checks if two follow up events are duplicate
func (a followUp) CompareFollowUps(s []followUp) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Coag == b.Coag && a.Date == b.Date &&
			a.LostOnDate == b.LostOnDate && a.Notes == b.Notes && a.Unusual == b.Unusual &&
			a.PTID == b.PTID && a.Plat == b.Plat && a.PoNYHA == b.PoNYHA && a.Status == b.Status {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true

		}
	}
	return false
}

func earlyDeathInfo(a death) string {
	var prmDth string
	var str string
	if a.PrmDth == 0 {
		prmDth = "Not applicable"
		if a.Operative == 0 {
			if a.DateEst == 0 {
				str = "another record had a different date: " + a.Date + "non-operative, primary death reason: " + prmDth
			}
			str = "another record had a different date: " + a.Date + ", date_estimated, non-operative, primary death reason: " + prmDth
		} else if a.DateEst == 1 {
			str = "another record had a different date: " + a.Date + ", date_estimated, operative, primary death reason: " + prmDth
		}
		str = "another record had a different date: " + a.Date + ", operative, primary death reason: " + prmDth
	} else if a.PrmDth == 1 {
		prmDth = "Valve-related cause"
		if a.Operative == 0 {
			if a.DateEst == 0 {
				str = "another record had a different date: " + a.Date + "non-operative, primary death reason: " + prmDth
			}
			str = "another record had a different date: " + a.Date + ", date_estimated, non-operative, primary death reason: " + prmDth
		} else if a.DateEst == 1 {
			str = "another record had a different date: " + a.Date + ", date_estimated, operative, primary death reason: " + prmDth
		}
		str = "another record had a different date: " + a.Date + ", operative, primary death reason: " + prmDth
	} else if a.PrmDth == 2 {
		prmDth = "Cardiac, non valve-related cause"
		if a.Operative == 0 {
			if a.DateEst == 0 {
				str = "another record had a different date: " + a.Date + "non-operative, primary death reason: " + prmDth
			}
			str = "another record had a different date: " + a.Date + ", date_estimated, non-operative, primary death reason: " + prmDth
		} else if a.DateEst == 1 {
			str = "another record had a different date: " + a.Date + ", date_estimated, operative, primary death reason: " + prmDth
		}
		str = "another record had a different date: " + a.Date + ", operative, primary death reason: " + prmDth
	} else if a.PrmDth == 3 {
		prmDth = "Non-cardiac cause"
		if a.Operative == 0 {
			if a.DateEst == 0 {
				str = "another record had a different date: " + a.Date + "non-operative, primary death reason: " + prmDth
			}
			str = "another record had a different date: " + a.Date + ", date_estimated, non-operative, primary death reason: " + prmDth
		} else if a.DateEst == 1 {
			str = "another record had a different date: " + a.Date + ", date_estimated, operative, primary death reason: " + prmDth
		}
		str = "another record had a different date: " + a.Date + ", operative, primary death reason: " + prmDth
	} else if a.PrmDth == 4 {
		prmDth = "Dissection"
		if a.Operative == 0 {
			if a.DateEst == 0 {
				str = "another record had a different date: " + a.Date + "non-operative, primary death reason: " + prmDth
			}
			str = "another record had a different date: " + a.Date + ", date_estimated, non-operative, primary death reason: " + prmDth
		} else if a.DateEst == 1 {
			str = "another record had a different date: " + a.Date + ", date_estimated, operative, primary death reason: " + prmDth
		}
		str = "another record had a different date: " + a.Date + ", operative, primary death reason: " + prmDth
	}
	return str
}

// CompareDeath checks if two death events are duplicate
func (a death) CompareDeath(s []death) bool {
	// i is the index of b
	for i, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Date == b.Date && a.PTID == b.PTID && a.Operative == b.Operative &&
			a.PrmDth == b.PrmDth && a.Reason == b.Reason {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
			// same person with different death date
		} else if a.Date != b.Date && a.PTID == b.PTID && a.MRN == b.MRN && a.ResearchID == b.ResearchID {
			// how to compare 2 dates?
			if helper.DateLaterThan(b.Date, a.Date) {
				earlyDeath := earlyDeathInfo(a)
				for _, e := range b.Fix {
					if e.Field == "date" {
						e.Msg += "; " + earlyDeath
					}
				}
				msg := errMessage{"date", earlyDeath}
				b.Fix = append(b.Fix, msg)
				return true
			} else if helper.DateLaterThan(a.Date, b.Date) {

				earlyDeath := earlyDeathInfo(b) // info of b
				for _, e := range b.Fix {
					if e.Field == "date" {
						earlyDeath += "; " + e.Msg
					}
				}
				msg := errMessage{"date", earlyDeath}
				a.Fix = append(a.Fix, msg)
				// try to delete b from the slice
				s = append(s[:i], s[i+1:]...)
				return false
			}

		}

	}
	return false
}

// CompareTia checks if two TIA events are duplicate
func (a tia) CompareTia(s []tia) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Agents == b.Agents && a.Date == b.Date &&
			a.Outcome == b.Outcome && a.PTID == b.PTID {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

// CompareStroke checks if two stroke events are duplicate
func (a stroke) CompareStroke(s []stroke) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Agents == b.Agents && a.Date == b.Date && a.When == b.When &&
			a.Outcome == b.Outcome && a.PTID == b.PTID {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

// CompareSbe checks if two sbe events are duplicate
func (a sbe) CompareSbe(s []sbe) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Date == b.Date && a.Organism == b.Organism &&
			a.PTID == b.PTID {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

// CompareEvents checks if two events (including FUMI, FUPACE, SVD, PVL, DVT,
// ARH, THRM, HEML) are duplicate
func (a general) CompareEvents(s []general) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Date == b.Date && a.PTID == b.PTID &&
			a.Notes == b.Notes {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

func (a arh) CompareARH(s []arh) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Code == b.Code && a.Date == b.Date &&
			a.PTID == b.PTID {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

func (a operation) CompareOperation(s []operation) bool {
	for _, b := range s {
		if &a == &b {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		} else if a.Date == b.Date && a.PTID == b.PTID &&
			a.Notes == b.Notes && a.Surgeon == b.Surgeon && reflect.DeepEqual(a.Fix, b.Fix) {
			b.Source.Path = append(b.Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
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
	columnsChecker := helper.GetUserInput()
	// Loop through all excel files
	for _, file := range fileList {
		ReadExcelData(e, file, jsonFile, columnsChecker)
	}

	WriteToJSON(jsonFile, allARH, allDVT, allDths, allFUMI, allFUPACE, allFix, allFollowUps,
		allLAK, allOperation, allPVL, allSBE, allSVD,
		allStroke, allTHRM, allTIA, alllHEML)
}

// ReadExcelData uses the returned values of the function ExcelToSlice to
// build different types of events, and stores events to a json file.
func ReadExcelData(e *log.Logger, path string, jsonFile *os.File, columnsChecker string) {
	// slices is a slice of slices of maps, each map is a row in a excel file
	// keyList is a slice of slices of strings, each slice of strings is a header row of a excel file
	slices, keyList := helper.ExcelToSlice(e, path, columnsChecker)
	// get the sub path of the original path
	path = helper.SubPath(path, "valve_registry")
	// j is the index of sheets
	// s is a slice of maps representing the excel sheet of index j
	for j, s := range slices {
		if s == nil {
			// s is not a follow_up sheet
			fmt.Println("oops! this is not a follow_up sheet: ", path, "sheet #:", j+1)
		} else {

			// s is a follow_up excel sheet
			fmt.Println("Bingo! this is a follow_up sheet: ", path, "sheet #:", j+1)
			// keys is the header row of the excel sheet of index j
			keys := keyList[j]
			// check the number of PTID and STATUS' colomns
			p1, p2 := helper.CheckPtidColumns(e, path, j, keys)
			st1, st2 := helper.CheckStatusColumns(e, path, j, keys)
			// i is the index of rows
			// m is the map representing the correspnding row with the index i
			for i, m := range s {
				// check PTID
				ID1, ID2 := m[p1], m[p2]
				// assign PTIDs

				diffID := helper.AssignPTID(&ID1, &ID2)
				// Check STATUS
				S1, S2 := m[st1], m[st2]
				// two different STATUS values

				diffStatus := helper.AssignStatus(&S1, &S2)
				// check if format of PTID is LLLFDDMMYY
				helper.CheckPtidFormat(ID1, e, path, j, i)

				// Event follow_up
				date, est := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])
				fuNotes := "Status:" + S1 + ", Fu Notes: " + m["FU NOTES"] +
					", Notes: " + m["NOTES"] + ", LostOnDate: " + m["STATUS=L DATE"] + ", OtherNote:" +
					m["STATUS=O REASON"] + ", Plat: " + m["PLAT"] + ", COAG: " + m["COAG"] + ", PO_NYHA: " + m["PO_NYHA"]

				if est == 0 || est == 1 {
					fu := followUp{
						PTID:       ID1,
						Date:       date,
						Type:       "followup",
						Status:     &S1,
						Notes:      m["NOTES"] + "; " + m["FU NOTES"],
						LostOnDate: m["STATUS=L DATE"],
						Unusual:    m["STATUS=O REASON"],
						DateEst:    est,
						Fix:        []errMessage{},
						Source:     source{Type: "followup", Path: []string{}}}
					// Source: add path
					fu.Source.Path = append(fu.Source.Path, path)
					// check PTID
					if diffID {
						msg := errMessage{"PTID", "two different PTIDs:" + ID1 + ", " + ID2}
						fu.Fix = append(fu.Fix, msg)
					}

					// check STATUS
					if diffStatus {
						msg := errMessage{"Status", "two different Statuses:" + S1 + ", " + S2}
						fu.Fix = append(fu.Fix, msg)
						if !helper.StringInSlice(1, S1, codes[:4]) {
							fu.Status = &S2
						}
					}

					// Validate fields' values
					if !helper.StringInSlice(1, *fu.Status, codes) {
						msg := errMessage{"code", "invalid value:" + *fu.Status}
						fu.Fix = append(fu.Fix, msg)
					}
					if !helper.CheckFloatValue(&fu.PoNYHA, m["PO_NYHA"], floats[1:]) {
						msg := errMessage{"PO_NYHA", "invalid value:" + m["PO_NYHA"]}
						fu.Fix = append(fu.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
					}
					if !helper.CheckIntValue(&fu.Coag, m["COAG"], nums[:3]) {
						msg := errMessage{"COAG", "invalid value:" + m["COAG"]}
						fu.Fix = append(fu.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
					}
					if !helper.CheckIntValue(&fu.Plat, m["PLAT"], nums[:3]) {
						msg := errMessage{"PLAT", "invalid value:" + m["PLAT"]}
						fu.Fix = append(fu.Fix, msg)
						//	errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !fu.CompareFollowUps(allFollowUps) {
						allFollowUps = append(allFollowUps, fu)
						helper.WriteTOFile(jsonFile, fu)
					}
					// est == 3 means that date has invalid format

				} else if est == 3 {

					// create a new fix event
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)
					f.Notes = "Invalid date: " + date + fuNotes
					//helper.WriteTOFile(jsonFile, f)
					allFix = append(allFix, f)

					// follow up date is empty, but has fu notes, then create lka events
				} else if est == 2 {
					var coag, plat int
					var poNYHA float64
					// Check int and float values
					helper.CheckIntValue(&coag, m["COAG"], nums[:3])
					helper.CheckFloatValue(&poNYHA, m["PO_NYHA"], floats[1:])
					helper.CheckIntValue(&plat, m["PLAT"], nums[:3])

					if m["FU NOTES"] != "" || coag != -9 || plat != -9 || poNYHA != -9 ||
						m["STATUS=O REASON"] != "" || m["NOTES"] != "" {
						lkaDate, lkaEst := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])
						// create LKA_D Event
						if lkaEst == 0 || lkaEst == 1 {
							l := general{
								PTID:    ID1,
								Type:    "LKA_D",
								Date:    lkaDate,
								Notes:   fuNotes,
								DateEst: lkaEst,
								Fix:     []errMessage{},
								Source:  source{Type: "followup", Path: []string{}}}
							// Source: add path
							l.Source.Path = append(l.Source.Path, path)

							// if no duplicates, write this object to the json file and store in a slice
							if !l.CompareEvents(allLAK) {
								allLAK = append(allLAK, l)
							}

						} //else if LKA_D is empty or invalid format, create Fix event
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						// LKA_D IS empty
						if lkaEst == 2 {
							f.Notes = "followup and LKA without date associated: here are my followup notes:" + fuNotes
						}
						// LKA date with invalid format
						f.Notes = "LKA Date:" + date + " ,FU NOTES without date associated: here are my followup notes:" + fuNotes

						// Source: add path
						f.Source.Path = append(f.Source.Path, path)

						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}

					}

				}

				// Event Death
				date, est = helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"])
				if est == 0 || est == 1 {
					d := death{
						PTID:    ID1,
						Type:    "death",
						Date:    date,
						Reason:  m["REASDTH"],
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					d.Source.Path = append(d.Source.Path, path)

					// check Operative
					operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
					if m["SURVIVAL"] == "0" {
						d.Operative = 1
						if operDate != date {
							msg := errMessage{"operative", "please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					} else if m["SURVIVAL"] == "1" {
						if operDate == date {
							msg := errMessage{"operative", "please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					}

					// Validate fields' values
					if !helper.CheckIntValue(&d.PrmDth, m["PRM_DTH"], nums[:6]) {
						msg := errMessage{"PRM_DTH", "invalid value:" + m["PRM_DTH"]}
						d.Fix = append(d.Fix, msg)
						//errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}

					if S1 != "D" {
						msg := errMessage{"status", "invalid value:" + S1}
						d.Fix = append(d.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !d.CompareDeath(allDths) {
						helper.WriteTOFile(jsonFile, d)
						allDths = append(allDths, d)
					}
				} else if est == 2 || est == 3 {

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// est == 3 means invalid date format
					if est == 3 {
						f.Notes = "Death event with invalid date format, here is the death info: death date: " +
							date + ", Primary cause of death: " + m["PRM_DTH"] + ", Reason of death:" + m["REASDTH"]
						// else est == 2
					} else if m["PRM_DTH"] != "" || m["REASDTH"] != "" || m["DIED"] == "1" {
						f.Notes = "Death event with no date associated, here is the death info: code: " + m["DIED"] +
							"Primary cause of death: " + m["PRM_DTH"] + ", Reason of death:" + m["REASDTH"]
					}

					// Source: add path
					f.Source.Path = append(f.Source.Path, path)
					//helper.WriteTOFile(jsonFile, f)
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}

				}

				// Event FUREOP -> Event operation

				date, est = helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"])
				if est == 0 || est == 1 {

					op := operation{
						PTID:      ID1,
						Type:      "operation",
						Date:      date,
						DateEst:   est,
						Surgeries: []string{},
						Children:  []string{},
						Fix:       []errMessage{},
						Source:    source{Type: "followup", Path: []string{}}}
					// Source: add path
					op.Source.Path = append(op.Source.Path, path)

					// Validate fields' values
					if S1 != "R" || m["FUREOP"] != "1" {
						msg := errMessage{"status", "status is not R or code is not 1," + "status:" + S1 + "code:" + m["FUREOP"]}
						op.Fix = append(op.Fix, msg)
						//errlog.Differ(e, 4, path, j, i, re.PTID, "INCORRECT INFO OF REOPERATION!")
					}

					var survival int
					if !helper.CheckIntValue(&survival, m["REOPSURVIVAL"], nums[:3]) {
						msg := errMessage{"survival", "invalid value: " + m["REOPSURVIVAL"]}
						op.Fix = append(op.Fix, msg)
					}

					// add re-op strirng

					if !(m["REASREOP"] == "" && m["REOPSURVIVAL"] == "0" && m["REOPNOTES"] == "" &&
						m["REOPSURG"] == "" && m["NONVALVE REOP"] == "") {
						string := helper.OperationString(m["REASREOP"], m["REOPSURVIVAL"], m["REOPNOTES"], m["REOPSURG"], m["NONVALVE REOP"])
						msg := errMessage{"operation", string}
						op.Fix = append(op.Fix, msg)
					}
					// check if these 2 columns are empty or not,
					// if empty, assign -9
					//helper.CheckEmpty(&op.Survival, m["REOPSURVIVAL"])

					// if no duplicates, write this object to the json file and store in a slice
					if !op.CompareOperation(allOperation) {
						helper.WriteTOFile(jsonFile, op)
						allOperation = append(allOperation, op)
					}

				} else if est == 2 || est == 3 {
					opString := helper.OperationString(m["REASREOP"], m["REOPSURVIVAL"], m["REOPNOTES"], m["REOPSURG"], m["NONVALVE REOP"])
					// create a event fix
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					if est == 3 {
						f.Notes = "Invalid REOP date format: " + m["FUREOP_D"] + ", " + opString
					} else if m["FUREOP"] == "1" || m["REASREOP"] != "" || m["REOPNOTES"] != "" || m["REOPSURG"] != "" || m["NONVALVE REOP"] != "" {
						f.Notes = "REOP fields without date associated: " + opString
					}

					// Source: add path
					f.Source.Path = append(f.Source.Path, path)

					//helper.WriteTOFile(jsonFile, f)
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// TE1
				date, est = helper.CheckDateFormat(e, path, j, i, "TE1_Date", m["TE1_D"])
				if est == 0 || est == 1 {

					if m["TE1"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE1"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field When
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE1_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE1"]}
							s.Fix = append(s.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareStroke(allStroke) {
							helper.WriteTOFile(jsonFile, s)
							allStroke = append(allStroke, s)
						}
					} else if m["TE1"] == "3" {
						// Event TIA
						t := tia{
							PTID:    ID1,
							Type:    "TIA",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE1_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE1"]}
							t.Fix = append(t.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTia(allTIA) {
							helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 || est == 3 {

					if m["TE1"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)
					// stroke
					if m["TE1"] == "2" {
						if est == 3 {
							f.Notes = "stroke with invalid date format: " + date + ", outcome:" + m["TE1_OUT"] + ", agents: " + m["ANTI_TE1"]
						} else if est == 2 {
							f.Notes = "stroke with no date but code exists, , outcome:" + m["TE1_OUT"] + ", agents: " + m["ANTI_TE1"]
						}
						// TIA
					} else if m["TE1"] == "3" {

						if est == 3 {
							f.Notes = "TIA with invalid date format: " + date + ", outcome:" + m["TE1_OUT"] + ", agents: " + m["ANTI_TE1"]
						} else if est == 2 {
							f.Notes = "TIA with no date but code exists, , outcome:" + m["TE1_OUT"] + ", agents: " + m["ANTI_TE1"]
						}

					}

					if !f.CompareEvents(allFix) {

						allFix = append(allFix, f)
					}

				}
				// TE2
				date, est = helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])

				if est == 0 || est == 1 {

					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE2"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field when
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE2_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE2"]}
							s.Fix = append(s.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareStroke(allStroke) {
							helper.WriteTOFile(jsonFile, s)
							allStroke = append(allStroke, s)
						}
					} else if m["TE2"] == "3" {
						// Event TIA
						t := tia{
							PTID:    ID1,
							Type:    "TIA",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE2_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE2"]}
							t.Fix = append(t.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTia(allTIA) {
							helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 || est == 3 {
					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)
					// stroke
					if m["TE2"] == "2" {
						if est == 3 {
							f.Notes = "stroke with invalid date format: " + date + ", outcome:" + m["TE2_OUT"] + ", agents: " + m["ANTI_TE2"]
						} else if est == 2 {
							f.Notes = "stroke with no date but code exists, , outcome:" + m["TE2_OUT"] + ", agents: " + m["ANTI_TE2"]
						}
						// TIA
					} else if m["TE2"] == "3" {

						if est == 3 {
							f.Notes = "TIA with invalid date format: " + date + ", outcome:" + m["TE2_OUT"] + ", agents: " + m["ANTI_TE2"]
						} else if est == 2 {
							f.Notes = "TIA with no date but code exists, , outcome:" + m["TE2_OUT"] + ", agents: " + m["ANTI_TE2"]
						}

					}

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}

				}

				// TE3
				date, est = helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])
				if est == 0 || est == 1 {

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE3"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field when
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE3_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE3"]}
							s.Fix = append(s.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareStroke(allStroke) {
							helper.WriteTOFile(jsonFile, s)
							allStroke = append(allStroke, s)
						}
					} else if m["TE3"] == "3" {
						// Event TIA
						t := tia{
							PTID:    ID1,
							Type:    "TIA",
							Date:    date,
							DateEst: est,
							Fix:     []errMessage{},
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE3_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE3"]}
							t.Fix = append(t.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTia(allTIA) {
							helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 || est == 3 {

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)
					// stroke
					if m["TE3"] == "2" {
						if est == 3 {
							f.Notes = "stroke with invalid date format: " + date + ", outcome:" + m["TE3_OUT"] + ", agents: " + m["ANTI_TE3"]
						} else if est == 2 {
							f.Notes = "stroke with no date but code exists, , outcome:" + m["TE3_OUT"] + ", agents: " + m["ANTI_TE3"]
						}
						// TIA
					} else if m["TE3"] == "3" {

						if est == 3 {
							f.Notes = "TIA with invalid date format: " + date + ", outcome:" + m["TE3_OUT"] + ", agents: " + m["ANTI_TE3"]
						} else if est == 2 {
							f.Notes = "TIA with no date but code exists, , outcome:" + m["TE3_OUT"] + ", agents: " + m["ANTI_TE3"]
						}

					}

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}

				}

				// Event FUMI
				date, est = helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])
				if est == 0 || est == 1 {
					f1 := general{
						PTID:    ID1,
						Type:    "FUMI",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					f1.Source.Path = append(f1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice

					//helper.WriteTOFile(jsonFile, f1)
					if !f1.CompareEvents(allFUMI) {
						allFUMI = append(allFUMI, f1)
					}

				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "FUMI with invalid date format: " + date + ", code: " + m["FUMI"]
					} else if m["FUMI"] == "1" {
						f.Notes = "FUMI with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event FUPACE
				date, est = helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])
				if est == 0 || est == 1 {
					f2 := general{
						PTID:    ID1,
						Type:    "FUPACE",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					f2.Source.Path = append(f2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !f2.CompareEvents(allFUPACE) {
						//helper.WriteTOFile(jsonFile, f2)
						allFUPACE = append(allFUPACE, f2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "FUPACE with invalid date format: " + date + ", code: " + m["FUPACE"]
					} else if m["FUPACE"] == "1" {
						f.Notes = "FUPACE with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event SBE
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE1_Date", m["SBE1_D"])
				ORGANISM := m["SBE1 ORGANISM"]
				organism := m["SBE1 organism"]
				if est == 0 || est == 1 {

					sbe1 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					sbe1.Source.Path = append(sbe1.Source.Path, path)

					// some sheets may have organism instead of ORGANISM
					if ORGANISM != "" {
						sbe1.Organism = &ORGANISM
					} else {
						sbe1.Organism = &organism
					}

					// 	Check Organism
					if !helper.CheckStringValue(sbe1.Organism) {
						msg := errMessage{"organism", "invalid organism value: " + *sbe1.Organism}
						sbe1.Fix = append(sbe1.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe1.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe1)
						allSBE = append(allSBE, sbe1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if ORGANISM != "" {
						organism = ORGANISM
					}

					if est == 3 {
						f.Notes = "SBE with invalid date format: " + date + ", code: " + m["SBE1"] + "Organism: " + organism
					} else if m["SBE1"] == "1" {
						f.Notes = "SBE with no date but code is 1, code: " + m["SBE1"] + "Organism: " + organism
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// SBE2
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE2_Date", m["SBE2_D"])
				ORGANISM = m["SBE2 ORGANISM"]
				organism = m["SBE2 organism"]
				if est == 0 || est == 1 {

					sbe2 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					sbe2.Source.Path = append(sbe2.Source.Path, path)

					// some sheets may have organism instead of ORGANISM
					if ORGANISM != "" {
						sbe2.Organism = &ORGANISM
					} else {
						sbe2.Organism = &organism
					}

					// 	Check Organism
					if !helper.CheckStringValue(sbe2.Organism) {
						msg := errMessage{"organism", "invalid organism value: " + *sbe2.Organism}
						sbe2.Fix = append(sbe2.Fix, msg)
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !sbe2.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe2)
						allSBE = append(allSBE, sbe2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)

					if ORGANISM != "" {
						organism = ORGANISM
					}

					if est == 3 {

						f.Notes = "SBE with invalid date format: " + date + ", code: " + m["SBE2"] + "Organism: " + organism
					} else if m["SBE2"] == "1" {
						f.Notes = "SBE with no date but code is 1, code: " + m["SBE2"] + "Organism: " + organism
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// SBE3
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE3_Date", m["SBE3_D"])
				ORGANISM = m["SBE3 ORGANISM"]
				organism = m["SBE3 organism"]
				if est == 0 || est == 1 {
					sbe3 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					sbe3.Source.Path = append(sbe3.Source.Path, path)

					// some sheets may have organism instead of ORGANISM
					if ORGANISM != "" {
						sbe3.Organism = &ORGANISM
					} else {
						sbe3.Organism = &organism
					}

					// 	Check Organism
					if !helper.CheckStringValue(sbe3.Organism) {
						msg := errMessage{"organism", "invalid organism value: " + *sbe3.Organism}
						sbe3.Fix = append(sbe3.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe3.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe3)
						allSBE = append(allSBE, sbe3)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Source.Path = append(f.Source.Path, path)
					if ORGANISM != "" {
						organism = ORGANISM
					}

					if est == 3 {
						f.Notes = "SBE with invalid date format: " + date + ", code: " + m["SBE3"] + "Organism: " + organism
					} else if m["SBE3"] == "1" {
						f.Notes = "SBE with no date but code is 1, code: " + m["SBE3"] + "Organism: " + organism
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event SVD
				date, est = helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])
				if est == 0 || est == 1 {
					s4 := general{
						PTID:    ID1,
						Type:    "SVD",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					s4.Source.Path = append(s4.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !s4.CompareEvents(allSVD) {
						//helper.WriteTOFile(jsonFile, s4)
						allSVD = append(allSVD, s4)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "PVL with invalid date format: " + date + ", code: " + m["SVD"]
					} else if m["SVD"] == "1" {
						f.Notes = "SVD with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event PVL
				date, est = helper.CheckDateFormat(e, path, j, i, "PVL1_Date", m["PVL1_D"])
				if est == 0 || est == 1 {
					pvl1 := general{
						PTID:    ID1,
						Type:    "PVL",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					pvl1.Source.Path = append(pvl1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl1.CompareEvents(allPVL) {
						//helper.WriteTOFile(jsonFile, pvl1)
						allPVL = append(allPVL, pvl1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "PVL with invalid date format: " + date + ", code: " + m["PVL1"]
					} else if m["PVL1"] == "1" {
						f.Notes = "PVL with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// PVL2
				date, est = helper.CheckDateFormat(e, path, j, i, "PVL2_Date", m["PVL2_D"])
				if est == 0 || est == 1 {
					pvl2 := general{
						PTID:    ID1,
						Type:    "PVL",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					pvl2.Source.Path = append(pvl2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl2.CompareEvents(allPVL) {
						//helper.WriteTOFile(jsonFile, pvl2)
						allPVL = append(allPVL, pvl2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "PVL with invalid date format: " + date + ", code: " + m["PVL2"]
					} else if m["PVL2"] == "1" {
						f.Notes = "PVL with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event DVT
				date, est = helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])
				if est == 0 || est == 1 {

					d1 := general{
						PTID:    ID1,
						Type:    "DVT",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					d1.Source.Path = append(d1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !d1.CompareEvents(allDVT) {
						//helper.WriteTOFile(jsonFile, d1)
						allDVT = append(allDVT, d1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "DVT with invalid date format: " + date + ", code: " + m["DVT"]
					} else if m["DVT"] == "1" {
						f.Notes = "DVT with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event ARH
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])
				if est == 0 || est == 1 {
					arh1 := arh{
						PTID:    ID1,
						Type:    "ARH",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					arh1.Source.Path = append(arh1.Source.Path, path)

					// Validate fields' values
					if !helper.CheckIntValue(&arh1.Code, m["ARH1"], nums[:]) {
						msg := errMessage{"code", "invalid value:" + m["ARH1"]}
						arh1.Fix = append(arh1.Fix, msg)
						//errlog.ErrorLog(e, path, j, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh1.CompareARH(allARH) {
						//helper.WriteTOFile(jsonFile, arh1)
						allARH = append(allARH, arh1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "ARH with invalid date format: " + date + ", code: " + m["ARH1"]
					} else if m["ARH1"] != "0" && m["ARH1"] != "" {
						f.Notes = "ARH with no date but code is not 0 or empty: code: " + m["ARH1"]
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// ARH2
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])
				if est == 0 || est == 1 {
					arh2 := arh{
						PTID:    ID1,
						Type:    "ARH",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					arh2.Source.Path = append(arh2.Source.Path, path)

					// Validate fields' values
					if !helper.CheckIntValue(&arh2.Code, m["ARH2"], nums[:]) {
						msg := errMessage{"code", "invalid value:" + m["ARH2"]}
						arh2.Fix = append(arh2.Fix, msg)
						//errlog.ErrorLog(e, path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh2.CompareARH(allARH) {
						//helper.WriteTOFile(jsonFile, arh2)
						allARH = append(allARH, arh2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "ARH with invalid date format: " + date + ", code: " + m["ARH2"]
					} else if m["ARH2"] != "0" && m["ARH2"] != "" {
						f.Notes = "ARH with no date but code is not 0 or empty: code: " + m["ARH2"]
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}
				// Event THRM
				date, est = helper.CheckDateFormat(e, path, j, i, "THRM1_Date", m["THRM1_D"])
				if est == 0 || est == 1 {

					thrm1 := general{
						PTID:    ID1,
						Type:    "THRM",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					thrm1.Source.Path = append(thrm1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !thrm1.CompareEvents(allTHRM) {
						//helper.WriteTOFile(jsonFile, thrm1)
						allTHRM = append(allTHRM, thrm1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "THRM with invalid date format: " + date + ", code: " + m["THRM1"]
					} else if m["THRM1"] == "1" {
						f.Notes = "THRM with empty date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// THRM2
				date, est = helper.CheckDateFormat(e, path, j, i, "THRM2_Date", m["THRM2_D"])
				if est == 0 || est == 1 {

					thrm2 := general{
						PTID:    ID1,
						Type:    "THRM",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					thrm2.Source.Path = append(thrm2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !thrm2.CompareEvents(allTHRM) {
						//helper.WriteTOFile(jsonFile, thrm2)
						allTHRM = append(allTHRM, thrm2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "THRM with invalid date format: " + date + ", code: " + m["THRM2"]
					} else if m["THRM2"] == "1" {
						f.Notes = "THRM with empty date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event HEML
				date, est = helper.CheckDateFormat(e, path, j, i, "HEML1_Date", m["HEML1_D"])
				if est == 0 || est == 1 {
					heml1 := general{
						PTID:    ID1,
						Type:    "HEML",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					heml1.Source.Path = append(heml1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !heml1.CompareEvents(alllHEML) {
						//helper.WriteTOFile(jsonFile, heml1)
						alllHEML = append(alllHEML, heml1)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "HEML with invalid date format: " + date + ", code: " + m["HEML1"]
					} else if m["HEML1"] == "1" {
						f.Notes = "HEML with empty date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}
				//HEML2
				date, est = helper.CheckDateFormat(e, path, j, i, "HEML2_Date", m["HEML2_D"])

				if est == 0 || est == 1 {

					heml2 := general{
						PTID:    ID1,
						Type:    "HEML",
						Date:    date,
						DateEst: est,
						Fix:     []errMessage{},
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					heml2.Source.Path = append(heml2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !heml2.CompareEvents(alllHEML) {
						//helper.WriteTOFile(jsonFile, heml2)
						alllHEML = append(alllHEML, heml2)
					}
				} else if est == 2 || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Notes = "HEML with invalid date format: " + date + ", code: " + m["HEML2"]
					} else if m["HEML2"] == "1" {
						f.Notes = "HEML with empty date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}
			}

		}
	}
}

// WriteToJSON writes from slices to JSON objects
func WriteToJSON(jsonFile *os.File, allARH []arh, allDVT []general, allDths []death, allFUMI []general,
	allFUPACE []general, allFix []general, allFollowUps []followUp, allLAK []general,
	allOperation []operation, allPVL []general, allSBE []sbe, allSVD []general,
	allStroke []stroke, allTHRM []general, allTIA []tia, alllHEML []general) {

	for _, o := range allARH {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allDVT {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allDths {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allFUMI {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allFUPACE {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allFix {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allFollowUps {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allLAK {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allOperation {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allPVL {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allSBE {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allSVD {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allStroke {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allTHRM {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allTIA {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range alllHEML {
		helper.WriteTOFile(jsonFile, o)
	}

}
