package main 

import (
	"github.com/miekg/pcap"
	"fmt"
	"strings"
	"net/http"
	"bytes"
	"os"
	"bufio"
	"time"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

type RequestPacket struct {
	SrcIp string
	DestIp string	
	SrcPort string
	DestPort string
	Time time.Time
	Flags string
	Request *http.Request
}

func Init() *pcap.Pcap {
	devices, err := pcap.FindAllDevs()
	if err != nil {
		fmt.Fprintf(errWriter, "[-] Error, pcap failed to iniaitilize")
	}

	if len(devices) == 0 {
		fmt.Fprintf(errWriter, "[-] No devices found, quitting!")
		os.Exit(1)
	}
	for _, x := range devices {
		fmt.Println(x.Name)
	}
	handle, err := pcap.OpenLive(devices[1].Name, 65535, true, 0)
	if err != nil {
		fmt.Fprintf(errWriter, "Konsoole: %s\n", err)
		errWriter.Flush()
	}
	return handle
}

func StringFromPacket(pkt *pcap.Packet) string {
	buffer := bytes.NewBufferString("");
	for i := uint32(0); i< pkt.Caplen; i++ {
		fmt.Fprintf(buffer, "%c", pkt.Data[i]);
	} 
	return string(buffer.Bytes())
}

func PrintPacket(r RequestPacket) {
	var s string = "%s %s %s %s %s %s"
	msg := fmt.Sprintf(s, r.SrcIp, r.DestIp, r.SrcPort, r.DestPort, r.Time, r.Flags)
	fmt.Println(msg)
}
func DecodePacket(pkt *pcap.Packet ) {
	httpMethods := [...]string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}

	IPhdr, ok1 := pkt.Headers[0].(*pcap.Iphdr)
	TCPhdr, ok2 := pkt.Headers[1].(*pcap.Tcphdr)
	fmt.Println(ok1, ok2)
	if ok1 && ok2 && TCPhdr.DestPort == 80 {
		pktString := StringFromPacket(pkt)
		fmt.Println(pktString)
		for _, method := range httpMethods {
			if strings.Contains(pktString, method) {
				unparsedReqs := strings.Split(pktString, method)
				for _, unparsed := range unparsedReqs {
					req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(method + unparsed)))
					if err == nil {
						rp := RequestPacket{string(IPhdr.SrcAddr()), string(IPhdr.DestAddr()), string(TCPhdr.SrcPort), string(TCPhdr.DestPort), pkt.Time, TCPhdr.FlagsString(), req}

						PrintPacket(rp)
					}
				}

			}
		} 
	}
}

func main () {
	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)

	c := Init()
	for pkt := c.Next(); pkt != nil; pkt = c.Next() {
		if pkt != nil {
			fmt.Println("hi")
		}
		pkt.Decode()
		DecodePacket(pkt)
	}
}