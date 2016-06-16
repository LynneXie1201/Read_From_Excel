package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tealeg/xlsx"
)

// Helper functions

func changeDateFormat(x string) string {
	value := strings.Replace(x, "\\", "", -1)
	test, _ := time.Parse("02-Jan-06", value)
	return test.Format("2006-01-02")
}

// a function that writes to json files
func writeTOFile(s interface{}, name string) {

	jsonFile, err := os.Create("./" + name + ".json")
	p, _ := json.Marshal(s)
	jsonFile.Write(p)
	if err != nil {
		fmt.Println(err)
	}
	//	defer jsonFile.Close()
	jsonFile.Close()
}

// a function returns a slice of maps
func excelToSlice(excelFileName string) []map[string]string {

	xlFile, err := xlsx.OpenFile(excelFileName)
	if err != nil {
		fmt.Printf(err.Error())
	}

	var fileSlice [][][]string
	fileSlice, _ = xlsx.FileToSlice(excelFileName) // Create a file slice
	col := xlFile.Sheets[0].MaxCol                 // get the colume number
	keys := []string{}
	for k := 0; k < col; k++ {
		keys = append(keys, fileSlice[0][0][k])
	}

	slices := []map[string]string{} // each row is a map

	for _, sheet := range xlFile.Sheets {
		for _, row := range sheet.Rows {
			m := map[string]string{}
			for j, cell := range row.Cells {
				value, _ := cell.String()
				if strings.Contains(value, "\\") {
					value = changeDateFormat(value)
				}
				m[keys[j]] = value
			}
			slices = append(slices, m)
		}
	}
	return slices[1:]
}

// Create types of events

type followUp struct {
	PTID, Type, Date, Status string
	Plat, Coag, PoNYHA       int
}

type lkaDate struct {
	PTID, Type, Date string
}

type death struct {
	PTID, Type, Date string
	Code, PrmDth     int
}

type reOperation struct {
	PTID, Type, Date string
	Code             int
}

type te struct {
	PTID, Type, Date    string
	Code, Outcome, Anti int
}

type fuMI struct {
	PTID, Type, Date string
	Code             int
}

type fuPACE struct {
	PTID, Type, Date string
	Code             int
}

type sbe struct {
	PTID, Type, Date string
	Code             int
}

type svd struct {
	PTID, Type, Date string
	Code             int
}

type pvl struct {
	PTID, Type, Date string
	Code             int
}

type dvt struct {
	PTID, Type, Date string
	Code             int
}

type arh struct {
	PTID, Type, Date string
	Code             int
}

type thrm struct {
	PTID, Type, Date string
	Code             int
}

type heml struct {
	PTID, Type, Date string
	Code             int
}

var bigMap map[string][]interface{}

func main() {
	var (
		allFollowUps []followUp
		allLKA       []lkaDate
		allDths      []death
		allReOper    []reOperation
		allTE        []te
		allFuMI      []fuMI
		allFuPace    []fuPACE
		allSBE       []sbe
		allSVD       []svd
		allPVL       []pvl
		allDVT       []dvt
		allARH       []arh
		allTHRM      []thrm
		allHEML      []heml
	)
	// Returns a slice of maps from excel files
	s := excelToSlice("L:/CVDMC Students/Yilin Xie/data/follow_up.xlsx")
	for _, m := range s {
		// Event follow_ups
		if m["FU_D"] != "" {
			fU := followUp{
				PTID:   m["PTID"],
				Date:   m["FU_D"],
				Type:   "followUps",
				Status: m["Followup_STATUS"]}
			fU.Coag, _ = strconv.Atoi(m["COAG"])
			fU.PoNYHA, _ = strconv.Atoi(m["PO_NYHA"])
			fU.Plat, _ = strconv.Atoi(m["PLAT"])
			allFollowUps = append(allFollowUps, fU)
		}
		// Event LAST KNOWN ALIVE DATE
		if m["LKA_D"] != "" {
			l := lkaDate{
				PTID: m["PTID"],
				Type: "LKA_D",
				Date: m["LKA_D"]}
			allLKA = append(allLKA, l)
		}

		// Event Death
		if m["DIED"] == "1" {
			d := death{
				PTID: m["PTID"],
				Type: "DeathFU",
				Date: m["DTH_D"]}
			d.PrmDth, _ = strconv.Atoi(m["PRM_DTH"])
			d.Code, _ = strconv.Atoi(m["DIED"])
			allDths = append(allDths, d)
		}

		// Event FUREOP
		if m["FUREOP"] == "1" {
			re := reOperation{
				PTID: m["PTID"],
				Type: "FUREOP",
				Date: m["FUREOP_D"]}
			re.Code, _ = strconv.Atoi(m["FUREOP"])
			allReOper = append(allReOper, re)
		}

		// Event TE

		if m["TE1_D"] != "" {
			te1 := te{
				PTID: m["PTID"],
				Type: "TE",
				Date: m["TE1_D"]}
			te1.Code, _ = strconv.Atoi(m["TE1"])
			te1.Outcome, _ = strconv.Atoi(m["TE1_OUT"])
			te1.Anti, _ = strconv.Atoi(m["ANTI_TE1"])
			allTE = append(allTE, te1)
		}

		if m["TE2_D"] != "" {
			te2 := te{
				PTID: m["PTID"],
				Type: "TE",
				Date: m["TE2_D"]}
			te2.Code, _ = strconv.Atoi(m["TE2"])
			te2.Outcome, _ = strconv.Atoi(m["TE2_OUT"])
			te2.Anti, _ = strconv.Atoi(m["ANTI_TE2"])
			allTE = append(allTE, te2)
		}
		if m["TE3_D"] != "" {
			te3 := te{
				PTID: m["PTID"],
				Type: "TE",
				Date: m["TE3_D"]}
			te3.Code, _ = strconv.Atoi(m["TE3"])
			te3.Outcome, _ = strconv.Atoi(m["TE3_OUT"])
			te3.Anti, _ = strconv.Atoi(m["ANTI_TE3"])
			allTE = append(allTE, te3)
		}

		// Event FUMI
		if m["FUMI_D"] != "" {
			f1 := fuMI{
				PTID: m["PTID"],
				Type: "FUMI",
				Date: m["FUMI_D"]}
			f1.Code, _ = strconv.Atoi(m["FUMI"])
			allFuMI = append(allFuMI, f1)
		}

		// Event FUPACE
		if m["FUPACE_D"] != "" {
			f2 := fuPACE{
				PTID: m["PTID"],
				Type: "FUPACE",
				Date: m["FUPACE_D"]}
			f2.Code, _ = strconv.Atoi(m["FUPACE"])
			allFuPace = append(allFuPace, f2)
		}

		// Event SBE
		if m["SBE1_D"] != "" {
			sbe1 := sbe{
				Type: "SBE",
				Date: m["SBE1_D"]}
			sbe1.Code, _ = strconv.Atoi(m["SBE1"])
			allSBE = append(allSBE, sbe1)
		}

		if m["SBE2_D"] != "" {
			sbe2 := sbe{
				Type: "SBE",
				Date: m["SBE2_D"]}
			sbe2.Code, _ = strconv.Atoi(m["SBE2"])
			allSBE = append(allSBE, sbe2)
		}

		if m["SBE3_D"] != "" {
			sbe3 := sbe{
				Type: "SBE",
				Date: m["SBE3_D"]}
			sbe3.Code, _ = strconv.Atoi(m["SBE3"])
			allSBE = append(allSBE, sbe3)
		}

		// Event SVD
		if m["SVD_D"] != "" {
			s4 := svd{
				PTID: m["PTID"],
				Type: "SVD",
				Date: m["SVD_D"]}
			s4.Code, _ = strconv.Atoi(m["SVD"])
			allSVD = append(allSVD, s4)
		}
		// Event PVL
		if m["PVL1_D"] != "" {
			pvl1 := pvl{
				PTID: m["PTID"],
				Type: "PVL",
				Date: m["PVL1_D"]}
			pvl1.Code, _ = strconv.Atoi(m["PVL1"])
			allPVL = append(allPVL, pvl1)
		}

		if m["PVL2_D"] != "" {
			pvl2 := pvl{
				PTID: m["PTID"],
				Type: "PVL",
				Date: m["PVL2_D"]}
			pvl2.Code, _ = strconv.Atoi(m["PVL2"])
			allPVL = append(allPVL, pvl2)
		}

		// Event DVT
		if m["DVT_D"] != "" {
			d1 := dvt{
				PTID: m["PTID"],
				Type: "DVT",
				Date: m["DVT_D"]}
			d1.Code, _ = strconv.Atoi(m["DVT"])
			allDVT = append(allDVT, d1)
		}
		// Event ARH
		if m["ARH1_D"] != "" {
			arh1 := arh{
				PTID: m["PTID"],
				Type: "ARH",
				Date: m["ARH1_D"]}
			arh1.Code, _ = strconv.Atoi(m["ARH1"])
			allARH = append(allARH, arh1)
		}

		if m["ARH2_D"] != "" {
			arh2 := arh{
				PTID: m["PTID"],
				Type: "ARH",
				Date: m["ARH2_D"]}
			arh2.Code, _ = strconv.Atoi(m["ARH2"])
			allARH = append(allARH, arh2)

		}

		// Event THRM
		if m["THRM1_D"] != "" {

			thrm1 := thrm{
				PTID: m["PTID"],
				Type: "THRM",
				Date: m["THRM1_D"]}
			thrm1.Code, _ = strconv.Atoi(m["THRM1"])
			allTHRM = append(allTHRM, thrm1)
		}

		if m["THRM2_D"] != "" {
			thrm2 := thrm{
				PTID: m["PTID"],
				Type: "THRM",
				Date: m["THRM2_D"]}
			thrm2.Code, _ = strconv.Atoi(m["THRM2"])
			allTHRM = append(allTHRM, thrm2)

		}

		// Event HEML
		if m["HEML1_D"] != "" {
			heml1 := heml{
				PTID: m["PTID"],
				Type: "HEML",
				Date: m["HEML1_D"]}
			heml1.Code, _ = strconv.Atoi(m["HEML1"])
			allHEML = append(allHEML, heml1)
		}

		if m["HEML2_D"] != "" {
			heml2 := heml{
				PTID: m["PTID"],
				Type: "HEML",
				Date: m["HEML2_D"]}
			heml2.Code, _ = strconv.Atoi(m["HEML2"])
			allHEML = append(allHEML, heml2)
		}

	}
	writeTOFile(allTE, "TE")
	//fmt.Println(allTE)

}
