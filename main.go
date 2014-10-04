package main 

import (
	"os"
	"bufio"
	"fmt"
	"flag"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer
var fileToOpen string

func main() {

	fileToSave := flag.String("f", "", "Custom log file")
	flag.Parse()

	if len(*fileToSave) == 0 {
		os.Chdir("/tmp")
		f, err := os.Create("log_konsoole.txt")
		if err != nil || f == nil {
			panic(err)
		}
		fileToOpen = "log_konsoole.txt"
		defer os.Remove("log_konsoole.txt")
	} else {
		fileToOpen = *fileToSave
	}

	fmt.Println("Konsoole: HTTP Monitor")

	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)
	handleToDevice := Init()
	go InitGUI()
	for {
		pkt := handleToDevice.Next()
		if pkt != nil {
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}