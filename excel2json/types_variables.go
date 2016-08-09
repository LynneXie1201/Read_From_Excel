package excel2json

var (
	allFollowUps     []followUp // store followUp events
	allDths          []death    // store death events
	allTIA           []tia      // store TIA events
	allStroke        []stroke   // store stroke events
	allSBE           []sbe      // store SBE events
	allARH           []arh
	allLostFollowups []lostFollowup
	allOperation     []operation
	allFUMI          []general
	allFUPACE        []general
	allSVD           []general
	allPVL           []general
	allDVT           []general
	allTHRM          []general
	alllHEML         []general
	allLKA           []lka
	allFix           []general
	codes            []string // status codes
	nums             []int    // numerical values for various codes
	floats           []float64
)

type source struct {
	Type string   `json:"type"`
	Path []string `json:"path"`
}

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

type errMessage struct {
	Field string `json:"field"`
	Msg   string `json:"msg"`
}

// FollowUp is follow up event
type followUp struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Status     *string      `json:"status"`
	Notes      string       `json:"notes"`
	Unusual    string       `json:"unusual"`
	Plat       int          `json:"anti_platelet"`
	Coag       int          `json:"anti_coagulants"`
	PoNYHA     float64      `json:"post_op_nyha"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// FollowUp is follow up event
type lostFollowup struct {
	Type       string
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Notes      string       `json:"notes"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

type lka struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Notes      string       `json:"notes"`
	Unusual    string       `json:"unusual"`
	Plat       int          `json:"anti_platelet"`
	Coag       int          `json:"anti_coagulants"`
	PoNYHA     float64      `json:"post_op_nyha"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// death event
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

// stroke event
type stroke struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Outcome    int          `json:"outcome"`
	Agents     int          `json:"anti_agents"`
	When       int          `json:"when"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// TIA event
type tia struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Outcome    int          `json:"outcome"`
	Agents     int          `json:"anti_agents"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// SBE event
type sbe struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Organism   *string      `json:"organism"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}

// type of events that share the same variables
type general struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Msg        string       `json:"msg,omitempty"` // some events don't have notes field
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix,omitempty"` // Fix events don't need fix field
}

// type of events that share the same variables
type arh struct {
	Type       string       `json:"type"`
	MRN        string       `json:"mrn"`
	ResearchID string       `json:"research_id"`
	PTID       string       `json:"patient_id"`
	Date       string       `json:"date"`
	DateEst    int          `json:"date_est"`
	Code       int          `json:"code"`
	Source     source       `json:"source"`
	Fix        []errMessage `json:"fix"`
}
