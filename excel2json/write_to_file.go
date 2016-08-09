package excel2json

import (
	"excel/helper"
	"os"
)

// WriteToJSON writes from slices to JSON objects
func WriteToJSON(jsonFile *os.File, allARH []arh, allDVT []general, allDths []death, allFUMI []general,
	allFUPACE []general, allFix []general, allFollowUps []followUp, allLKA []lka,
	allOperation []operation, allPVL []general, allSBE []sbe, allSVD []general,
	allStroke []stroke, allTHRM []general, allTIA []tia, alllHEML []general, allLostFollowups []lostFollowup) {

	for _, o := range allFollowUps {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allLKA {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allSBE {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allFUMI {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allFUPACE {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allDVT {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allARH {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allTIA {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allFix {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allOperation {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allDths {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allTHRM {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range alllHEML {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allSVD {
		helper.WriteTOFile(jsonFile, o)
	}
	for _, o := range allPVL {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allStroke {
		helper.WriteTOFile(jsonFile, o)
	}

	for _, o := range allLostFollowups {
		helper.WriteTOFile(jsonFile, o)
	}
}
