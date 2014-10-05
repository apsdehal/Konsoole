// Main file for Konsoole initializes the software
package main 

// Import required packages
import (
	"os"
	"bufio"
	"fmt"
	"flag"
)

// Get a bufio output and error Writer
var outWriter *bufio.Writer 
var errWriter *bufio.Writer
// Global variable used for the log file
var fileToOpen string

// Main function that handles initialization for network interfaces and listens 
// forever for the network packets
func main() {
	// Check if custom file is required
	fileToSave := flag.String("f", "", "Custom log file")
	flag.Parse()

	// If not create a temp file
	if len(*fileToSave) == 0 {
		os.Chdir("/tmp")
		f, err := os.Create("log_konsoole.txt")
		if err != nil || f == nil {
			panic(err)
		}
		fileToOpen = "log_konsoole.txt"
		defer os.Remove("log_konsoole.txt")
	} else {
		// If yes, use it as log file
		fileToOpen = *fileToSave
	}

	fmt.Println("Konsoole: HTTP Monitor")

	outWriter = bufio.NewWriter(os.Stdout)
	errWriter = bufio.NewWriter(os.Stderr)
	// Initialize the network intefaces and start listening
	handleToDevice := Init()
	// Use a go routine to initialize the GUI
	go InitGUI()
	// Listen forever for packets
	for {
		pkt := handleToDevice.Next()
		if pkt != nil {
			pkt.Decode()
			DecodePacket(pkt)
		}
	}
}