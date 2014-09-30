package main 

import (
	"github.com/miekg/pcap"
	"fmt"
	"strings"
	"http"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

type RequestPacket struct {
	SrcIp string
	DestIp string	
	SrcPort string
	DestPort string
	Time string
	Flags string
	Request *http.Request
}

func Init() *pcap.Pcap {
	devices, err := pcap.Findalldevs()
	if err != nil {
		fmt.Fprintf(errWriter, "[-] Error, pcap failed to iniaitilize")
	}

	if len(devices) == 0 {
		fmt.Fprintf(errWriter, "[-] No devices found, quitting!")
		os.Exit(1)
	}

	handle, err := pcap.Openlive(devices[0].Name, 65535, true, 0)
	if err != nil {
		fmt.Fprintf(errWriter, "Konsoole: %s\n", err)
		errWriter.Flush()
	}
	return handle
}

func main () {
}