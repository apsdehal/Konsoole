package main 

import (
	// "github.com/miekg/pcap"
	"github.com/akrennmair/gopcap"
	"fmt"
	"strings"
	"net/http"
	// "encoding/binary"
	// "strconv"
	"bytes"
	"os"
	"bufio"
	// "time"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

type RequestPacket struct {
	SrcIp     string
	DestIp    string
	SrcPort   uint16
	DestPort  uint16
	Time      string
	Method 	  string
	Flags     string
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
	// for _, x := range devices {
	// 	fmt.Println(x.Name)
	// }
	handle, err := pcap.Openlive(devices[1].Name, 65535, true, 0)
	if err != nil {
		fmt.Fprintf(errWriter, "Konsoole: %s\n", err)
		errWriter.Flush()
	}
	return handle
}

func StringFromPacket(pkt *pcap.Packet) string {
	buf := bytes.NewBufferString("")
	// payload := binary.BigEndian.Uint16(pkt.Payload[0:2])
	// fmt.Println(pkt.Payload)
	for i := uint32(0); int(i) < len(pkt.Data); i++ {
		if pkt.Data[i] >= 65 && pkt.Data[i] <=122 {
			fmt.Fprintf(buf, "%c", pkt.Data[i])
		}
	}
	return string(buf.Bytes())
}

func PrintPacket(r RequestPacket) {
	// var s string = "%s %s %d %d %s %s"
	// msg := fmt.Sprintf(s, r.SrcIp, r.DestIp, r.SrcPort, r.DestPort, r.Time, r.Flags)
	fmt.Println(r.Method, r.SrcIp, r.DestIp, r.SrcPort, r.DestPort)
}
func DecodePacket(pkt *pcap.Packet ) {
	httpMethods := [...]string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}

	if len(pkt.Headers) == 2 {
		ip, ok1  := pkt.Headers[0].(*pcap.Iphdr)
		tcp, ok2 := pkt.Headers[1].(*pcap.Tcphdr)
		if ok1 && ok2 && tcp.DestPort == 80 {
			pktString := StringFromPacket(pkt)
			for _, method := range httpMethods {
				if strings.Contains(pktString, method) {
					rp := RequestPacket{ ip.SrcAddr(), ip.DestAddr(), tcp.SrcPort, tcp.DestPort, "yo", method, tcp.FlagsString() }
					PrintPacket(rp)
					unparsedReqs := strings.Split(pktString, method)
					for _, unparsed := range unparsedReqs {
							req, err := http.ReadRequest(bufio.NewReader(strings.NewReader(method + unparsed)))
						// if err == nil {
							rp := RequestPacket{ ip.SrcAddr(), ip.DestAddr(), tcp.SrcPort, tcp.DestPort, "yo", method, tcp.FlagsString() }
							PrintPacket(rp)
						// }	
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
	for pkt := c.Next();; pkt = c.Next() {
		if pkt != nil {
			// fmt.Println(pkt.Data)
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}