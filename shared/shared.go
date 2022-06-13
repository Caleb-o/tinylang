package shared

import (
	"io/ioutil"
	"log"
	"os"
	"tiny/lexer"
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

func ReportErrToken(msg string, token *lexer.Token) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s [%d:%d]", msg, token.Line, token.Column)
}

func ReportErrFatal(msg string) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s", msg)
	os.Exit(0)
}

func ReportErrTokenFatal(msg string, token *lexer.Token) {
	log.Printf("\u001b[31;1mError:\u001b[0m %s [%d:%d]", msg, token.Line, token.Column)
	os.Exit(0)
}
