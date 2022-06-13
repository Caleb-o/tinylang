package shared

import (
	"io/ioutil"
	"log"
	"os"
)

func ReadFile(path string) string {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(src)
}

func ReportErr(msg string) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s", msg)
}

func ReportErrFatal(msg string) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s", msg)
	os.Exit(0)
}
