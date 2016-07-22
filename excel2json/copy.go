// Package excel2json provides functions that loop through excel files,
// read data from these files and create different events.
package excel2json

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

var (
	allFollowUps []followUp    // store followUp events
	allDths      []death       // store death events
	allReOper    []reOperation // store reoperation events
	allTIA       []tia         // store TIA events
	allStroke    []stroke      // store stroke events
	allSBE       []sbe         // store SBE events
	allARH       []arh
	events       []general //store events including FUMI, FUPACE, SVD, PVL, DVT, ARH, THRM, HEML, LKA_D, Fix
	codes        []string  // status codes
	nums         []int     // numerical values for various codes
	floats       []float64
)

type source struct {
	Type string   `json:"type"`
	Path []string `json:"path"`
}

type operation struct {
	Type                 string
	MRN                  string `json:"mrn"`
	ResearchID           string `json:"research_id"`
	PeriopID, PTID, Date string
	DateEst              int `json:"date_est"`
	Surgeon              string
	Surgeries            []string
	Children             []string
	Parent               string
	Notes                string
	Source               source
	Fix                  []errMessage
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
	Status, FuNotes,
	Notes, LostOnDate, OtherNote string
	Plat, Coag int
	PoNYHA     float64
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
	Fix               []errMessage
}

// re-operation event
type reOperation struct {
	Type                   string
	MRN                    string `json:"mrn"`
	ResearchID             string `json:"research_id"`
	PTID, Date             string
	DateEst                int `json:"date_est"`
	Reason, Surgery, Notes string
	Survival               int
	Fix                    []errMessage
}

// stroke event
type stroke struct {
	Type                  string
	MRN                   string `json:"mrn"`
	ResearchID            string `json:"research_id"`
	PTID, Date            string
	DateEst               int `json:"date_est"`
	Outcome, Agents, When int
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
	Fix             []errMessage
}

// SBE event
type sbe struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int `json:"date_est"`
	Organism   string
	Fix        []errMessage
}

// type of events that share the same variables
type general struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int          `json:"date_est"`
	Notes      string       `json:"omitempty"`
	Fix        []errMessage `json:"omitempty"`
}

// type of events that share the same variables
type arh struct {
	Type       string
	MRN        string `json:"mrn"`
	ResearchID string `json:"research_id"`
	PTID, Date string
	DateEst    int `json:"date_est"`
	Code       int
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
			return true
		} else if a.Coag == b.Coag && a.Date == b.Date && a.FuNotes == b.FuNotes &&
			a.LostOnDate == b.LostOnDate &&
			a.Notes == b.Notes && a.OtherNote == b.OtherNote && a.PTID == b.PTID &&
			a.Plat == b.Plat && a.PoNYHA == b.PoNYHA && a.Status == b.Status &&
			a.Type == b.Type {
			return true
			// check if same PTIDs in on file
		} //else if a.PTID == b.PTID && a.Path == b.Path {
		//e.Println("follow up events: Same PTID !", a.PTID, "Path:", a.Path, "sheet#:", j+1, "row#", i+2, b.PTID, b.Path)
		//}
	}
	return false
}

// CompareDeath checks if two death events are duplicate
func (a death) CompareDeath(s []death) bool {
	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Date == b.Date && a.PTID == b.PTID &&
			a.PrmDth == b.PrmDth && a.Reason == b.Reason && a.Type == b.Type {
			return true
			// check if same PTIDs in on file
		} //else if a.PTID == b.PTID && a.Date != b.Date {
		//e.Println("death events: Same person with different death dates!",
		//a.PTID, a.Date, "Path:", a.Path, "sheet#:", j+1, "row#", i+2, b.PTID, b.Date, b.Path)
		//}
	}
	return false
}

// CompareReOperation checks if two re-operation events are duplicate
func (a reOperation) CompareReOperation(s []reOperation) bool {

	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Date == b.Date && a.Notes == b.Notes &&
			a.PTID == b.PTID && a.Reason == b.Reason && a.Surgery == b.Surgery &&
			a.Survival == b.Survival && a.Type == b.Type {
			return true
			// check if same PTIDs in on file
		} //else if a.PTID == b.PTID && a.Path == b.Path {
		//e.Println("re operation events: Same PTID !", a.PTID, "Path:", a.Path, "sheet#:", j+1, "row#", i+2, b.PTID, b.Path)
		//}
	}
	return false
}

// CompareTia checks if two TIA events are duplicate
func (a tia) CompareTia(s []tia) bool {
	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Agents == b.Agents && a.Date == b.Date &&
			a.Outcome == b.Outcome && a.PTID == b.PTID && a.Type == b.Type {
			return true
			// check if same PTIDs in on file
		} //else if a.PTID == b.PTID && a.Path == b.Path {
		//e.Println("last know alive events: Same PTID !", a.PTID, "Path:", a.Path, "sheet#:", j+1, "row#", i+2, b.PTID, b.Path)
		//}
	}
	return false
}

// CompareStroke checks if two stroke events are duplicate
func (a stroke) CompareStroke(s []stroke) bool {
	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Agents == b.Agents && a.Date == b.Date && a.When == b.When &&
			a.Outcome == b.Outcome && a.PTID == b.PTID && a.Type == b.Type {
			return true
		}
	}
	return false
}

// CompareSbe checks if two sbe events are duplicate
func (a sbe) CompareSbe(s []sbe) bool {
	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Date == b.Date && a.Organism == b.Organism &&
			a.PTID == b.PTID && a.Type == b.Type {
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
			return true
		} else if a.Date == b.Date && a.PTID == b.PTID &&
			a.Type == b.Type {
			return true
		}
	}
	return false
}

func (a arh) CompareARH(s []arh) bool {
	for _, b := range s {
		if &a == &b {
			return true
		} else if a.Code == b.Code && a.Date == b.Date && a.PTID == b.PTID &&
			a.Type == b.Type {
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
}

// ReadExcelData uses the returned values of the function ExcelToSlice to
// build different types of events, and stores events to a json file.
func ReadExcelData(e *log.Logger, path string, jsonFile *os.File, columnsChecker string) {
	// slices is a slice of slices of maps, each map is a row in a excel file
	// keyList is a slice of slices of strings, each slice of strings is a header row of a excel file
	slices, keyList := helper.ExcelToSlice(e, path, columnsChecker)
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
				if m["FU_D"] != "" {

					date, est := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])

					fU := followUp{
						PTID:       ID1,
						Date:       date,
						Type:       "followup",
						Status:     S1,
						FuNotes:    m["FU NOTES"],
						Notes:      m["NOTES"],
						LostOnDate: m["STATUS=L DATE"],
						OtherNote:  m["STATUS=O REASON"],
						DateEst:    est}

					// check PTID
					// if 2 different PTIDs
					if diffID {
						msg := errMessage{"PTID", "two different PTIDs:" + ID1 + ", " + ID2}
						fU.Fix = append(fU.Fix, msg)
					}

					// check STATUS
					if diffStatus {
						msg := errMessage{"Status", "two different Statuses:" + S1 + ", " + S2}
						fU.Fix = append(fU.Fix, msg)
						if !helper.StringInSlice(1, S1, codes[:4]) {
							fU.Status = S2
						}
					}

					// check if these 3 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&fU.Coag, m["COAG"])
					helper.CheckFloatEmpty(&fU.PoNYHA, m["PO_NYHA"])
					helper.CheckEmpty(&fU.Plat, m["PLAT"])

					// Validate fields' values

					// date_est == 2 means invalid date value
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						fU.Fix = append(fU.Fix, msg)
						fU.DateEst, fU.Date = 0, ""
					}
					if !helper.StringInSlice(1, fU.Status, codes) {
						msg := errMessage{"code", "invalid value:" + fU.Status}
						fU.Fix = append(fU.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "Status", fU.Status)
					}
					if !helper.FloatInSlice(fU.PoNYHA, floats[1:]) {
						msg := errMessage{"PO_NYHA", "invalid value:" + m["PO_NYHA"]}
						fU.Fix = append(fU.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
					}
					if !helper.IntInSlice(fU.Coag, nums[:3]) {
						msg := errMessage{"COAG", "invalid value:" + m["COAG"]}
						fU.Fix = append(fU.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
					}
					if !helper.IntInSlice(fU.Plat, nums[:3]) {
						msg := errMessage{"PLAT", "invalid value:" + m["PLAT"]}
						fU.Fix = append(fU.Fix, msg)
						//	errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !fU.CompareFollowUps(allFollowUps) {
						allFollowUps = append(allFollowUps, fU)
						//helper.WriteTOFile(jsonFile, fU)
					}
				} else if m["FU NOTES"] != "" {
					// create LKA_D Event
					if m["LKA_D"] != "" {
						date, est := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])
						l := general{
							PTID:    ID1,
							Type:    "LKA_D",
							Date:    date,
							DateEst: est,
							Notes:   "FU Notes:" + m["FU NOTES"]}

						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							l.Fix = append(l.Fix, msg)
							l.Date, l.DateEst = "", 0

						}
						// if no duplicates, write this object to the json file and store in a slice

						//helper.WriteTOFile(jsonFile, l)
						events = append(events, l)

					}
					//if LKA_D is empty
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Notes:   "FU NOTES without date associated: here are my followup notes:" + m["FU NOTES"]}

					//helper.WriteTOFile(jsonFile, f)
					events = append(events, f)
				}

				// Event Death
				if m["DTH_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"])
					d := death{
						PTID:    ID1,
						Type:    "death",
						Date:    date,
						Reason:  m["REASDTH"],
						DateEst: est}

					// check if these two columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&d.PrmDth, m["PRM_DTH"])

					// Check date value
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						d.Fix = append(d.Fix, msg)
						d.Date, d.DateEst = "", 0
					}

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
					if !helper.IntInSlice(d.PrmDth, nums[:6]) {
						msg := errMessage{"PRM_DTH", "invalid value:" + m["PRM_DTH"]}
						d.Fix = append(d.Fix, msg)
						//errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !d.CompareDeath(allDths) {
						//helper.WriteTOFile(jsonFile, d)
						allDths = append(allDths, d)
					}
				}

				// Event FUREOP -> Event operation
				if m["FUREOP_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"])
					op := operation{
						PTID:    ID1,
						Type:    "operation",
						Date:    date,
							DateEst: est
							PeriopID: nil,
						//Reason:  m["REASREOP"],
						//Surgery: m["REOPSURG"],
						//Notes:   m["REOPNOTES"],
			
					}
					// check if these 2 columns are empty or not,
					// if empty, assign -9
					helper.CheckEmpty(&re.Survival, m["REOPSURVIVAL"])

					// date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						re.Fix = append(re.Fix, msg)
						re.Date, re.DateEst = "", 0
					}
					// Validate fields' values
					if S1 != "R" || m["FUREOP"] != "1" {
						msg := errMessage{"re-operation", "status is not R," + "status:" + S1 + "code:" + m["FUREOP"]}
						re.Fix = append(re.Fix, msg)
						//errlog.Differ(e, 4, path, j, i, re.PTID, "INCORRECT INFO OF REOPERATION!")
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !re.CompareReOperation(allReOper) {
						//helper.WriteTOFile(jsonFile, re)
						allReOper = append(allReOper, re)
					}

				} //else if m["REASREOP"] != "" || m["REOPSURG"] != "" || m["REOPNOTES"] != "" || m["NONVALVE REOP"] != "" {
				//e.Println(path, "INFO: No re-operation date, but re-operation info exists.", "Row#", i+2)
				//}

				// TE1

				if m["TE1_Date"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "TE1_Date", m["TE1_D"])

					if m["TE1"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE1"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est}

						// field when
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&s.Outcome, m["TE1_OUT"])
						helper.CheckEmpty(&s.Agents, m["ANTI_TE1"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							s.Fix = append(s.Fix, msg)
							s.Date, s.DateEst = "", 0
						}
						// Validate fields' values
						if !helper.IntInSlice(s.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE1_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(s.Agents, nums[:5]) && (s.Agents != 8) {
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
							DateEst: est}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&t.Outcome, m["TE1_OUT"])
						helper.CheckEmpty(&t.Agents, m["ANTI_TE1"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							t.Fix = append(t.Fix, msg)
							t.Date, t.DateEst = "", 0

						}
						// Validate fields' values
						if !helper.IntInSlice(t.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE1_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(t.Agents, nums[:5]) && (t.Agents != 8) {
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
				} //else if m["TE1"] != "0" {
				//e.Println(path, "INFO: No TE1 date, but TE1 info exists.", "Row#", i+2)
				//}

				// TE2

				if m["TE2_Date"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])

					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE2"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est}

						// field when
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&s.Outcome, m["TE2_OUT"])
						helper.CheckEmpty(&s.Agents, m["ANTI_TE2"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							s.Fix = append(s.Fix, msg)
							s.Date, s.DateEst = "", 0

						}
						// Validate fields' values
						if !helper.IntInSlice(s.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE2_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(s.Agents, nums[:5]) && (s.Agents != 8) {
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
							DateEst: est}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&t.Outcome, m["TE2_OUT"])
						helper.CheckEmpty(&t.Agents, m["ANTI_TE2"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							t.Fix = append(t.Fix, msg)
							t.Date, t.DateEst = "", 0

						}
						// Validate fields' values
						if !helper.IntInSlice(t.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE2_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(t.Agents, nums[:5]) && (t.Agents != 8) {
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
				} //else if m["TE2"] != "0" {
				//e.Println(path, "INFO: No TE2 date, but TE2 info exists.", "Row#", i+2)
				//}

				// TE3

				if m["TE3_Date"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE3"] == "2" {

						// Event stroke
						s := stroke{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est}

						// field when
						if m["DATEOR"] != "" {
							operDate, _ := helper.CheckDateFormat(e, path, j, i, "DATEOR", m["DATEOR"])
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&s.Outcome, m["TE3_OUT"])
						helper.CheckEmpty(&s.Agents, m["ANTI_TE3"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							s.Fix = append(s.Fix, msg)
							s.Date, s.DateEst = "", 0

						}
						// Validate fields' values
						if !helper.IntInSlice(s.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE3_OUT"]}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(s.Agents, nums[:5]) && (s.Agents != 8) {
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
							DateEst: est}

						// check if these 2 columns are empty or not,
						// if empty, assign -9
						helper.CheckEmpty(&t.Outcome, m["TE3_OUT"])
						helper.CheckEmpty(&t.Agents, m["ANTI_TE3"])
						// Check date value
						if est == 2 {
							msg := errMessage{"date", "invalid date:" + date}
							t.Fix = append(t.Fix, msg)
							t.Date, t.DateEst = "", 0

						}
						// Validate fields' values
						if !helper.IntInSlice(t.Outcome, nums[:5]) {
							msg := errMessage{"outcome", "invalid value:" + m["TE3_OUT"]}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.IntInSlice(t.Agents, nums[:5]) && (t.Agents != 8) {
							msg := errMessage{"agents", "invalid value:" + m["ANTI_TE3"]}
							t.Fix = append(t.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTia(allTIA) {
							//	helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} //else if m["TE3"] != "0" {
				//e.Println(path, "INFO: No TE3 date, but TE3 info exists.", "Row#", i+2)
				//	}

				// Event FUMI
				if m["FUMI_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])
					f1 := general{
						PTID:    ID1,
						Type:    "FUMI",
						Date:    date,
						DateEst: est}

					// date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						f1.Fix = append(f1.Fix, msg)
						f1.Date, f1.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !f1.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, f1)
						events = append(events, f1)
					}
				}

				// Event FUPACE
				if m["FUPACE_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])
					f2 := general{
						PTID:    ID1,
						Type:    "FUPACE",
						Date:    date,
						DateEst: est}

					// date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						f2.Fix = append(f2.Fix, msg)
						f2.Date, f2.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !f2.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, f2)
						events = append(events, f2)
					}
				}

				// Event SBE
				if m["SBE1_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "SBE1_Date", m["SBE1_D"])
					sbe1 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est}

					// some sheets may have organism instead of ORGANISM
					if m["SBE1 ORGANISM"] != "" {
						sbe1.Organism = m["SBE1 ORGANISM"]
					} else {
						sbe1.Organism = m["SBE1 organism"]
					}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						sbe1.Fix = append(sbe1.Fix, msg)
						sbe1.Date, sbe1.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe1.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe1)
						allSBE = append(allSBE, sbe1)
					}
				}

				if m["SBE2_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "SBE2_Date", m["SBE2_D"])
					sbe2 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est}

					// some sheets may have organism instead of ORGANISM
					if m["SBE2 ORGANISM"] != "" {
						sbe2.Organism = m["SBE2 ORGANISM"]
					} else {
						sbe2.Organism = m["SBE2 organism"]
					}
					// check date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						sbe2.Fix = append(sbe2.Fix, msg)
						sbe2.Date, sbe2.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe2.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe2)
						allSBE = append(allSBE, sbe2)
					}
				}

				if m["SBE3_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "SBE3_Date", m["SBE3_D"])
					sbe3 := sbe{
						PTID:    ID1,
						Type:    "SBE",
						Date:    date,
						DateEst: est}

					// some sheets may have organism instead of ORGANISM
					if m["SBE3 ORGANISM"] != "" {
						sbe3.Organism = m["SBE3 ORGANISM"]
					} else {
						sbe3.Organism = m["SBE3 organism"]
					}
					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						sbe3.Fix = append(sbe3.Fix, msg)
						sbe3.Date, sbe3.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe3.CompareSbe(allSBE) {
						//helper.WriteTOFile(jsonFile, sbe3)
						allSBE = append(allSBE, sbe3)
					}
				}

				// Event SVD
				if m["SVD_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])
					s4 := general{
						PTID:    ID1,
						Type:    "SVD",
						Date:    date,
						DateEst: est}

					// date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						s4.Fix = append(s4.Fix, msg)
						s4.Date, s4.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !s4.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, s4)
						events = append(events, s4)
					}
				}
				// Event PVL
				if m["PVL1_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "PVL1_Date", m["PVL1_D"])
					pvl1 := general{
						PTID:    ID1,
						Type:    "PVL",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						pvl1.Fix = append(pvl1.Fix, msg)
						pvl1.Date, pvl1.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl1.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, pvl1)
						events = append(events, pvl1)
					}
				}

				if m["PVL2_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "PVL2_Date", m["PVL2_D"])
					pvl2 := general{
						PTID:    ID1,
						Type:    "PVL",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						pvl2.Fix = append(pvl2.Fix, msg)
						pvl2.Date, pvl2.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl2.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, pvl2)
						events = append(events, pvl2)
					}
				}

				// Event DVT
				if m["DVT_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])
					d1 := general{
						PTID:    ID1,
						Type:    "DVT",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						d1.Fix = append(d1.Fix, msg)
						d1.Date, d1.DateEst = "", 0
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !d1.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, d1)
						events = append(events, d1)
					}
				}

				// Event ARH
				if m["ARH1_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])
					arh1 := arh{
						PTID:    ID1,
						Type:    "ARH",
						Date:    date,
						DateEst: est}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&arh1.Code, m["ARH1"])
					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						arh1.Fix = append(arh1.Fix, msg)
						arh1.Date, arh1.DateEst = "", 0
					}
					// Validate fields' values
					if !helper.IntInSlice(arh1.Code, nums[:]) {
						msg := errMessage{"code", "invalid value:" + m["ARH1"]}
						arh1.Fix = append(arh1.Fix, msg)
						//errlog.ErrorLog(e, path, j, arh1.PTID, i, arh1.Type, "ARH1", m["ARH1"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh1.CompareARH(allARH) {
						//helper.WriteTOFile(jsonFile, arh1)
						allARH = append(allARH, arh1)
					}
				}

				if m["ARH2_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])
					arh2 := arh{
						PTID:    ID1,
						Type:    "ARH",
						Date:    date,
						DateEst: est}
					// check if this column is empty or not;
					// if empty, assign -9
					helper.CheckEmpty(&arh2.Code, m["ARH2"])
					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						arh2.Fix = append(arh2.Fix, msg)
						arh2.Date, arh2.DateEst = "", 0
					}
					// Validate fields' values
					if !helper.IntInSlice(arh2.Code, nums[:]) {
						msg := errMessage{"code", "invalid value:" + m["ARH2"]}
						arh2.Fix = append(arh2.Fix, msg)
						//errlog.ErrorLog(e, path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh2.CompareARH(allARH) {
						//helper.WriteTOFile(jsonFile, arh2)
						allARH = append(allARH, arh2)
					}
				}
				// Event THRM
				if m["THRM1_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "THRM1_Date", m["THRM1_D"])
					thrm1 := general{
						PTID:    ID1,
						Type:    "THRM",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						thrm1.Fix = append(thrm1.Fix, msg)
						thrm1.Date, thrm1.DateEst = "", 0
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !thrm1.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, thrm1)
						events = append(events, thrm1)
					}
				}

				if m["THRM2_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "THRM2_Date", m["THRM2_D"])
					thrm2 := general{
						PTID:    ID1,
						Type:    "THRM",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						thrm2.Fix = append(thrm2.Fix, msg)
						thrm2.Date, thrm2.DateEst = "", 0
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !thrm2.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, thrm2)
						events = append(events, thrm2)
					}
				}

				// Event HEML
				if m["HEML1_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "HEML1_Date", m["HEML1_D"])
					heml1 := general{
						PTID:    ID1,
						Type:    "HEML",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						heml1.Fix = append(heml1.Fix, msg)
						heml1.Date, heml1.DateEst = "", 0
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !heml1.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, heml1)
						events = append(events, heml1)
					}
				}

				if m["HEML2_D"] != "" {
					date, est := helper.CheckDateFormat(e, path, j, i, "HEML2_Date", m["HEML2_D"])
					heml2 := general{
						PTID:    ID1,
						Type:    "HEML",
						Date:    date,
						DateEst: est}

					//date
					if est == 2 {
						msg := errMessage{"date", "invalid date:" + date}
						heml2.Fix = append(heml2.Fix, msg)
						heml2.Date, heml2.DateEst = "", 0
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !heml2.CompareEvents(events) {
						//helper.WriteTOFile(jsonFile, heml2)
						events = append(events, heml2)
					}
				}
			}
		}
	}
}
