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
	folderPath string
	errlogPath string
	jsonPath   string
)

func init() {

	// use command-line arguments to issue paths
	flag.StringVar(&folderPath, "folder", "", "a path to the folder")
	flag.StringVar(&errlogPath, "errlog", "", "a path to the errorlog file")
	flag.StringVar(&jsonPath, "json", "", "a path to the JSON file")
	flag.Parse()

}

func main() {

	fmt.Println(jsonPath)
	errLog, err := os.OpenFile(errlogPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		fmt.Println(err)
	}
	defer errLog.Close()
	e := log.New(errLog, "ERROR: ", 0)

	jsonFile, err := os.OpenFile(jsonPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	helper.CheckErr(e, err) // check for errors
	defer jsonFile.Close()

	excel2json.LoopAllFiles(e, folderPath, jsonFile)

	helper.Close(e, jsonPath)
	helper.Close(e, errlogPath)

}
