package main

import (
	lte "github.com/farisey-ru/gotest/lte_listener"
	"log"
	"os"
)

func main() {
	patterns := []string{
		"^/devices/platform/1e1e0000\\.xchi/usb1/1-2/1-2\\.[0-9]$",
		"^/devices/pci0000:00/0000:00:13\\.2/usb4/4-5/4-5:[0-9]\\.[0-9]+$",
		"^/devices/pci0000:00/0000:00:02\\.4/0000:01:00\\.0/usb3/3-5/3-5:1\\.[0-9]+$",
		"^/devices/pci0000:00/0000:00:02\\.4/0000:01:00\\.0/usb3/3-4/3-4:1\\.[0-9]+$",
		"^/devices/pci0000:00/0000:00:12\\.2/usb3/3-1/3-1:1\\.[0-9]+",
	}

	drivers := []string{
		"^ftdi_sio$",
		"^usb-storage$",
	}

	lte, err := lte.Subscribe(os.Getpagesize(), patterns, drivers)
	if err != nil {
		panic(err)
	}

	defer lte.Close()

	in := lte.Listen()
	//log.Printf("in: %T: %v\n", in, in)

	finished := make(chan bool) // avoid wait group
	go func() {
		// while 'in' is not closed
		for msg := range in {
			log.Printf("msg: %+v\n", msg)
		}
		finished <- true
	}()

	<-finished
}
