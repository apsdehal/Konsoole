// Parser module used for parsing of HTTP packets 
package main 

import (
	"github.com/akrennmair/gopcap"
	"fmt"
	"strings"
	"bytes"
	"regexp"
	"errors"
	"time"
	"os"
)

// Defines a typical request structure
type Request struct {
	DestIp    string
	Time      string
	Method 	  string
	Host 	  string
	HTTPType  int
	UserAgent string
	Path 	  string
}

// Initializes the network interface by finding all the available devices
// displays them to user and finally selects one of them as per the user
func Init() *pcap.Pcap {

	devices := getDevices()

	if len(devices) == 0 {
		fmt.Fprintf(errWriter, "[-] No devices found, quitting!")
		os.Exit(1)
	}

	fmt.Println("Select one of the devices:")
	var i int = 1
	for _, x := range devices {
		fmt.Println(i, x.Name)
		i++
	}

	var index int

	fmt.Scanf("%d", &index)

	handle := getHandle(devices, index)

	return handle
}

func getDevices() (devices []pcap.Interface) {
	devices, err := pcap.Findalldevs()
	if err != nil {
		fmt.Fprintf(errWriter, "[-] Error, pcap failed to iniaitilize")
	}
	return devices
}

func getHandle(devices []pcap.Interface, index int) *pcap.Pcap {
	handle, err := pcap.Openlive(devices[index-1].Name, 65535, true, 0)
	if err != nil {
		fmt.Fprintf(errWriter, "Konsoole: %s\n", err)
		errWriter.Flush()
	}
	return handle
} 
// Converts packet to a proper string and returns it
func StringFromPacket(pkt *pcap.Packet) string {
	buf := bytes.NewBufferString("")
	for i := uint32(0); int(i) < len(pkt.Data); i++ {
		fmt.Fprintf(buf, "%c", pkt.Data[i])
	}
	return string(buf.Bytes())
}

// Parse the string from request and decodes it into a structured request using regex
func ParsePayload(pktString string, ip *pcap.Iphdr, tcp *pcap.Tcphdr, method string) (Request, error) {
	DestIp   := ip.DestAddr()

	reqRegex, _		  := regexp.Compile("/(.+)\\s+HTTP/1.([0-1])\\s+")
	hostRegex, _      := regexp.Compile("Host: (.+)\\s+")
	useragentRegex, _ := regexp.Compile("User-Agent: (.+)")

	host	  := hostRegex.FindStringSubmatch(pktString)[1] 
	useragent := useragentRegex.FindStringSubmatch(pktString)[1] 
	req  	  := reqRegex.FindStringSubmatch(pktString) 
	
	if len(req) == 0 {
		return Request{}, errors.New("not correct")
	}
	
	path 	    := req[1]
	httpType    := req[2]

	rp := Request{ 
					DestIp, 
					time.Now().Format(time.RFC3339), 
					method,
					host, 
					int(httpType[0]) - '0', 
					useragent, 
					path,
				}
	return rp, nil
}

// Decodes a packet and checks for IP, TCP , 80 Destination port and finally HTTP request method to find a valid HTTP request
// If found uses a go routine to reinitialize the GUI
func DecodePacket(pkt *pcap.Packet) {
	httpMethods := [...]string{"OPTIONS", "GET", "HEAD", "POST", "PUT", "DELETE", "TRACE", "CONNECT"}

	if len(pkt.Headers) == 2 {
		ip, ok1  := pkt.Headers[0].(*pcap.Iphdr)
		tcp, ok2 := pkt.Headers[1].(*pcap.Tcphdr)
		if ok1 && ok2 && tcp.DestPort == 80 {
			pktString := StringFromPacket(pkt)
			for _, method := range httpMethods {
				if strings.Contains(pktString, method) {
					req, err := ParsePayload(pktString, ip, tcp, method)
					if err == nil {
						logToFile(req)
						go InitGUI()
					}
				}
			}
		}
	}
}
