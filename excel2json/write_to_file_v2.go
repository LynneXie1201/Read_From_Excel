package excel2json

import (
	"excel/helper"
	"os"
)

// WriteToJSON writes from different types of slice to JSON objects
func WriteToJSON(jsonFile *os.File, allARH []general, allDVT []general, allDths []death, allFUMI []general,
	allFUPACE []general, allFix []general, allFollowUps []followups, allLKA []followups,
	allOperation []operation, allPVL []general, allSBE []general, allSVD []general,
	allStroke []te, allTHRM []general, allTIA []te, alllHEML []general, allLostFollowups []lostFollowup) {

	// followup events
	for _, o := range allFollowUps {
		helper.WriteTOFile(jsonFile, o)
	}
	// last known alive date events
	for _, o := range allLKA {
		helper.WriteTOFile(jsonFile, o)
	}
	// sbe events
	for _, o := range allSBE {
		helper.WriteTOFile(jsonFile, o)
	}
	// myocardial_infarction
	for _, o := range allFUMI {
		helper.WriteTOFile(jsonFile, o)
	}
	// perm_pacemaker
	for _, o := range allFUPACE {
		helper.WriteTOFile(jsonFile, o)
	}
	// deep_vein_thrombosis
	for _, o := range allDVT {
		helper.WriteTOFile(jsonFile, o)
	}
	// arh events
	for _, o := range allARH {
		helper.WriteTOFile(jsonFile, o)
	}
	//  T.I.A.
	for _, o := range allTIA {
		helper.WriteTOFile(jsonFile, o)
	}
	// fix events
	for _, o := range allFix {
		helper.WriteTOFile(jsonFile, o)
	}
	// operation events
	for _, o := range allOperation {
		helper.WriteTOFile(jsonFile, o)
	}
	// death events
	for _, o := range allDths {
		helper.WriteTOFile(jsonFile, o)
	}
	// thromb_prost_valve
	for _, o := range allTHRM {
		helper.WriteTOFile(jsonFile, o)
	}
	// hemolysis_dx
	for _, o := range alllHEML {
		helper.WriteTOFile(jsonFile, o)
	}
	// struct_valve_det
	for _, o := range allSVD {
		helper.WriteTOFile(jsonFile, o)
	}
	// perivalvular_leak
	for _, o := range allPVL {
		helper.WriteTOFile(jsonFile, o)
	}
	// stroke events
	for _, o := range allStroke {
		helper.WriteTOFile(jsonFile, o)
	}
	// lost to followup events
	for _, o := range allLostFollowups {
		helper.WriteTOFile(jsonFile, o)
	}
}
