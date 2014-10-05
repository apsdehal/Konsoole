// Logger module used for logging various requests to a log file
package main

import (
	"os"
	"fmt"
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

