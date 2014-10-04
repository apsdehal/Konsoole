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
	"errors"
	"github.com/nsf/termbox-go"
)

var outWriter *bufio.Writer 
var errWriter *bufio.Writer

var gui *gocui.Gui
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

func InitGUI() {
	termbox.Flush()
	gui := gocui.NewGui()
	if err := gui.Init(); err != nil {
		panic(err)
	}
	defer gui.Close()
	gui.SetLayout(GUILayout)
	if err := KeyBindingsForGUI(gui); err != nil {
		panic(err)
	}
	gui.SelBgColor = gocui.ColorGreen
	gui.SelFgColor = gocui.ColorBlack
	gui.ShowCursor = true

	err := gui.MainLoop()
	if err != nil && err != gocui.ErrorQuit {
		panic(err)
	}
}

func GUILayout(g *gocui.Gui) error {
	var requestCount = map[string]int {
		"OPTIONS" : 0, "GET" : 0, "HEAD" : 0, "POST" : 0, "PUT" : 0, "DELETE" : 0, "TRACE" : 0, "CONNECT" : 0,
	}
	requests := []Request{}

	content, err := ioutil.ReadFile("./log.txt")
	if err != nil {
		panic(err)
	}
	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		if len(line) == 0 {
			break;
		}
		phrase := strings.SplitN(line, " ", 7)
		requestCount[phrase[2]]++
		// fmt.Println(phrase)
		requests = append(requests, Request{ phrase[0], phrase[1], phrase[2], phrase[3], int(phrase[4][0]), phrase[6], phrase[5] })
	}

	maxX, maxY :=  g.Size()
	
	if v, err := g.SetView("main-view", 15, -1, maxX, maxY); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		for _, request := range requests {
			timeWithZone := strings.Split(request.Time, "T")
			dateWithYear := timeWithZone[0]
			date := strings.SplitN(dateWithYear, "-", 2)[1]
			time := strings.Split(timeWithZone[1], "+")[0]
			fmt.Fprintf(v, "%s %s ▶ %s : %s\n", date, time, request.Method, request.Host)
		}
		v.Highlight = true
		if err := g.SetCurrentView("main-view"); err != nil {
			return err
		}
	}
	
	if v, err := g.SetView("side-view", -1, -1, 15, maxY); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		for key, value := range requestCount {
			if value != 0 {
				fmt.Fprintf(v, "%-8s %d\n", key, value)
			}
		}
	}
	return nil
}

func GUIUpdateLayout(g *gocui.Gui) error {
	if err := g.DeleteView("main-view"); err != nil {
		return err
	}

	if err := g.DeleteView("side-view"); err != nil {
		return err
	}
	GUILayout(g)
	return nil
}
func KeyBindingsForGUI(g *gocui.Gui) error {
	if err := g.SetKeybinding("side-view", gocui.KeyCtrlSpace, 0, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("main-view", gocui.KeyCtrlSpace, 0, nextView); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowDown, 0, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, 0, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowLeft, 0, cursorLeft); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowRight, 0, cursorRight); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, 0, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side-view", gocui.KeyEnter, 0, getLine); err != nil {
		return err
	}
	if err := g.SetKeybinding("msg", gocui.KeyEnter, 0, delMsg); err != nil {
		return err
	}

	return nil
}

func nextView(g *gocui.Gui, v *gocui.View) error {
	currentView := g.CurrentView()
	if currentView == nil || currentView.Name() == "side-view" {
		return g.SetCurrentView("main-view")
	}
	return g.SetCurrentView("side-view")
}

func cursorDown(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy+1); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox, oy+1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorUp(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx, cy-1); err != nil && oy > 0 {
			if err := v.SetOrigin(ox, oy-1); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorLeft(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		ox, oy := v.Origin()
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx-1, cy); err != nil && ox > 0 {
			if err := v.SetOrigin(ox-1, oy); err != nil {
				return err
			}
		}
	}
	return nil
}

func cursorRight(g *gocui.Gui, v *gocui.View) error {
	if v != nil {
		cx, cy := v.Cursor()
		if err := v.SetCursor(cx+1, cy); err != nil {
			ox, oy := v.Origin()
			if err := v.SetOrigin(ox+1, oy); err != nil {
				return err
			}
		}
	}
	return nil
}

func getLine(g *gocui.Gui, v *gocui.View) error {
	var l string
	var err error

	_, cy := v.Cursor()
	if l, err = v.Line(cy); err != nil {
		l = ""
	}

	maxX, maxY := g.Size()
	if v, err := g.SetView("msg", maxX/2-30, maxY/2, maxX/2+30, maxY/2+2); err != nil {
		if err != gocui.ErrorUnkView {
			return err
		}
		fmt.Fprintln(v, l)
		if err := g.SetCurrentView("msg"); err != nil {
			return err
		}
	}
	return nil
}

func delMsg(g *gocui.Gui, v *gocui.View) error {
	if err := g.DeleteView("msg"); err != nil {
		return err
	}
	if err := g.SetCurrentView("side-view"); err != nil {
		return err
	}
	return nil
}

func quit(g *gocui.Gui, v *gocui.View) error {
	return gocui.ErrorQuit
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

func LogToFile(r Request) {
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
						LogToFile(req)
						go InitGUI()
					}
				}
			}
		}
	}
}

func main () {
	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)
	handleToDevice := Init()
	go InitGUI()
	// InitGUI()
	for {
		pkt := handleToDevice.Next()
		if pkt != nil {
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}