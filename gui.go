package main

import (
	"github.com/jroimartin/gocui"
	"fmt"
	"strings"
	"io/ioutil"
	"os/exec"
	"os"
)

var gui *gocui.Gui

func InitGUI() {
	clearScreen()
	gui := gocui.NewGui()
	if err := gui.Init(); err != nil {
		panic(err)
	}
	gui.Flush()
	defer gui.Close()
	gui.SetLayout(GUILayout)
	if err := KeyBindingsForGUI(gui); err != nil {
		panic(err)
	}
	gui.SelBgColor = gocui.ColorGreen
	gui.SelFgColor = gocui.ColorBlack
	// gui.Show	Cursor = true

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

	content, err := ioutil.ReadFile(fileToOpen)
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
		v.Clear()
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
		v.Clear()
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

func KeyBindingsForGUI(g *gocui.Gui) error {
	if err := g.SetKeybinding("", gocui.KeyArrowDown, 0, cursorDown); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyArrowUp, 0, cursorUp); err != nil {
		return err
	}
	if err := g.SetKeybinding("", gocui.KeyCtrlC, 0, quit); err != nil {
		return err
	}
	if err := g.SetKeybinding("side-view", gocui.KeyEnter, 0, getLine); err != nil {
		return err
	}
	return nil
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

func clearScreen() {
    cmd := exec.Command("clear")
    cmd.Stdout = os.Stdout
    cmd.Run()
}