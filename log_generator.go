// Logger module used for logging various requests to a log file
package main

import (
	"os"
	"fmt"
	"io/ioutil"
	"strings"
)

// Takes a request and logs it to a mentioned file
func logToFile(r Request) {
	f, err := os.OpenFile(fileToOpen, os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
	    panic(err)
	}

	defer f.Close()
	format := "%s %s %s %s %d %s %s\n"
	msg := fmt.Sprintf(format, r.DestIp, r.Time, r.Method, r.Host, r.HTTPType, r.Path, r.UserAgent )
	if _, err = f.WriteString(msg); err != nil {
	    panic(err)
	}
}

func getLogsFromFile() ([]Request, map[string]int) {
	var requestCount = map[string]int {
		"OPTIONS" : 0, "GET" : 0, "HEAD" : 0, "POST" : 0, "PUT" : 0, "DELETE" : 0, "TRACE" : 0, "CONNECT" : 0,
	}

	requests := []Request{}

	content, err := ioutil.ReadFile(fileToOpen)
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			break;
		}
		phrase := strings.SplitN(line, " ", 7)
		requestCount[phrase[2]]++
		requests = append(requests, Request{ phrase[0], phrase[1], phrase[2], phrase[3], int(phrase[4][0]), phrase[6], phrase[5] })
	}
	return requests, requestCount
}