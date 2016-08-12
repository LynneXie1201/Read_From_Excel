package excel2json

import (
	"excel/helper"
	"reflect"
)

// earlyDeathInfo returns a full-text meaning string of the same person's earlier death info
func (a death) earlyDeathInfo() string {
	var s, prmText, opText, dateEst string

	if a.PrmDth == 0 || a.PrmDth == -9 {
		prmText = "Not applicable"
	} else if a.PrmDth == 1 {
		prmText = "Valve-related cause"
	} else if a.PrmDth == 2 {
		prmText = "Cardiac, non valve-related cause"
	} else if a.PrmDth == 3 {
		prmText = "Non-cardiac cause"
	} else if a.PrmDth == 4 {
		prmText = "Dissection (* Used only for David op FU, otherwise PRM_DTH=3)"
	} else {
		prmText = "no invalid primary death reason avaliable"
	}

	if a.Operative == 1 {
		opText = "operative"
	} else {
		opText = "non-operative"
	}

	if a.DateEst == 1 {
		dateEst = "date estimated"
	} else {
		dateEst = "date not estimated"
	}

	s = "another record had a different date: '" + a.Date + "', '" + dateEst +
		"', '" + opText + "', primary death reason: '" + prmText + "'"
	return s
}

// CompareFollowups checks if two followup events or
// two last_known_alive events are duplicate
func (a followups) CompareFollowups(s []followups) bool {
	for i, b := range s {
		if a.Status == nil && b.Status == nil {
			if a.Coag == b.Coag && a.Date == b.Date &&
				a.PTID == b.PTID && a.Plat == b.Plat && a.PoNYHA == b.PoNYHA {
				s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
				return true
			}
		} else if a.Status != nil && b.Status != nil {
			if a.Coag == b.Coag && a.Date == b.Date &&
				a.PTID == b.PTID && a.Plat == b.Plat && a.PoNYHA == b.PoNYHA && *(a.Status) == *(b.Status) {
				s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
				return true
			}
		}
	}
	return false
}

// CompareDeath checks if two death events are duplicate
func (a *death) CompareDeath(s *[]death) bool {
	// i is the index of b
	for i, b := range *s {
		if a == &b {
			(*s)[i].Source.Path = append((*s)[i].Source.Path, a.Source.Path[0])
			return true
		} else if (*a).Date == b.Date && (*a).PTID == b.PTID && (*a).Operative == b.Operative &&
			(*a).PrmDth == b.PrmDth && (*a).Reason == b.Reason {
			(*s)[i].Source.Path = append((*s)[i].Source.Path, a.Source.Path[0])
			return true
			// same person with different death date
		} else if (*a).Date != b.Date && (*a).PTID == b.PTID && (*a).MRN == b.MRN && (*a).ResearchID == b.ResearchID {
			// how to compare 2 dates?
			if helper.DateLaterThan(b.Date, (*a).Date) {
				earlyDeath := (*a).earlyDeathInfo()
				for j, e := range b.Fix {
					if e.Field == "date" {
						(*s)[i].Fix[j].Msg += "; " + earlyDeath + ", path: " + (*a).Source.Path[0]
						return true
					}
				}

				msg := errMessage{"date", earlyDeath + ", path: " + (*a).Source.Path[0]}
				(*s)[i].Fix = append((*s)[i].Fix, msg)
				return true
			} else if helper.DateLaterThan((*a).Date, b.Date) {
				earlyDeath := b.earlyDeathInfo() // info of b

				for _, e := range b.Fix {
					if e.Field == "date" {
						earlyDeath += ", " + e.Msg
					}

				}

				for _, p := range b.Source.Path {
					earlyDeath += ", path: " + p
				}
				msg := errMessage{"date", earlyDeath}
				(*a).Fix = append((*a).Fix, msg)

				// try to delete b from the slice
				*s = append((*s)[:i], (*s)[i+1:]...)
				return false
			}

		}

	}
	return false
}

// CompareTE checks if two stroke events or two tia events are duplicate
func (a te) CompareTE(s []te) bool {
	for i, b := range s {
		if a.Agents == b.Agents && a.Date == b.Date && a.When == b.When &&
			a.Outcome == b.Outcome && a.PTID == b.PTID {
			s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}

// CompareEvents checks if two events (including SBE, lost_to_followup, FUMI,
// FUPACE, SVD, PVL, DVT, ARH, THRM, HEML, Fix) are duplicate
func (a general) CompareEvents(s []general) bool {
	for i, b := range s {
		if a.Organism == nil && b.Organism == nil {
			if a.Date == b.Date && a.PTID == b.PTID && a.Msg == b.Msg &&
				a.Code == b.Code && a.Notes == b.Notes {
				s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
				return true
			}
		} else if a.Organism != nil && b.Organism != nil {
			if *(a.Organism) == *(b.Organism) && a.Date == b.Date &&
				a.PTID == b.PTID && a.Msg == b.Msg && a.Code == b.Code && a.Notes == b.Notes {
				s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
				return true
			}
		}
	}
	return false
}

// CompareOperation checks if two operation events are duplicate
func (a operation) CompareOperation(s []operation) bool {
	for i, b := range s {
		if a.Date == b.Date && a.PTID == b.PTID && a.Notes == b.Notes &&
			a.Surgeon == b.Surgeon && reflect.DeepEqual(a.Fix, b.Fix) {
			s[i].Source.Path = append(s[i].Source.Path, a.Source.Path[0])
			return true
		}
	}
	return false
}
