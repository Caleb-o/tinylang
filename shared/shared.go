package shared

import (
	"io/fs"
	"io/ioutil"
	"log"
	"os"
)

func WriteFile(path string, contents string) bool {
	return ioutil.WriteFile(path, []byte(contents), fs.ModeAppend) == nil
}

func ReadFile(path string) string {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		log.Fatal(err)
	}
	return string(src)
}

func ReadFileErr(path string) (string, bool) {
	src, err := ioutil.ReadFile(path)
	if err != nil {
		return "", false
	}
	return string(src), true
}

func ReportErr(msg string) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s", msg)
}

func ReportErrFatal(msg string) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s", msg)
	os.Exit(0)
}

func SameFile(path1 string, path2 string) bool {
	h1, err := os.Open(path1)

	if err != nil {
		return false
	}
	defer h1.Close()

	h2, err := os.Open(path2)

	if err != nil {
		return false
	}
	defer h2.Close()

	s1, _ := h1.Stat()
	s2, _ := h2.Stat()

	return os.SameFile(s1, s2)
}
