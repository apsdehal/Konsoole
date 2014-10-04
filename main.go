package main 

import (
	"github.com/akrennmair/gopcap"
	"fmt"
	"strings"
	"bytes"
	"os"
	"bufio"
	"regexp"
	"time"
	"errors"
	"os/exec"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

type Request struct {
	DestIp    string
	Time      string
	Method 	  string
	Host 	  string
	HTTPType  int
	UserAgent string
	Path 	  string
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
	for i := uint32(0); int(i) < len(pkt.Data); i++ {
		fmt.Fprintf(buf, "%c", pkt.Data[i])
	}
	return string(buf.Bytes())
}

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

func logToFile(r Request) {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0777)
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

func DecodePacket(pkt *pcap.Packet ) {
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
						InitGUI()
					}
				}
			}
		}
	}
}
func clearScreen() {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}

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