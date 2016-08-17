// Package excel2json provides functions that loop through excel files,
// read data from these files and create different events.
package excel2json

import (
	"excel/helper"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
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

				// followup event
				var coag, plat int
				var poNYHA float64
				var unusual, notes string
				// assign values to unusual and notes
				unusual = m["STATUS=O REASON"]
				notes = strings.Replace(strings.TrimSpace(m["FU NOTES"]+" "+m["NOTES"]), " ", ", ", -1)
				// validate int and float values
				coagValid := helper.CheckIntValue(&coag, m["COAG"], nums[:3])
				nyhaValid := helper.CheckFloatValue(&poNYHA, m["PO_NYHA"], floats[1:])
				platValid := helper.CheckIntValue(&plat, m["PLAT"], nums[:3])
				// create followup notes
				fuNotes := helper.FollowupNotes(S1, m["FU NOTES"], m["NOTES"], m["STATUS=O REASON"], plat, coag, poNYHA)

				// check FU_D format
				date, est := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])
				// est equals 0 or 1 means that the date format was parsed to YYYY-MM-DD
				if est == 0 || est == 1 {
					// create followup event
					fu := followups{
						PTID:    ID1,
						Date:    date,
						Type:    "followup",
						Status:  &S1,
						Plat:    plat,
						PoNYHA:  poNYHA,
						Coag:    coag,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// source: add path
					fu.Source.Path = append(fu.Source.Path, path)
					// add Notes
					if !(m["NOTES"] == "" && m["FU NOTES"] == "") {
						fu.Notes = &notes
					}
					// add Unusual
					if unusual != "" {
						fu.Unusual = &unusual
					}
					// check PTID
					if diffID {
						msg := errMessage{"patient_id", "two different PTIDs: '" + ID1 + "', '" + ID2 + "' "}
						fu.Fix = append(fu.Fix, msg)
					}
					// check STATUS
					// if true means both statuses are non-empty and not equal
					if diffStatus {
						msg := errMessage{"status", "two different Statuses: '" + S1 + "', '" + S2 + "'"}
						fu.Fix = append(fu.Fix, msg)
						// if one of the codes is D, L, N, or O and the other code is A or R, put the D, L, N or O
						if helper.StringInSlice(0, S1, codes[4:6]) && helper.StringInSlice(0, S2, codes[:4]) {
							fu.Status = &S2
						}
					}

					// validate status' values
					if *fu.Status == "" {
						fu.Status = nil
					} else if !helper.StringInSlice(0, S1, codes) {
						msg := errMessage{"code", "invalid value: '" + S1 + "'"}
						fu.Fix = append(fu.Fix, msg)
					}

					if !nyhaValid {
						msg := errMessage{"post_op_nyha", "invalid value: '" + m["PO_NYHA"] + "'"}
						fu.Fix = append(fu.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PO_NYHA", m["PO_NYHA"])
					}
					if !coagValid {
						msg := errMessage{"anti_coagulants", "invalid value: '" + m["COAG"] + "'"}
						fu.Fix = append(fu.Fix, msg)
						//errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "COAG", m["COAG"])
					}
					if !platValid {
						msg := errMessage{"anti_platelet", "invalid value: '" + m["PLAT"] + "'"}
						fu.Fix = append(fu.Fix, msg)
						//	errlog.ErrorLog(e, path, j, fU.PTID, i, fU.Type, "PLAT", m["PLAT"])
					}
					// if no duplicates, store in a slice
					if !fu.CompareFollowups(allFollowUps) {
						allFollowUps = append(allFollowUps, fu)
					}

					// est == 3 means that date has invalid format,
					// then create a fix event
				} else if est == 3 {

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

					// est equal 2: follow up date is empty
				} else if est == 2 {
					// estimate last_known_alive date
					lkaDate, lkaEst := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])

					// if last_known_alive date is valid,
					// create a last_known_alive event
					if lkaEst == 0 || lkaEst == 1 {

						lka := followups{
							PTID:    ID1,
							Type:    "last_known_alive",
							Date:    lkaDate,
							Coag:    coag,
							PoNYHA:  poNYHA,
							Plat:    plat,
							DateEst: lkaEst,
							Source:  source{Type: "followup", Path: []string{}}}

						// source: add path
						lka.Source.Path = append(lka.Source.Path, path)

						// add notes if exists
						if !(m["NOTES"] == "" && m["FU NOTES"] == "") {
							lka.Notes = &notes
						}
						// add Unusual
						if unusual != "" {
							lka.Unusual = &unusual
						}
						// check PTID
						if diffID {
							msg := errMessage{"patient_id", "two different PTIDs: '" + ID1 + "', '" + ID2 + "' "}
							lka.Fix = append(lka.Fix, msg)
						}
						// check status
						// if true means both statuses are non-empty and not equal
						if diffStatus {
							msg := errMessage{"status", "two different Statuses: '" + S1 + "', '" + S2 + "'"}
							lka.Fix = append(lka.Fix, msg)
						}
						// validate PO_NYHA
						if !nyhaValid {
							msg := errMessage{"post_op_nyha", "invalid value: '" + m["PO_NYHA"] + "'"}
							lka.Fix = append(lka.Fix, msg)
						}
						// validate COAG
						if !coagValid {
							msg := errMessage{"anti_coagulants", "invalid value: '" + m["COAG"] + "'"}
							lka.Fix = append(lka.Fix, msg)
						}

						// validate PLAT
						if !platValid {
							msg := errMessage{"anti_platelet", "invalid value: '" + m["PLAT"] + "'"}
							lka.Fix = append(lka.Fix, msg)
						}

						// if no duplicates, store in a slice
						if !lka.CompareFollowups(allLKA) {
							allLKA = append(allLKA, lka)
						}
						// if last_known_alive date has invalid date format,
						// create a fix event
					} else if lkaEst == 3 {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}

						// LKA date with invalid format
						f.Msg = "last_known_alive date with invalid format: '" + lkaDate +
							"' , and FU NOTES without date associated. Here are the followup notes: " + fuNotes

						// Source: add path
						f.Source.Path = append(f.Source.Path, path)

						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
						// else if last_known_alive date is also empty, but at least one of the followup fields is not empty,
						// create a fix event
					} else if lkaEst == 2 && (m["FU NOTES"] != "" || (coag != -9 && coag != 0) || (plat != -9 && plat != 0) ||
						(poNYHA != -9 && poNYHA != 0) || m["STATUS=O REASON"] != "" || m["NOTES"] != "") {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}

						// LKA date is missing
						f.Msg = "followup and last_known_alive events without date associated, here are the followup notes: " + fuNotes

						// Source: add path
						f.Source.Path = append(f.Source.Path, path)
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// last_known_alive event
				// estimate the value of "LKA_D"
				lkaDate, lkaEst := helper.CheckDateFormat(e, path, j, i, "LKA_Date", m["LKA_D"])

				if lkaEst == 0 || lkaEst == 1 {
					// if "LKA_D" and "FU_D" both have valid values,
					// create a last_known_alive event and set all -9s in the event
					if m["FU_D"] != "" {

						lka := followups{
							PTID:    ID1,
							Type:    "last_known_alive",
							Date:    lkaDate,
							Coag:    -9,
							PoNYHA:  -9,
							Plat:    -9,
							DateEst: lkaEst,
							Source:  source{Type: "followup", Path: []string{}}}

						// source: add path
						lka.Source.Path = append(lka.Source.Path, path)

						// if no duplicates, store in a slice
						if !lka.CompareFollowups(allLKA) {
							allLKA = append(allLKA, lka)
						}
					}

				} else if lkaEst == 3 {
					// else if "LKA_D" has invalid date and FU_D exists,
					// then create a fix event
					if m["FU_D"] != "" {

						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Msg:     "LKA Date with invalid format: '" + lkaDate + "'",
							Source:  source{Type: "followup", Path: []string{}}}

						// Source: add path
						f.Source.Path = append(f.Source.Path, path)

						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// lost_to_followup event

				// if one of the STATUS columns is “L” and the other is “D” or “N”, do not create the “lost_to_followup”,
				// otherwise, create the “lost_to_followup” event if one is “L”.
				// Date will be (in order of preference) either the “STATUS=L DATE” field, or the STATUSDATE or the FU_D if LKA_D not exists.
				// If none of those dates are available, put 1900-02-02 as the date.
				if S1 == "L" && !helper.StringInSlice(0, S2, codes[:2]) || (S2 == "L" && !helper.StringInSlice(0, S1, codes[:2])) {
					// estimate the value of "Status=L Date"
					date, est = helper.CheckDateFormat(e, path, j, i, "Status=L Date", m["STATUS=L DATE"])
					// create notes string
					notes := strings.Replace(strings.TrimSpace(m["FU NOTES"]+" "+m["NOTES"]+" "+m["STATUS=O REASON"]), " ", ", ", -1)
					// if "Status=L Date" has valid value, create a lost_to_followup event,
					// and set "Status=L Date" as the date
					if est == 0 || est == 1 {
						lost := lostFollowup{
							PTID:    ID1,
							Type:    "lost_to_followup",
							Date:    date,
							DateEst: est,
							LkaDate: &lkaDate,
							Source:  source{Type: "followup", Path: []string{}}}

						// check LKA_Date
						// LKA_Date is empty, set null in json
						if lkaEst == 2 {
							lost.LkaDate = nil
							msg := errMessage{"lka_date", "missing last_known_alive date, cannot compare with the lost_to_followup date."}
							lost.Fix = append(lost.Fix, msg)
							// invalid format, add fix message
						} else if lkaEst == 3 {
							lost.LkaDate = nil
							msg := errMessage{"lka_date", "invalid format of last_known_alive date: '" + lkaDate +
								"', cannot compare with the lost_to_followup date."}
							lost.Fix = append(lost.Fix, msg)
						} else if helper.DateLaterThan(lkaDate, date) {
							msg := errMessage{"lka_date", "conflict of 'L' status before 'LKA date' - has patient been recovered?"}
							lost.Fix = append(lost.Fix, msg)
						}

						// add notes
						if !(m["FU NOTES"] == "" && m["NOTES"] == "" && m["STATUS=O REASON"] == "") {
							lost.Notes = &notes
						}
						// Source: add path
						lost.Source.Path = append(lost.Source.Path, path)

						// if no duplicates, store in a slice
						if !lost.CompareLostFollowup(allLostFollowups) {
							allLostFollowups = append(allLostFollowups, lost)
						}

						// if “STATUS=L DATE” field is empty
					} else if est == 2 {
						// get the ".*STATUSDATE"
						var statusDate string
						var statusEst int

						for _, k := range keys {
							matched, _ := regexp.MatchString("^.*STATUSDATE$", k)
							if matched {
								statusDate, statusEst = helper.CheckDateFormat(e, path, j, i, "Status Date", m[k])
								break
							}
						}
						// estimate the value of "FU_D"
						fuDate, fuEst := helper.CheckDateFormat(e, path, j, i, "follow_up Date", m["FU_D"])

						// if STATUSDATE is valid, create a lost_to_followup event,
						// and set the STATUSDATE as the date value
						if statusEst == 0 || statusEst == 1 {
							lost := lostFollowup{
								PTID:    ID1,
								Type:    "lost_to_followup",
								Date:    statusDate,
								DateEst: statusEst,
								LkaDate: &lkaDate,
								Source:  source{Type: "followup", Path: []string{}}}

							// check LKA_Date
							// LKA_Date is empty, set null in json
							if lkaEst == 2 {
								lost.LkaDate = nil
								msg := errMessage{"lka_date", "missing last_known_alive date, cannot compare with the lost_to_followup date."}
								lost.Fix = append(lost.Fix, msg)
								// invalid format, add fix message
							} else if lkaEst == 3 {
								lost.LkaDate = nil
								msg := errMessage{"lka_date", "invalid format of last_known_alive date: '" + lkaDate +
									"', cannot compare with the lost_to_followup date."}
								lost.Fix = append(lost.Fix, msg)
							} else if helper.DateLaterThan(lkaDate, statusDate) {
								msg := errMessage{"lka_date", "conflict of 'L' status before 'LKA date' - has patient been recovered?"}
								lost.Fix = append(lost.Fix, msg)
							}

							// add Notes
							if !(m["FU NOTES"] == "" && m["NOTES"] == "" && m["STATUS=O REASON"] == "") {
								lost.Notes = &notes
							}
							// Source: add path
							lost.Source.Path = append(lost.Source.Path, path)

							// if no duplicates, store in a slice
							if !lost.CompareLostFollowup(allLostFollowups) {
								allLostFollowups = append(allLostFollowups, lost)
							}
							// else if STATUSDATE has invalid format, create a fix event with date "1900-02-02"
						} else if statusEst == 3 {

							f := general{
								PTID:    ID1,
								Type:    "fix",
								Date:    "1900-02-02",
								DateEst: 1,
								Msg:     "Invalid STATUSDATE: '" + statusDate + "', Notes: '" + notes + "'",
								Source:  source{Type: "followup", Path: []string{}}}

							// add lka_date
							if lkaEst != 2 {
								f.Msg += ", lka_date: '" + lkaDate + "'"
							}
							// add path
							f.Source.Path = append(f.Source.Path, path)
							// if no duplicates, store in a slice
							if !f.CompareEvents(allFix) {
								allFix = append(allFix, f)
							}
							// else if STATUSDATE is empty, then consider the value of followup date: FU_D
						} else if statusEst == 2 {

							if fuEst == 0 || fuEst == 1 {
								// if FU_D is valid, create a lost_to_followup event,
								// and set the FU_D as the date value

								lost := lostFollowup{
									PTID:    ID1,
									Type:    "lost_to_followup",
									Date:    fuDate,
									DateEst: fuEst,
									LkaDate: &lkaDate,
									Source:  source{Type: "followup", Path: []string{}}}

								// check LKA_Date
								// LKA_Date is empty, set null in json
								if lkaEst == 2 {
									lost.LkaDate = nil
									msg := errMessage{"lka_date", "missing last_known_alive date, cannot compare with the lost_to_followup date."}
									lost.Fix = append(lost.Fix, msg)
									// invalid format, add fix message
								} else if lkaEst == 3 {
									lost.LkaDate = nil
									msg := errMessage{"lka_date", "invalid format of last_known_alive date: '" + lkaDate +
										"', cannot compare with the lost_to_followup date."}
									lost.Fix = append(lost.Fix, msg)
								} else if helper.DateLaterThan(lkaDate, fuDate) {
									msg := errMessage{"lka_date", "conflict of 'L' status before 'LKA date' - has patient been recovered?"}
									lost.Fix = append(lost.Fix, msg)
								}

								// add Notes
								if !(m["FU NOTES"] == "" && m["NOTES"] == "" && m["STATUS=O REASON"] == "") {
									lost.Notes = &notes
								}
								// Source: add path
								lost.Source.Path = append(lost.Source.Path, path)

								// if no duplicates, store in a slice
								if !lost.CompareLostFollowup(allLostFollowups) {
									allLostFollowups = append(allLostFollowups, lost)
								}

								// if FU_D has invalid date format, create a fix event
							} else if fuEst == 3 {

								f := general{
									PTID:    ID1,
									Type:    "fix",
									Date:    "1900-02-02",
									DateEst: 1,
									Msg:     "Invalid followup date: '" + fuDate + "', Notes: '" + notes + "'",
									Source:  source{Type: "followup", Path: []string{}}}

								// add lka_date
								if lkaEst != 2 {
									f.Msg += ", lka_date: '" + lkaDate + "'"
								}

								// add path
								f.Source.Path = append(f.Source.Path, path)
								// if no duplicates, store in a slice
								if !f.CompareEvents(allFix) {
									allFix = append(allFix, f)
								}
								// else if FU_D is missing, then create a lost_to_followup event,
								// and set the date as "1900-02-02"
							} else if fuEst == 2 {

								// create a fix event
								f := general{
									PTID:    ID1,
									Type:    "fix",
									Date:    "1900-02-02",
									DateEst: 1,
									Msg:     "The status was L but there was no date to associate with it. Notes: '" + notes + "'",
									Source:  source{Type: "followup", Path: []string{}}}

								if lkaEst != 2 {
									f.Msg += ", lka_date: '" + lkaDate + "'"
								}
								// add path
								f.Source.Path = append(f.Source.Path, path)
								// if no duplicates, store in a slice
								if !f.CompareEvents(allFix) {
									allFix = append(allFix, f)
								}
							}
						}
						// if "STATUS=L DATE" has invalid date format, create a fix event
					} else {
						// create a fix event
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-02-02",
							DateEst: 1,
							Msg:     "Invalid STATUS=L DATE: '" + date + "', Notes: '" + notes + "'",
							Source:  source{Type: "followup", Path: []string{}}}

						// add lka_date
						if lkaEst != 2 {
							f.Msg += ", lka_date: '" + lkaDate + "'"
						}

						// add path
						f.Source.Path = append(f.Source.Path, path)
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// Event Death
				var operDate, operative string
				var operEst int
				// estimate death date
				date, est = helper.CheckDateFormat(e, path, j, i, "DTH_Date", m["DTH_D"])
				// get the date of surgery
				for _, k := range keys {
					matched, _ := regexp.MatchString("^.*DATEOR$", k)
					if matched {
						operDate, operEst = helper.CheckDateFormat(e, path, j, i, "DATEOR", m[k])
						break
					}
				}
				// assign operative
				if m["SURVIVAL"] == "0" {
					operative = "1"
				} else {
					operative = "0"
				}
				// death date with valid format
				if est == 0 || est == 1 {
					d := death{
						PTID:    ID1,
						Type:    "death",
						Date:    date,
						Reason:  m["REASDTH"],
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// source: add path
					d.Source.Path = append(d.Source.Path, path)
					// check Operative

					if m["SURVIVAL"] == "0" {
						d.Operative = 1
						// if date of surgery and date of death is not the same day
						if operDate != date {
							msg := errMessage{"operative", "Date of surgery is '" + operDate + "', please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					} else if m["SURVIVAL"] == "1" {
						// if date of surgery and date of death is the same day
						if operDate == date {
							msg := errMessage{"operative", "Date of surgery is '" + operDate + "', please indicate if death was operative"}
							d.Fix = append(d.Fix, msg)
						}
					}

					// if primary cause of death is not valid code
					if !helper.CheckIntValue(&d.PrmDth, m["PRM_DTH"], nums[:6]) {
						msg := errMessage{"primary_cause", "invalid value: '" + m["PRM_DTH"] + "'"}
						d.Fix = append(d.Fix, msg)
					}
					// if status is not "D" or "N"
					if S1 != "D" && S1 != "N" {
						msg := errMessage{"status", "invalid value: '" + S1 + "'"}
						d.Fix = append(d.Fix, msg)
					}

					// if no duplicates, store in a slice
					if !(&d).CompareDeath(&allDths) {
						allDths = append(allDths, d)
					}
					// est == 3 means invalid date format
				} else if est == 3 {
					//create a fix event
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// create msg
					f.Msg = "Death event with invalid date format: '" + date + "'" +
						helper.DeathNotes(m["PRM_DTH"], m["REASDTH"], operative)

					// source: add path
					f.Source.Path = append(f.Source.Path, path)
					// if no duplicates, store in a slice of the same type
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
					// else est == 2 and at least one of the following fields is not empty,
					// create a fix event
				} else if !(m["PRM_DTH"] == "0" || m["PRM_DTH"] == "") || m["REASDTH"] != "" || m["DIED"] == "1" {

					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}

					f.Msg = "Death event with no date associated" +
						helper.DeathNotes(m["PRM_DTH"], m["REASDTH"], operative)

					// add path to source
					f.Source.Path = append(f.Source.Path, path)
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event FUREOP -> Event operation

				// estimate operation date
				date, est = helper.CheckDateFormat(e, path, j, i, "FUREOP_Date", m["FUREOP_D"])
				// create operation notes
				opString := helper.OperationNotes(m["REASREOP"], m["REOPSURVIVAL"],
					m["REOPNOTES"], m["REOPSURG"], m["NONVALVE REOP"])

				// operation date with invalid format
				if est == 0 || est == 1 {
					// create an operation event
					op := operation{
						PTID:    ID1,
						Type:    "operation",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to source
					op.Source.Path = append(op.Source.Path, path)

					// check the value of REOPSURVIVAL
					var survival int
					if !helper.CheckIntValue(&survival, m["REOPSURVIVAL"], nums[:3]) {
						msg := errMessage{"survival", "invalid value: '" + m["REOPSURVIVAL"] + "'"}
						op.Fix = append(op.Fix, msg)
					}

					// add operation string to fix field
					if !(m["REASREOP"] == "" && m["REOPSURVIVAL"] == "0" && m["REOPNOTES"] == "" &&
						m["REOPSURG"] == "" && m["NONVALVE REOP"] == "") {

						msg := errMessage{"operation", opString}
						op.Fix = append(op.Fix, msg)
					}

					// if no duplicates, store in a slice
					if !op.CompareOperation(allOperation) {
						allOperation = append(allOperation, op)
					}
					// invalid date format
				} else if est == 3 {
					// create a fix event
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add Msg
					f.Msg = "Invalid REOP date format: '" + m["FUREOP_D"] + "', here is the re-operation info: " + opString
					// add path to source
					f.Source.Path = append(f.Source.Path, path)
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
					// when date is empty, other fields have at least one value,
					// create a fix event
				} else if m["FUREOP"] == "1" || m["REASREOP"] != "" || m["REOPNOTES"] != "" ||
					m["REOPSURG"] != "" || m["NONVALVE REOP"] != "" {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add msg
					f.Msg = "REOP fields without date associated, here is the re-operation info: " + opString
					// add path to source
					f.Source.Path = append(f.Source.Path, path)
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// TE1
				// estimate TE date
				date, est = helper.CheckDateFormat(e, path, j, i, "TE1_Date", m["TE1_D"])
				// TE date with valid format
				if est == 0 || est == 1 {
					// TE code is 1 or 2, create a stroke event
					if m["TE1"] == "1" || m["TE1"] == "2" {
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add fix object if TE coded 1
						if m["TE1"] == "1" {
							msg := errMessage{"stroke", "coded as ‘1’, uncertain if stroke or TIA"}
							s.Fix = append(s.Fix, msg)
						}
						// add path to source
						s.Source.Path = append(s.Source.Path, path)

						// if date of surgery has valid format,
						// compare it with the TE_D to decide the value of when:
						// if the TE date is the same day as the operation or up to 30 days after the operation, set field “when” : 1;
						// otherwise, set field “when” : 2
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// validate outcome
						if !helper.CheckIntValue(&s.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE1_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// validate anti_agents
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE1"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !s.CompareTE(allStroke) {
							allStroke = append(allStroke, s)
						}
						// TE code is 3, create a tia event
					} else if m["TE1"] == "3" {
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						t.Source.Path = append(t.Source.Path, path)

						// validate outcome value
						if !helper.CheckIntValue(&t.Outcome, m["TE1_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE1_OUT"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// validate anti_agents value
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE1"], nums[:5]) && (m["ANTI_TE1"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE1"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !t.CompareTE(allTIA) {
							allTIA = append(allTIA, t)
						}
					}
					// TE date is empty or with invalid format
				} else if est == 2 || est == 3 {
					if m["TE1"] == "1" || m["TE1"] == "2" || m["TE1"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						f.Source.Path = append(f.Source.Path, path)

						// add Msg
						if m["TE1"] == "1" {
							f.Msg = "TE was coded 1 and had no valid date associated"
						} else if m["TE1"] == "2" {
							if est == 2 {
								f.Msg = "stroke with missing date but code exists, " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"]) +
									", when: 'not applicable because of empty date'"
							} else if est == 3 {
								f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"]) +
									", when: 'not applicable because of invalid date format'"
							}
						} else if m["TE1"] == "3" {
							if est == 2 {
								f.Msg = "tia with missing date but code exists, " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])
							} else if est == 3 {
								f.Msg = "tia with invalid date format: '" + date + "', " + helper.TeNotes(m["TE1_OUT"], m["ANTI_TE1"])
							}
						}
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// TE2
				// estimate TE date
				date, est = helper.CheckDateFormat(e, path, j, i, "TE2_Date", m["TE2_D"])

				// TE date with valid format
				if est == 0 || est == 1 {
					// TE code is 1 or 2, create a stroke event
					if m["TE2"] == "1" || m["TE2"] == "2" {
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add fix object if TE coded 1
						if m["TE2"] == "1" {
							msg := errMessage{"stroke", "coded as ‘1’, uncertain if stroke or TIA"}
							s.Fix = append(s.Fix, msg)
						}

						// add path to source
						s.Source.Path = append(s.Source.Path, path)

						// if date of surgery has valid format,
						// compare it with the TE_D to decide the value of when:
						// if the TE date is the same day as the operation or up to 30 days after the operation, set field “when” : 1;
						// otherwise, set field “when” : 2
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// validate outcome
						if !helper.CheckIntValue(&s.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE2_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// validate anti_agents
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE2"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !s.CompareTE(allStroke) {
							allStroke = append(allStroke, s)
						}
						// TE code is 3, create a tia event
					} else if m["TE2"] == "3" {
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						t.Source.Path = append(t.Source.Path, path)

						// validate outcome value
						if !helper.CheckIntValue(&t.Outcome, m["TE2_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE2_OUT"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// validate anti_agents value
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE2"], nums[:5]) && (m["ANTI_TE2"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE2"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !t.CompareTE(allTIA) {
							allTIA = append(allTIA, t)
						}
					}
					// TE date is empty or with invalid format
				} else if est == 2 || est == 3 {
					if m["TE2"] == "1" || m["TE2"] == "2" || m["TE2"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						f.Source.Path = append(f.Source.Path, path)

						// add Msg
						if m["TE2"] == "1" {
							f.Msg = "TE was coded 1 and had no valid date associated"
						} else if m["TE2"] == "2" {
							if est == 2 {
								f.Msg = "stroke with missing date but code exists, " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"]) +
									", when: 'not applicable because of empty date'"
							} else if est == 3 {
								f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"]) +
									", when: 'not applicable because of invalid date format'"
							}
						} else if m["TE2"] == "3" {
							if est == 2 {
								f.Msg = "tia with missing date but code exists, " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])
							} else if est == 3 {
								f.Msg = "tia with invalid date format: '" + date + "', " + helper.TeNotes(m["TE2_OUT"], m["ANTI_TE2"])
							}
						}
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// TE3
				// estimate TE date
				date, est = helper.CheckDateFormat(e, path, j, i, "TE3_Date", m["TE3_D"])
				// TE date with valid format
				if est == 0 || est == 1 {
					// TE code is 1 or 2, create a stroke event
					if m["TE3"] == "1" || m["TE3"] == "2" {
						s := te{
							PTID:    ID1,
							Type:    "stroke",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add fix object if TE coded 1
						if m["TE3"] == "1" {
							msg := errMessage{"stroke", "coded as ‘1’, uncertain if stroke or TIA"}
							s.Fix = append(s.Fix, msg)
						}

						// add path to source
						s.Source.Path = append(s.Source.Path, path)

						// if date of surgery has valid format,
						// compare it with the TE_D to decide the value of when:
						// if the TE date is the same day as the operation or up to 30 days after the operation, set field “when” : 1;
						// otherwise, set field “when” : 2
						if operEst == 0 || operEst == 1 {
							s.When = helper.CompareDates(e, date, operDate)
						} else {
							msg := errMessage{"when", "cannot compare with DATEOR, it is empty or has different name."}
							s.Fix = append(s.Fix, msg)
							e.Println(path, "sheet:", j+1, "row:", i+2, "INFO: DATEOR is empty or has different name.")
						}

						// validate outcome
						if !helper.CheckIntValue(&s.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE3_OUT"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// validate anti_agents
						if !helper.CheckIntValue(&s.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE3"] + "'"}
							s.Fix = append(s.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !s.CompareTE(allStroke) {
							allStroke = append(allStroke, s)
						}
						// TE code is 3, create a tia event
					} else if m["TE3"] == "3" {
						t := te{
							PTID:    ID1,
							Type:    "tia",
							Date:    date,
							DateEst: est,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						t.Source.Path = append(t.Source.Path, path)

						// validate outcome value
						if !helper.CheckIntValue(&t.Outcome, m["TE3_OUT"], nums[:5]) {
							msg := errMessage{"outcome", "invalid value: '" + m["TE3_OUT"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// validate anti_agents value
						if !helper.CheckIntValue(&t.Agents, m["ANTI_TE3"], nums[:5]) && (m["ANTI_TE3"] != "8") {
							msg := errMessage{"anti_agents", "invalid value: '" + m["ANTI_TE3"] + "'"}
							t.Fix = append(t.Fix, msg)
						}
						// if no duplicates, store in a slice
						if !t.CompareTE(allTIA) {
							allTIA = append(allTIA, t)
						}
					}
					// TE date is empty or with invalid format
				} else if est == 2 || est == 3 {
					if m["TE3"] == "1" || m["TE3"] == "2" || m["TE3"] == "3" {
						f := general{
							PTID:    ID1,
							Type:    "fix",
							Date:    "1900-01-01",
							DateEst: 1,
							Source:  source{Type: "followup", Path: []string{}}}

						// add path to source
						f.Source.Path = append(f.Source.Path, path)

						// add Msg
						if m["TE3"] == "1" {
							f.Msg = "TE was coded 1 and had no valid date associated."
						} else if m["TE3"] == "2" {
							if est == 2 {
								f.Msg = "stroke with missing date but code exists, " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"]) +
									", when: 'not applicable because of empty date'"
							} else if est == 3 {
								f.Msg = "stroke with invalid date format: '" + date + "', " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"]) +
									", when: 'not applicable because of invalid date format'"
							}
						} else if m["TE3"] == "3" {
							if est == 2 {
								f.Msg = "tia with missing date but code exists, " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])
							} else if est == 3 {
								f.Msg = "tia with invalid date format: '" + date + "', " + helper.TeNotes(m["TE3_OUT"], m["ANTI_TE3"])
							}
						}
						// if no duplicates, store in a slice
						if !f.CompareEvents(allFix) {
							allFix = append(allFix, f)
						}
					}
				}

				// Event FUMI
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "FUMI_Date", m["FUMI_D"])
				// FUMI date has valid format
				if est == 0 || est == 1 {
					// create a myocardial_infarction event
					mi := general{
						PTID:    ID1,
						Type:    "myocardial_infarction",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to source
					mi.Source.Path = append(mi.Source.Path, path)

					// if no duplicates, store in a slice
					if !mi.CompareEvents(allFUMI) {
						allFUMI = append(allFUMI, mi)
					}
					// invalid date format or date is empty,
					// create a fix event
				} else if est == 3 || (est == 2 && m["FUMI"] == "1") {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "FUMI with invalid date format: '" + date + "'"
					} else {
						f.Msg = "FUMI with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event FUPACE

				// estimate date's value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "FUPACE_Date", m["FUPACE_D"])
				// date has valid format,
				// create a perm_pacemaker event
				if est == 0 || est == 1 {
					pace := general{
						PTID:    ID1,
						Type:    "perm_pacemaker",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// add path to the source
					pace.Source.Path = append(pace.Source.Path, path)

					// if no duplicates, store in a slice
					if !pace.CompareEvents(allFUPACE) {
						allFUPACE = append(allFUPACE, pace)
					}
					// if date is empty or has invalid format,
					// create a fix event
				} else if (est == 2 && m["FUPACE"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "FUPACE with invalid date format: '" + date + "'"
					} else if m["FUPACE"] == "1" {
						f.Msg = "FUPACE with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event SBE
				// estimate date's value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE1_Date", m["SBE1_D"])
				// get value for Organism
				ORGANISM := m["SBE1 ORGANISM"]
				organism := m["SBE1 organism"]
				// if date has valid format, create a sbe event
				if est == 0 || est == 1 {
					sbe1 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// add path to the source
					sbe1.Source.Path = append(sbe1.Source.Path, path)

					// assign value to Organism
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

					// if no duplicates, store in a slice
					if !sbe1.CompareEvents(allSBE) {
						allSBE = append(allSBE, sbe1)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["SBE1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// check for Organism
					if ORGANISM != "" {
						organism = ORGANISM
					}
					// add Msg
					if est == 3 {
						f.Msg = "SBE with invalid date format: '" + date + "', organism: '" + organism + "'"
					} else if m["SBE1"] == "1" {
						f.Msg = "SBE with no date but code is 1, code: '" + m["SBE1"] + "', organism: '" + organism + "'"
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// SBE2
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE2_Date", m["SBE2_D"])
				// get value for Organism
				ORGANISM = m["SBE2 ORGANISM"]
				organism = m["SBE2 organism"]
				// if date has valid format, create a sbe event
				if est == 0 || est == 1 {
					sbe2 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					sbe2.Source.Path = append(sbe2.Source.Path, path)

					// assign value to Organism
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

					// if no duplicates, store in a slice
					if !sbe2.CompareEvents(allSBE) {
						allSBE = append(allSBE, sbe2)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["SBE2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// check for Organism
					if ORGANISM != "" {
						organism = ORGANISM
					}
					// add Msg
					if est == 3 {
						f.Msg = "SBE with invalid date format: '" + date + "', organism: '" + organism + "'"
					} else if m["SBE2"] == "1" {
						f.Msg = "SBE with no date but code is 1, organism: '" + organism + "'"
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// SBE3
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "SBE3_Date", m["SBE3_D"])
				// get value for Organism
				ORGANISM = m["SBE3 ORGANISM"]
				organism = m["SBE3 organism"]
				// if date has valid format, create a sbe event
				if est == 0 || est == 1 {
					sbe3 := general{
						PTID:    ID1,
						Type:    "sbe",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					sbe3.Source.Path = append(sbe3.Source.Path, path)

					// assign value to Organism
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

					// if no duplicates, store in a slice
					if !sbe3.CompareEvents(allSBE) {
						allSBE = append(allSBE, sbe3)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["SBE3"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// check for Organism
					if ORGANISM != "" {
						organism = ORGANISM
					}
					// add Msg
					if est == 3 {
						f.Msg = "SBE with invalid date format: '" + date + "',  organism: '" + organism + "'"
					} else if m["SBE3"] == "1" {
						f.Msg = "SBE with no date but code is 1, organism: '" + organism + "'"
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event SVD
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "SVD_Date", m["SVD_D"])
				// if date has valid format, create a struct_valve_det event
				if est == 0 || est == 1 {
					svd := general{
						PTID:    ID1,
						Type:    "struct_valve_det",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// Source: add path
					svd.Source.Path = append(svd.Source.Path, path)
					// if no duplicates, store in a slice
					if !svd.CompareEvents(allSVD) {
						allSVD = append(allSVD, svd)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["SVD"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)

					// add Msg
					if est == 3 {
						f.Msg = "SVD with invalid date format: '" + date + "'"
					} else if m["SVD"] == "1" {
						f.Msg = "SVD with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event PVL
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "PVL1_Date", m["PVL1_D"])
				// if date has valid format, create a perivalvular_leak event
				if est == 0 || est == 1 {
					pvl1 := general{
						PTID:    ID1,
						Type:    "perivalvular_leak",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					pvl1.Source.Path = append(pvl1.Source.Path, path)

					// if no duplicates, store in a slice
					if !pvl1.CompareEvents(allPVL) {
						allPVL = append(allPVL, pvl1)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["PVL1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "PVL with invalid date format: '" + date + "'"
					} else if m["PVL1"] == "1" {
						f.Msg = "PVL with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// PVL2
				// estimate date value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "PVL2_Date", m["PVL2_D"])
				// if date has valid format, create a perivalvular_leak event
				if est == 0 || est == 1 {
					pvl2 := general{
						PTID:    ID1,
						Type:    "perivalvular_leak",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// add path to the source
					pvl2.Source.Path = append(pvl2.Source.Path, path)
					// if no duplicates, store in a slice
					if !pvl2.CompareEvents(allPVL) {
						allPVL = append(allPVL, pvl2)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["PVL2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "PVL with invalid date format: '" + date + "'"
					} else if m["PVL2"] == "1" {
						f.Msg = "PVL with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event DVT
				// estimate value and format of the date
				date, est = helper.CheckDateFormat(e, path, j, i, "DVT_Date", m["DVT_D"])
				// if date has valid format, create a deep_vein_thrombosis event
				if est == 0 || est == 1 {
					dvt := general{
						PTID:    ID1,
						Type:    "deep_vein_thrombosis",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to source
					dvt.Source.Path = append(dvt.Source.Path, path)
					// if no duplicates, store in a slice
					if !dvt.CompareEvents(allDVT) {
						allDVT = append(allDVT, dvt)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["DVT"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "DVT with invalid date format: '" + date + "'"
					} else if m["DVT"] == "1" {
						f.Msg = "DVT with no date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event ARH
				// estimate the value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH1_Date", m["ARH1_D"])
				// if date has valid format, create an arh event
				if est == 0 || est == 1 {
					arh1 := general{
						PTID:    ID1,
						Type:    "arh",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					arh1.Source.Path = append(arh1.Source.Path, path)

					// validate arh code
					if !helper.CheckIntValue(&arh1.Code, m["ARH1"], nums[:]) {
						msg := errMessage{"code", "invalid value: '" + m["ARH1"] + "'"}
						arh1.Fix = append(arh1.Fix, msg)
					}
					// if no duplicates, store in a slice
					if !arh1.CompareEvents(allARH) {
						allARH = append(allARH, arh1)
					}
					// if date has invalid format or is empty but other fields have values,
					// create a fix event
				} else if (est == 2 && m["ARH1"] != "0" && m["ARH1"] != "") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "ARH with invalid date format: '" + date + "', " + helper.ArhCode(m["ARH1"])
					} else if m["ARH1"] != "0" && m["ARH1"] != "" {
						f.Msg = "ARH with no date but code is not 0 or empty, " + helper.ArhCode(m["ARH1"])
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// ARH2
				// estimate the format and value
				date, est = helper.CheckDateFormat(e, path, j, i, "ARH2_Date", m["ARH2_D"])
				// if date has valid format, create an arh event
				if est == 0 || est == 1 {
					arh2 := general{
						PTID:    ID1,
						Type:    "arh",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					arh2.Source.Path = append(arh2.Source.Path, path)

					// validate arh code
					if !helper.CheckIntValue(&arh2.Code, m["ARH2"], nums[:]) {
						msg := errMessage{"code", "invalid value: '" + m["ARH2"] + "'"}
						arh2.Fix = append(arh2.Fix, msg)
					}
					// if no duplicates, store in a slice
					if !arh2.CompareEvents(allARH) {
						allARH = append(allARH, arh2)
					}
					// if date is empty or has invalid format
				} else if (est == 2 && m["ARH2"] != "0" && m["ARH2"] != "") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "ARH with invalid date format: '" + date + "', " + helper.ArhCode(m["ARH2"])
					} else if m["ARH2"] != "0" && m["ARH2"] != "" {
						f.Msg = "ARH with no date but code is not 0 or empty, " + helper.ArhCode(m["ARH2"])
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event THRM
				// estimate the value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "THRM1_Date", m["THRM1_D"])
				// if date has valid format, create a thromb_prost_valve event
				if est == 0 || est == 1 {
					thrm1 := general{
						PTID:    ID1,
						Type:    "thromb_prost_valve",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// add path to the source
					thrm1.Source.Path = append(thrm1.Source.Path, path)

					// if no duplicates, store in a slice
					if !thrm1.CompareEvents(allTHRM) {
						allTHRM = append(allTHRM, thrm1)
					}
					// if date has invalid format or is empty, create a fix event
				} else if (est == 2 && m["THRM1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source field
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "THRM with invalid date format: '" + date + "'"
					} else if m["THRM1"] == "1" {
						f.Msg = "THRM with empty date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// THRM2
				// estimate the value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "THRM2_Date", m["THRM2_D"])
				// if date has valid format, create a thromb_prost_valve event
				if est == 0 || est == 1 {
					thrm2 := general{
						PTID:    ID1,
						Type:    "thromb_prost_valve",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}

					// add path to the source
					thrm2.Source.Path = append(thrm2.Source.Path, path)

					// if no duplicates, store in a slice
					if !thrm2.CompareEvents(allTHRM) {
						allTHRM = append(allTHRM, thrm2)
					}
					// if date has invalid format or is empty, create a fix event
				} else if (est == 2 && m["THRM2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source field
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "THRM with invalid date format: '" + date + "'"
					} else if m["THRM2"] == "1" {
						f.Msg = "THRM with empty date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				// Event HEML
				// estimate value and format
				date, est = helper.CheckDateFormat(e, path, j, i, "HEML1_Date", m["HEML1_D"])
				// if date has valid format, create a hemolysis_dx event
				if est == 0 || est == 1 {
					heml1 := general{
						PTID:    ID1,
						Type:    "hemolysis_dx",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					heml1.Source.Path = append(heml1.Source.Path, path)

					// if no duplicates, store in a slice
					if !heml1.CompareEvents(alllHEML) {
						alllHEML = append(alllHEML, heml1)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["HEML1"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "HEML with invalid date format: '" + date + "'"
					} else if m["HEML1"] == "1" {
						f.Msg = "HEML with empty date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}

				//HEML2

				// estimate value and format of the date
				date, est = helper.CheckDateFormat(e, path, j, i, "HEML2_Date", m["HEML2_D"])
				// if date has a valid format, create a hemolysis_dx event
				if est == 0 || est == 1 {
					heml2 := general{
						PTID:    ID1,
						Type:    "hemolysis_dx",
						Date:    date,
						DateEst: est,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					heml2.Source.Path = append(heml2.Source.Path, path)
					// if no duplicates, store in a slice
					if !heml2.CompareEvents(alllHEML) {
						alllHEML = append(alllHEML, heml2)
					}
					// if date is empty or has invalid format, create a fix event
				} else if (est == 2 && m["HEML2"] == "1") || est == 3 {
					f := general{
						PTID:    ID1,
						Type:    "fix",
						Date:    "1900-01-01",
						DateEst: 1,
						Source:  source{Type: "followup", Path: []string{}}}
					// add path to the source
					f.Source.Path = append(f.Source.Path, path)
					// add Msg
					if est == 3 {
						f.Msg = "HEML with invalid date format: '" + date + "'"
					} else if m["HEML2"] == "1" {
						f.Msg = "HEML with empty date but code is 1."
					}
					// if no duplicates, store in a slice
					if !f.CompareEvents(allFix) {
						allFix = append(allFix, f)
					}
				}
			}
		}
	}
}
