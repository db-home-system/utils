package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/jacobsa/go-serial/serial"
)

func main() {

	// Set up options.
	options := serial.OpenOptions{
		PortName:        "/dev/ttyS2",
		BaudRate:        115200,
		DataBits:        8,
		StopBits:        1,
		MinimumReadSize: 1,
	}

	// Open the port.
	port, err := serial.Open(options)
	if err != nil {
		log.Fatalf("serial.Open: %v", err)
	}

	// Make sure to close it later.
	defer port.Close()

	for {
		buf := make([]byte, 100)
		n, err := port.Read(buf)
		if err != nil {
			log.Fatalf("port.Read: %v", err)
		}
		line := string(buf[:n])
		fmt.Print(line)
		if strings.HasPrefix(line, "$") {
			line = line[1:]
			fields := line.Split(line, ";")
			fmt.Printf(fields)
		}
	}
}
