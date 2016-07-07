package main

import (
	"excel/excel2json"
	"excel/helper"
	"flag"
	"fmt"
	"log"
	"os"
)

var (
	folderPath string // path to the folder of excel files
	errlogPath string // path to the error log file
	jsonPath   string // path to the JSON file
)

func init() {

	// use command-line flags to assign paths
	flag.StringVar(&folderPath, "folder", "", "a path to the folder")
	flag.StringVar(&errlogPath, "errlog", "", "a path to the errorlog file")
	flag.StringVar(&jsonPath, "json", "", "a path to the JSON file")
	flag.Parse()

}

func main() {
	// open an error log file for writing and appending
	errLog, err := os.OpenFile(errlogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer errLog.Close()
	// create a new logger e
	e := log.New(errLog, "ERROR: ", 0)

	// open a JSON file for further writing and appending
	jsonFile, err := os.OpenFile(jsonPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	helper.CheckErr(e, err) // check for errors
	defer jsonFile.Close()

	// loop through the excel files from folderPath,
	// using logger e to record error messages, and writes events to jsonFile
	excel2json.LoopAllFiles(e, folderPath, jsonFile)

	// close the JSON file and error logs
	helper.Close(e, jsonPath)
	helper.Close(e, errlogPath)

}
