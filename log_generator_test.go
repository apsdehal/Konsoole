package main 

import (
	"testing"
	"os"
)

func TestlogToFile(t *testing.T) {
	r := Request{"192.168.2.3", "00:00:00", "GET", "konsoole.com", 1, "Gecko", "/github.com"}
	fileToOpen = "log.txt"
	logToFile(r)
	requests , requestCount:= getLogsFromFile()
	request := requests[0]
	if len(requests) != 1 {
		t.Errorf("Failed on length of requests")
	}
	if requestCount["GET"] != 1 {
		t.Errorf("Failed on request count")
	}
	if request != r {
		t.Errorf("Failed on back request")
	}
	os.Remove("log.txt")	
}