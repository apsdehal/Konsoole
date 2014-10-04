package main 

import (
	"github.com/akrennmair/gopcap"
	"github.com/jroimartin/gocui"
	"fmt"
	"strings"
	"bytes"
	"os"
	"bufio"
	"regexp"
	"time"
	"io/ioutil"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

type Request struct {
	DestIp    string
	Time      time.Time
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

func InitGUI() {
	g := gocui.NewGui()
	if err := g.Init(); err := nil {
		panic(err)
	}
	defer g.Close()
	g.SetLayout(GUILayout)
	if err := keybindings(g); err != nil {
		panic(err)
	}
	g.SetBgColor = gocui.ColorGreen
	g.SetFgColor = gocui.ColorBlack
	g.ShowCursor = true

	err := g.MainLoop()
	if err != nil && err != gocui.ErrorQuit {
		panic(err)
	}
}

func GUILayout(g *gocui.Gui) error {
	var requestCount := map[string]int {
		"OPTIONS" : 0, "GET" : 0, "HEAD" : 0, "POST" : 0, "PUT" : 0, "DELETE" : 0, "TRACE" : 0, "CONNECT" : 0,
	}
	requests := []Request{}
	content, err := ioutil.ReadFile("log.txt")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), '\n')
	var int k = 0
	for _, line := range lines {
		phrase := strings.Spilt(line, ' ')
		requestCount[phrase[2]]++
		requests = append(requests, Request{ phrase[0], phrase[1], phrase[2], phrase[3], phrase[4], phrase[5], phrase[6] })
	}

	maxX, maxY :=  g.Size()
	
	if v, err := g.SetView("main-view", 30, -1, maxX, maxY); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		for _, request := range requests {
			fmt.Fprintln(v, "%s %s", request.Method, request.Host)
		}
		v.Highlight = true
		if err := g.SetCurrentView("main-view"); err != nil {
			return err
		}
	}
	
	if v, err := g.SetView("side-view", -1, -1, 30, maxY); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		for _, count := range requestCount {
			fmt.Fprintln(v, requestCount[count])
		}
	}
	return nil
}

func StringFromPacket(pkt *pcap.Packet) string {
	buf := bytes.NewBufferString("")
	for i := uint32(0); int(i) < len(pkt.Data); i++ {
		fmt.Fprintf(buf, "%c", pkt.Data[i])
	}
	return string(buf.Bytes())
}

func ParsePayload(pktString string, ip *pcap.Iphdr, tcp *pcap.Tcphdr, method string) Request {
	SrcIp    := ip.SrcAddr()
	DestIp   := ip.DestAddr()
	SrcPort  := tcp.SrcPort
	DestPort := tcp.DestPort

	reqRegex, _		  := regexp.Compile("/(.+)\\s+HTTP/1.([01])\\s+")
	hostRegex, _      := regexp.Compile("Host: (.+)\\s+")
	useragentRegex, _ := regexp.Compile("User-Agent:(.+)")

	host 	  := hostRegex.FindStringSubmatch(pktString)[1] 
	useragent := useragentRegex.FindStringSubmatch(pktString)[1] 
	req  	  := reqRegex.FindStringSubmatch(pktString) 
	path 	  := req[1]
	httpType  := req[2]

	rp := Request{ 
					DestIp, 
					time.Now(), 
					method,
					host, 
					int(httpType[0]), 
					useragent, 
					path,
				}
	return rp
}

func LogToFile(r Request) {
	f, err := os.OpenFile("log.txt", os.O_APPEND|os.O_WRONLY, 0600)
	if err != nil {
	    panic(err)
	}

	defer f.Close()
	format := "%s %s %s %s %d %s %s\n"
	msg := fmt.Sprintf(format, r.DestIp, r.Time, r.Method, r.Host, r.HTTPType, r.UserAgent, r.Path )
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
					req := ParsePayload(pktString, ip, tcp, method)
					LogToFile(req)
				}
			}
		}
	}
}

func main () {
	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)

	handleToDevice := Init()
	for {
		pkt := handleToDevice.Next()
		if pkt != nil {
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}