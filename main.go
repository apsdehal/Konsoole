package main 

import (
	"os"
	"bufio"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

func main () {
	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)
	handleToDevice := Init()
	InitGUI()
	for {
		pkt := handleToDevice.Next()
		if pkt != nil {
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}