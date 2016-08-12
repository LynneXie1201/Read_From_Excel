package excel2json

// contains all the variables and types that the package needs.
var (
	allFollowUps     []followups // store followup events
	allDths          []death     // store death events
	allTIA           []te        // store tia events
	allStroke        []te        // store stroke events
	allSBE           []general   // store sbe events
	allARH           []general   // store arh events
	allLostFollowups []general   // store lost_to_followup events
	allOperation     []operation // store operation events
	allFUMI          []general   // store myocardial_infarction events
	allFUPACE        []general   // store perm_pacemaker events
	allSVD           []general   // store struct_valve_det events
	allPVL           []general   // store perivalvular_leak events
	allDVT           []general   // store deep_vein_thrombosis events
	allTHRM          []general   // store thromb_prost_valve events
	alllHEML         []general   // store hemolysis_dx events
	allLKA           []followups // store last_known_alive events
	allFix           []general   // store fix events
	codes            []string    // status codes
	nums             []int       // int values for various codes
	floats           []float64   // float points values for various codes
)

// type source
type source struct {
	Type string   `json:"type"`
	Path []string `json:"path"`
}

// error message
type errMessage struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

// operation
type operation struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PeriopID   *string      `json:"periop_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Surgeon    string       `json:"surgeon"`
	Surgeries  []string     `json:"surgeries"`
	Children   []string     `json:"children"`
	Parent     *int         `json:"parent"`
	Notes      string       `json:"notes"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// including followup and last_known_alive events
type followups struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Status     *string      `json:"status,omitempty"` // last_known_alive events don't have status field
	Notes      *string      `json:"notes"`
	Unusual    *string      `json:"unusual"`
	Plat       int          `json:"anti_platelet"`
	Coag       int          `json:"anti_coagulants"`
	PoNYHA     float64      `json:"post_op_nyha"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// death
type death struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Reason     string       `json:"reason"`
	PrmDth     int          `json:"primary_cause"`
	Operative  int          `json:"operative"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// including stroke and tia
type te struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Outcome    int          `json:"outcome"`
	Agents     int          `json:"anti_agents"`
	When       int          `json:"when,omitempty"` // only stroke events
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// type of events that share the same variables,
// including arh, lost_to_followup, myocardial_infarction, perm_pacemaker, struct_valve_det,
// perivalvular_leak, deep_vein_thrombosis, thromb_prost_valve, hemolysis_dx events
type general struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Organism   *string      `json:"organism,omitempty"` // only sbe events have
	Code       int          `json:"code,omitempty"`     // only arh events have
	Msg        string       `json:"msg,omitempty"`      // some events don't have msg field
	Notes      string       `json:"notes,omitempty"`    // some events don't have notes field
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix,omitempty"` // fix events don't need fix field
}
