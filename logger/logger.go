package logger

import (
	"encoding/csv"
	"fmt"
	"os"
)

var f = fmt.Println

func Log(text string) {
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		logFile, err = os.Create("log.txt")
		if err != nil {
			f(err)
		}
	}
	err = csv.NewWriter(logFile).WriteAll([][]string{{text}})
	if err != nil {
		f(err)
	}
	err = logFile.Close()
	if err != nil {
		f(err)
	}
}
