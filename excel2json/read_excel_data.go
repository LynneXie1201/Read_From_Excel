// Package excel2json provides functions that loop through excel files,
// read data from these files and create different events.
package excel2json

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"regexp"
)

// Initialize before other functions get executed
func init() {
	nums = []int{0, -9, 1, 2, 3, 4, 5}                           // list of int numbers for validation
	codes = []string{"N", "D", "L", "O", "A", "R", ""}           // valid status codes
	floats = []float64{0, -9, 1, 2, 3, 4, 5, 1.5, 2.5, 3.5, 4.5} // list of float numbers for validation
}

// ReadExcelData uses the returned values from the function ExcelToSlice to
// create different types of events, and stores in the slices.
func ReadExcelData(e *log.Logger, path string, jsonFile *os.File, columnsChecker string) {
	// slices is a slice of slices of maps, each slice of maps represents a sheet, and each map is a row in a sheet
	// keyList is a slice of slices of strings, each slice of strings is a header row of a sheet
	slices, keyList := ExcelToSlice(e, path, columnsChecker)
	// get the sub path of the original path
	path = helper.SubPath(path, "valve_registry")
	// j is the index of sheets
	// s is a slice of maps representing the excel sheet of index j
	for j, s := range slices {
		// if s equals nil, s is not a follow_up sheet
		if s == nil {
			fmt.Println("oops! this is not a follow_up sheet: ", path, "sheet #:", j+1)
		} else {
			// s is a follow_up excel sheet
			fmt.Println("Bingo! this is a follow_up sheet: ", path, "sheet #:", j+1)
			// keys is the header row of the excel sheet of index j
			keys := keyList[j]
			// check the number of PTID and STATUS' colomns
			// p1, p2 is the PTID column names
			p1, p2 := helper.CheckPtidColumns(e, path, j, keys)
			// st1, st2 is the status column names
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

				// assign status
				diffStatus := helper.AssignStatus(&S1, &S2)
				// check if format of PTID is LLLFDDMMYY
				helper.CheckPtidFormat(ID1, e, path, j, i)

				// Event followup
				var coag, plat int
				var poNYHA float64
				// Validate int and float values
				coagValid := helper.CheckIntValue(&coag, m["COAG"], nums[:3])
				nyhaValid := helper.CheckFloatValue(&poNYHA, m["PO_NYHA"], floats[1:])
				platValid := helper.CheckIntValue(&plat, m["PLAT"], nums[:3])

				// if status equals "L", it belongs to lost_to_followup event, not followup event
				if S1 != "L" && S2 != "L" {
					// check FU_D format
					date, est := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])
					// create followup notes
					fuNotes := helper.FollowupNotes(S1, m["FU NOTES"], m["NOTES"], m["STATUS=O REASON"], plat, coag, poNYHA)
					// est equals 0 or 1 means that the date format was parsed to YYYY-MM-DD
					if est == 0 || est == 1 {
						// create event
						fu := followups{
							PTID:    ID1,
							Date:    date,
							Type:    "followup",
							Status:  &S1,
							Plat:    plat,
							PoNYHA:  poNYHA,
							Coag:    coag,
							Unusual: m["STATUS=O REASON"],
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// source: add path
						fu.Source.Path = append(fu.Source.Path, path)

						// add Notes
						if !(m["NOTES"] == "" && m["FU NOTES"] == "") {
							fu.Notes = m["FU NOTES"] + " " + m["NOTES"]
						}

						// check PTID
						if diffID {
							msg := errMessage{"PTID", "two different PTIDs: '" + ID1 + "', '" + ID2 + "' "}
							fu.Fix = append(fu.Fix, msg)
						}

						// check STATUS

						// if true means both statuses are non-empty and not equal
						if diffStatus {
							msg := errMessage{"Status", "two different Statuses: '" + S1 + "', '" + S2 + "'"}
							fu.Fix = append(fu.Fix, msg)
							// if one of the codes is D, L, N, or O and the other code is A or R, put the D, L, N or O
							if helper.StringInSlice(0, S1, codes[4:6]) && helper.StringInSlice(0, S2, codes[:4]) {
								fu.Status = &S2
							}
						}

						// validate fields' values
						if *fu.Status == "" {
							fu.Status = nil
						} else if !helper.StringInSlice(0, S1, codes) {
							msg := errMessage{"code", "invalid value: '" + S1 + "'"}
							fu.Fix = append(fu.Fix, msg)
						}

						if !nyhaValid {
							msg := errMessage{"PO_NYHA", "invalid value: '" + m["PO_NYHA"] + "'"}
							fu.Fix = append(fu.Fix, msg)
							//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
						}
						if !coagValid {
							msg := errMessage{"COAG", "invalid value: '" + m["COAG"] + "'"}
							fu.Fix = append(fu.Fix, msg)
							//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
						}
						if !platValid {
							msg := errMessage{"PLAT", "invalid value: '" + m["PLAT"] + "'"}
							fu.Fix = append(fu.Fix, msg)
							//	errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
						}
						// if no duplicates, store in a slice
						if !fu.CompareFollowups(allFollowUps) {
							allFollowUps = append(allFollowUps, fu)
						}

						// est == 3 means that date has invalid format
					} else if est == 3 {

						// create a  fix event
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						// add path
						f.Source.Path = append(f.Source.Path, path)
						// add msg
						f.Msg = "followup event with invalid date: '" + date +
							"', here is the follow up info: " + fuNotes

						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}

						// est equal 2: follow up date is empty, but has fu notes, then create lka events
					} else if est == 2 {

						if m["FU NOTES"] != "" || (coag != -9 && coag != 0) || (plat != -9 && plat != 0) ||
							(poNYHA != -9 && poNYHA != 0) || m["STATUS=O REASON"] != "" || m["NOTES"] != "" {
							lkaDate, lkaEst := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])

							// if last_known_alive date exists
							if lkaEst == 0 || lkaEst == 1 {
								// create LKA_D Event
								lka := followups{
									PTID:    ID1,
									Type:    "last_known_alive",
									Date:    lkaDate,
									Coag:    coag,
									PoNYHA:  poNYHA,
									Plat:    plat,
									Unusual: m["STATUS=O REASON"],
									DateEst: lkaEst,
									Source:  source{Type: "followup", Path: []string{}}}

								// source: add path
								lka.Source.Path = append(lka.Source.Path, path)

								// add Notes if exists
								if !(m["NOTES"] == "" && m["FU NOTES"] == "") {
									lka.Notes = m["FU NOTES"] + " " + m["NOTES"]
								}
								// validate PO_NYHA
								if !nyhaValid {
									msg := errMessage{"PO_NYHA", "invalid value: '" + m["PO_NYHA"] + "'"}
									lka.Fix = append(lka.Fix, msg)
								}

								// validate COAG
								if !coagValid {
									msg := errMessage{"COAG", "invalid value: '" + m["COAG"] + "'"}
									lka.Fix = append(lka.Fix, msg)
								}

								// validate PLAT
								if !platValid {
									msg := errMessage{"PLAT", "invalid value: '" + m["PLAT"] + "'"}
									lka.Fix = append(lka.Fix, msg)

								}

								// if no duplicates, store in a slice
								if !lka.CompareFollowups(allLKA) {
									allLKA = append(allLKA, lka)
								}
								// last_known_alive is not available
							} else if lkaEst == 2 || lkaEst == 3 {
								f := general{
									PTID:    ID1,
									Type:    "fix",
									Date:    "1900-01-01",
									DateEst: 1,
									Source:  source{Type: "followup", Path: []string{}}}
								// last_known_alive date is empty
								if lkaEst == 2 {
									f.Msg = "followup and LKA without date associated, here are my followup notes: " + fuNotes
								} else {
									// LKA date with invalid format
									f.Msg = "LKA Date with invalid format: '" + date +
										"' ,and FU NOTES without date associated. Here are my followup notes: " + fuNotes
								}

								// Source: add path
								f.Source.Path = append(f.Source.Path, path)
								// if no duplicates, store in a slice
								if !f.CompareEvents(allFix) {
									allFix = append(allFix, f)
								}
							}
						}
					}
					// status equals L, but COAG or PLAT or PO_NYHA have values
				} else if (coag != -9 && coag != 0) || (plat != -9 && plat != 0) || (poNYHA != -9 && poNYHA != 0) {
					e.Println(path, "Status = L, but COAG OR PLAT OR PO_NYHA HAVE VALUES!", "row: ", i+2)
				}

				// Event Lost followups
				// if one of the STATUS columns is “L” and the other is “D” or “N”, do not create the “lost_to_followup”
				// otherwise, create the “lost_to_followup” event if one is “L”
				if S1 == "L" && !helper.StringInSlice(0, S2, codes[:2]) || (S2 == "L" && !helper.StringInSlice(0, S1, codes[:2])) {
					date, est := helper.CheckDateFormat(e, path, j, i, "Status=L_Date", m["STATUS=L DATE"])
					if est != 3 {
						// date will be (in order of preference) either the “STATUS=L DATE” field, or the FU_D or LKA_D or
						// if none of those are available, put 1900-02-02

						lost := general{
							PTID:    ID1,
							Type:    "lost_to_followup",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add Notes

						if !(m["FU NOTES"] == "" && m["NOTES"] == "" && m["STATUS=O REASON"] == "") {
							lost.Notes = m["FU NOTES"] + " " + m["NOTES"] + " " + m["STATUS=O REASON"]
						}

						// date is empty
						if est == 2 {
							lkaDate, lkaEst := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])
							fuDate, fuEst := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])
							if fuEst == 0 || fuEst == 1 {
								lost.Date = fuDate
								lost.DateEst = fuEst
							} else if lkaEst == 0 || lkaEst == 1 {
								lost.Date = lkaDate
								lost.DateEst = lkaEst
							} else {
								lost.Date = "1900-02-02"
								lost.DateEst = 1
							}
						}
						// Source: add path
						lost.Source.Path = append(lost.Source.Path, path)

						// if no duplicates, store in a slice
						if !lost.CompareEvents(allLostFollowups) {
							allLostFollowups = append(allLostFollowups, lost)
						}
						// est == 3, date has invalid format
					} else {
						// create a fix event
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-02-02",
							DateEst: 1,
							Msg: "Invalid STATUS=L DATE: '" + m["STATUS=L DATE"] + "', Notes: '" + m["FU NOTES"] +
								" " + m["NOTES"] + " " + m["STATUS=O REASON"] + "'",
							Source: source{Type: "followup", Path: []string{}}}

						// add path
						f.Source.Path = append(f.Source.Path, path)
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}
				////////////////////////////////////////////////////////////ends here//////////////////////////////////////////////////////
				// Event Death
				var operDate, operative string
				var operEst int
				date, est := helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"])

				for _, k := range keys {
					matched, _ := regexp.MatchString("^.*DATEOR$", k)
					if matched {
						operDate, operEst = helper.CheckDateFormat(e, path, j, i, "DATEOR", m[k])
						break
					}
				}
				// check Operative

				if m["SURVIVAL"] == "0" {
					operative = "1"
				} else {
					operative = "0"
				}

				if est == 0 || est == 1 {
					d := death{
						PTID:    ID1,
						Type:    "death",
						Date:    date,
						Reason:  m["REASDTH"],
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					d.Source.Path = append(d.Source.Path, path)

					// check Operative

					if m["SURVIVAL"] == "0" {
						d.Operative = 1
						if operDate != date {
							msg := errMessage{"operative", "Date of surgery is '" + operDate + "', please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					} else if m["SURVIVAL"] == "1" {
						if operDate == date {
							msg := errMessage{"operative", "Date of surgery is '" + operDate + "', please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					}

					// Validate fields' values
					if !helper.CheckIntValue(&d.PrmDth, m["PRM_DTH"], nums[:6]) {
						msg := errMessage{"PRM_DTH", "invalid value: '" + m["PRM_DTH"] + "'"}
						d.Fix = append(d.Fix, msg)
						//errlog.ErrorLog(e, path, j, d.PTID, i, d.Type, "PRM_DTH", m["PRM_DTH"])
					}

					if S1 != "D" && S1 != "N" {
						msg := errMessage{"status", "invalid value: '" + S1 + "'"}
						d.Fix = append(d.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !(&d).CompareDeath(&allDths) {
						//helper.WriteTOFile(jsonFile, d)
						allDths = append(allDths, d)
					}
				} else if est == 3 {
					// est == 3 means invalid date format
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Msg = "Death event with invalid date format: '" + date + "'" +
						helper.DeathNotes(m["PRM_DTH"], m["REASDTH"], operative)

					// Source: add path
					f.Source.Path = append(f.Source.Path, path)

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
					// else est == 2
				} else if m["PRM_DTH"] != "0" || m["REASDTH"] != "" || m["DIED"] == "1" {

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Msg = "Death event with no date associated" +
						helper.DeathNotes(m["PRM_DTH"], m["REASDTH"], operative)

					// Source: add path
					f.Source.Path = append(f.Source.Path, path)

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}

				}

				// Event FUREOP -> operation Event

				date, est = helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"])
				opString := helper.OperationNotes(m["REASREOP"], m["REOPSURVIVAL"],
					m["REOPNOTES"], m["REOPSURG"], m["NONVALVE REOP"])
				if est == 0 || est == 1 {

					op := operation{
						PTID:    ID1,
						Type:    "operation",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					op.Source.Path = append(op.Source.Path, path)

					var survival int
					if !helper.CheckIntValue(&survival, m["REOPSURVIVAL"], nums[:3]) {
						msg := errMessage{"survival", "invalid value: '" + m["REOPSURVIVAL"] + "'"}
						op.Fix = append(op.Fix, msg)
					}

					// add re-op strirng

					if !(m["REASREOP"] == "" && m["REOPSURVIVAL"] == "0" && m["REOPNOTES"] == "" &&
						m["REOPSURG"] == "" && m["NONVALVE REOP"] == "") {

						msg := errMessage{"operation", opString}
						op.Fix = append(op.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !op.CompareOperation(allOperation) {

						allOperation = append(allOperation, op)
					}

				} else if est == 3 {

					// create a event fix
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Msg = "Invalid REOP date format: '" + m["FUREOP_D"] + "', here is the re-operation info: " + opString
					// Source: add path
					f.Source.Path = append(f.Source.Path, path)

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				} else if m["FUREOP"] == "1" || m["REASREOP"] != "" || m["REOPNOTES"] != "" || m["REOPSURG"] != "" || m["NONVALVE REOP"] != "" {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					f.Msg = "REOP fields without date associated, here is the re-operation info: " + opString

					// Source: add path
					f.Source.Path = append(f.Source.Path, path)

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
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field When
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)

							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE1_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)

						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE1"] + "'"}
							s.Fix = append(s.Fix, msg)

						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareTE(allStroke) {
							allStroke = append(allStroke, s)
						}
					} else if m["TE1"] == "3" {
						// Event TIA
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE1_OUT"] + "'"}
							t.Fix = append(t.Fix, msg)

						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE1"] + "'"}
							t.Fix = append(t.Fix, msg)

						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTE(allTIA) {

							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 {

					if m["TE1"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke
					if m["TE1"] == "2" {

						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "stroke with no date but code exists, " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])

						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

						// TIA
					} else if m["TE1"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "TIA with no date but code exists, " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])
						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

					}

				} else if est == 3 {

					if m["TE1"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke or TIA
					if m["TE1"] == "2" || m["TE1"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						if m["TE1"] == "2" {

							f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])

						} else if m["TE1"] == "3" {
							f.Msg = "TIA with invalid date format: '" + date + "', " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])
						}

						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}

				}
				// TE2
				date, est = helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])

				if est == 0 || est == 1 {

					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE2"] == "2" {

						// Event stroke
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field When
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)

							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE2_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE2"] + "'"}
							s.Fix = append(s.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareTE(allStroke) {
							//helper.WriteTOFile(jsonFile, s)
							allStroke = append(allStroke, s)
						}
					} else if m["TE2"] == "3" {
						// Event TIA
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE2_OUT"] + "'"}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE2"] + "'"}
							t.Fix = append(t.Fix, msg)

						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTE(allTIA) {
							//helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 {

					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke
					if m["TE2"] == "2" {

						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "stroke with no date but code exists, " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])

						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

						// TIA
					} else if m["TE2"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "TIA with no date but code exists, " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])
						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

					}

				} else if est == 3 {

					if m["TE2"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke
					if m["TE2"] == "2" || m["TE2"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						if m["TE2"] == "2" {
							f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])

						} else if m["TE1"] == "3" {
							f.Msg = "TIA with invalid date format: '" + date + "', " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])
						}

						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}

				}

				// TE3
				date, est = helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])
				if est == 0 || est == 1 {

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					} else if m["TE3"] == "2" {

						// Event stroke
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						s.Source.Path = append(s.Source.Path, path)

						// field When
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)

							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// Validate fields' values
						if !helper.CheckIntValue(&s.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE3_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE3"] + "' "}
							s.Fix = append(s.Fix, msg)

						}
						// if no duplicates, write this object to the json file and store in a slice
						if !s.CompareTE(allStroke) {
							//helper.WriteTOFile(jsonFile, s)
							allStroke = append(allStroke, s)
						}
					} else if m["TE3"] == "3" {
						// Event TIA
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						t.Source.Path = append(t.Source.Path, path)

						// Validate fields' values
						if !helper.CheckIntValue(&t.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE3_OUT"] + "' "}
							t.Fix = append(t.Fix, msg)
							//	errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "TE1_OUT", m["TE1_OUT"])
						}
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"agents", "invalid value: '" + m["ANTI_TE3"] + "' "}
							t.Fix = append(t.Fix, msg)
							//errlog.ErrorLog(e, path, j, te1.PTID, i, te1.Type, "ANTI_TE1", m["ANTI_TE1"])
						}
						// if no duplicates, write this object to the json file and store in a slice
						if !t.CompareTE(allTIA) {
							//helper.WriteTOFile(jsonFile, t)
							allTIA = append(allTIA, t)
						}
					}
				} else if est == 2 {

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke
					if m["TE3"] == "2" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "stroke with no date but code exists, " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])

						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

						// TIA
					} else if m["TE3"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						f.Msg = "TIA with no date but code exists, " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])
						if !f.CompareEvents(allFix) {

							allFix = append(allFix, f)
						}

					}

				} else if est == 3 {

					if m["TE3"] == "1" {
						e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: TE code is 1.")
					}

					// stroke
					if m["TE3"] == "2" || m["TE3"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}
						f.Source.Path = append(f.Source.Path, path)

						if m["TE3"] == "2" {
							f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])

						} else if m["TE1"] == "3" {
							f.Msg = "TIA with invalid date format: '" + date + "', " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])
						}

						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}

				}

				// Event FUMI

				// estimate FUMI date
				date, est = helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])
				if est == 0 || est == 1 {
					mi := general{
						PTID:    ID1,
						Type:    "myocardial_infarction",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					mi.Source.Path = append(mi.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !mi.CompareEvents(allFUMI) {
						allFUMI = append(allFUMI, mi)
					}

				} else if est == 3 || (est == 2 && m["FUMI"] == "1") {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)

					// add Notes
					if est == 3 {
						f.Msg = "FUMI with invalid date format: '" + date + "'"
					} else {
						f.Msg = "FUMI with no date but code is 1."
					}

					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event FUPACE
				date, est = helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])
				if est == 0 || est == 1 {
					pace := general{
						PTID:    ID1,
						Type:    "perm_pacemaker",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					pace.Source.Path = append(pace.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !pace.CompareEvents(allFUPACE) {

						allFUPACE = append(allFUPACE, pace)
					}
				} else if (est == 2 && m["FUPACE"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "FUPACE with invalid date format: '" + date + "'"
					} else if m["FUPACE"] == "1" {
						f.Msg = "FUPACE with no date but code is 1."
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

					sbe1 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
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
					if *sbe1.Organism == "" {
						sbe1.Organism = nil
					} else if !helper.CheckStringValue(*sbe1.Organism) {
						msg := errMessage{"organism", "invalid organism value: '" + *sbe1.Organism + "'"}
						sbe1.Fix = append(sbe1.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe1.CompareEvents(allSBE) {

						allSBE = append(allSBE, sbe1)
					}
				} else if (est == 2 && m["SBE1"] == "1") || est == 3 {
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
						f.Msg = "SBE with invalid date format: '" + date + "', organism: '" + organism + "'"
					} else if m["SBE1"] == "1" {
						f.Msg = "SBE with no date but code is 1, code: '" + m["SBE1"] + "', organism: '" + organism + "'"
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

					sbe2 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
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
					if *sbe2.Organism == "" {
						sbe2.Organism = nil
					} else if !helper.CheckStringValue(*sbe2.Organism) {
						msg := errMessage{"organism", "invalid organism value: '" + *sbe2.Organism + "'"}
						sbe2.Fix = append(sbe2.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe2.CompareEvents(allSBE) {

						allSBE = append(allSBE, sbe2)
					}
				} else if (est == 2 && m["SBE2"] == "1") || est == 3 {
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

						f.Msg = "SBE with invalid date format: '" + date + "', organism: '" + organism + "'"
					} else if m["SBE2"] == "1" {
						f.Msg = "SBE with no date but code is 1, organism: '" + organism + "'"
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
					sbe3 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
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

					if *sbe3.Organism == "" {
						sbe3.Organism = nil
					} else if !helper.CheckStringValue(*sbe3.Organism) {
						msg := errMessage{"organism", "invalid organism value: '" + *sbe3.Organism + "'"}
						sbe3.Fix = append(sbe3.Fix, msg)
					}

					// if no duplicates, write this object to the json file and store in a slice
					if !sbe3.CompareEvents(allSBE) {

						allSBE = append(allSBE, sbe3)
					}
				} else if (est == 2 && m["SBE3"] == "1") || est == 3 {
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
						f.Msg = "SBE with invalid date format: '" + date + "',  organism: '" + organism + "'"
					} else if m["SBE3"] == "1" {
						f.Msg = "SBE with no date but code is 1, organism: '" + organism + "'"
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event SVD
				date, est = helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])
				if est == 0 || est == 1 {
					svd := general{
						PTID:    ID1,
						Type:    "struct_valve_det",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					svd.Source.Path = append(svd.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !svd.CompareEvents(allSVD) {
						//helper.WriteTOFile(jsonFile, s4)
						allSVD = append(allSVD, svd)
					}
				} else if (est == 2 && m["SVD"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "SVD with invalid date format: '" + date + "'"
					} else if m["SVD"] == "1" {
						f.Msg = "SVD with no date but code is 1."
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
						Type:    "perivalvular_leak",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					pvl1.Source.Path = append(pvl1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl1.CompareEvents(allPVL) {
						//helper.WriteTOFile(jsonFile, pvl1)
						allPVL = append(allPVL, pvl1)
					}
				} else if (est == 2 && m["PVL1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "PVL with invalid date format: '" + date + "'"
					} else if m["PVL1"] == "1" {
						f.Msg = "PVL with no date but code is 1."
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
						Type:    "perivalvular_leak",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					pvl2.Source.Path = append(pvl2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !pvl2.CompareEvents(allPVL) {
						//helper.WriteTOFile(jsonFile, pvl2)
						allPVL = append(allPVL, pvl2)
					}
				} else if (est == 2 && m["PVL2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "PVL with invalid date format: '" + date + "'"
					} else if m["PVL2"] == "1" {
						f.Msg = "PVL with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event DVT
				date, est = helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])
				if est == 0 || est == 1 {

					dvt := general{
						PTID:    ID1,
						Type:    "deep_vein_thrombosis",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					dvt.Source.Path = append(dvt.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !dvt.CompareEvents(allDVT) {
						//helper.WriteTOFile(jsonFile, d1)
						allDVT = append(allDVT, dvt)
					}
				} else if (est == 2 && m["DVT"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "DVT with invalid date format: '" + date + "'"
					} else if m["DVT"] == "1" {
						f.Msg = "DVT with no date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event ARH
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])
				if est == 0 || est == 1 {
					arh1 := general{
						PTID:    ID1,
						Type:    "arh",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					arh1.Source.Path = append(arh1.Source.Path, path)

					// Validate fields' values
					if !helper.CheckIntValue(&arh1.Code, m["ARH1"], nums[:]) {
						msg := errMessage{"code", "invalid value: '" + m["ARH1"] + "'"}
						arh1.Fix = append(arh1.Fix, msg)

					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh1.CompareEvents(allARH) {
						//helper.WriteTOFile(jsonFile, arh1)
						allARH = append(allARH, arh1)
					}
				} else if (est == 2 && m["ARH1"] != "0" && m["ARH1"] != "") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "ARH with invalid date format: '" + date + "', " + helper.ArhCode(m["ARH1"])
					} else if m["ARH1"] != "0" && m["ARH1"] != "" {
						f.Msg = "ARH with no date but code is not 0 or empty, " + helper.ArhCode(m["ARH1"])
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// ARH2
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])
				if est == 0 || est == 1 {
					arh2 := general{
						PTID:    ID1,
						Type:    "arh",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					arh2.Source.Path = append(arh2.Source.Path, path)

					// Validate fields' values
					if !helper.CheckIntValue(&arh2.Code, m["ARH2"], nums[:]) {
						msg := errMessage{"code", "invalid value: '" + m["ARH2"] + "'"}
						arh2.Fix = append(arh2.Fix, msg)
						//errlog.ErrorLog(e, path, j, arh2.PTID, i, arh2.Type, "ARH2", m["ARH2"])
					}
					// if no duplicates, write this object to the json file and store in a slice
					if !arh2.CompareEvents(allARH) {
						//helper.WriteTOFile(jsonFile, arh2)
						allARH = append(allARH, arh2)
					}
				} else if (est == 2 && m["ARH2"] != "0" && m["ARH2"] != "") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "ARH with invalid date format: '" + date + "', " + helper.ArhCode(m["ARH2"])
					} else if m["ARH2"] != "0" && m["ARH2"] != "" {
						f.Msg = "ARH with no date but code is not 0 or empty, " + helper.ArhCode(m["ARH2"])
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
						Type:    "thromb_prost_valve",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					thrm1.Source.Path = append(thrm1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !thrm1.CompareEvents(allTHRM) {

						allTHRM = append(allTHRM, thrm1)
					}
				} else if (est == 2 && m["THRM1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "THRM with invalid date format: '" + date + "'"
					} else if m["THRM1"] == "1" {
						f.Msg = "THRM with empty date but code is 1."
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
						Type:    "thromb_prost_valve",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// Source: add path
					thrm2.Source.Path = append(thrm2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !thrm2.CompareEvents(allTHRM) {
						//helper.WriteTOFile(jsonFile, thrm2)
						allTHRM = append(allTHRM, thrm2)
					}
				} else if (est == 2 && m["THRM2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "THRM with invalid date format: '" + date + "'"
					} else if m["THRM2"] == "1" {
						f.Msg = "THRM with empty date but code is 1."
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
						Type:    "hemolysis_dx",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					heml1.Source.Path = append(heml1.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !heml1.CompareEvents(alllHEML) {

						alllHEML = append(alllHEML, heml1)
					}
				} else if (est == 2 && m["HEML1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "HEML with invalid date format: '" + date + "'"
					} else if m["HEML1"] == "1" {
						f.Msg = "HEML with empty date but code is 1."
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
						Type:    "hemolysis_dx",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					heml2.Source.Path = append(heml2.Source.Path, path)

					// if no duplicates, write this object to the json file and store in a slice
					if !heml2.CompareEvents(alllHEML) {

						alllHEML = append(alllHEML, heml2)
					}
				} else if (est == 2 && m["HEML2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Source.Path = append(f.Source.Path, path)
					if est == 3 {
						f.Msg = "HEML with invalid date format: '" + date + "'"
					} else if m["HEML2"] == "1" {
						f.Msg = "HEML with empty date but code is 1."
					}
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}
			}

		}
	}
}
