package main

import (
	lte "github.com/farisey-ru/gotest/lte_listener"
	nl "github.com/farisey-ru/gotest/nl_kobj"
	"log"
	"os"
	"os/exec"
)

func main() {
	patterns := []string{
		"^/devices/platform/1e1e0000\\.xchi/usb1/1-2/1-2\\.[0-9a-fA-F]+/",
		// stend pathes are below, remove us
		"^/devices/pci0000:00/0000:00:13\\.2/usb4/4-5/4-5:[0-9]\\.[0-9a-fA-F]+/",
		"^/devices/pci0000:00/0000:00:02\\.4/0000:01:00\\.0/usb3/3-5/3-5:1\\.[0-9a-fA-F]+/",
		"^/devices/pci0000:00/0000:00:02\\.4/0000:01:00\\.0/usb3/3-4/3-4:1\\.[0-9a-fA-F]+/",
		"^/devices/pci0000:00/0000:00:12\\.2/usb3/3-1/3-1:1\\.[0-9a-fA-F]+/",
	}

	drivers := []string{
		// set your drivers
		"^option",
		//"^ftdi_sio$",
		//"^usb-storage$",
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
			switch msg.Event() {
			case nl.NLKEV_UNBIND:
				log.Println("What should I do on Unbind of", msg.Device(), "?")
			case nl.NLKEV_BIND:
				/* Not debugged since I do not have
				 * either an OpenWrt-based device or
				 * free time to prepare that.
				 */
				dev := msg.Device()
				out, err := exec.Command("uqmi", "-d", dev,
					"--get-signal-info").Output()
				if err != nil {
					log.Printf("%s get-signal error: %v\n", dev, err)
					continue
				}
				log.Println("signal:", out)

				cmd := exec.Command("uqmi", "-d", dev,
					"--start-network", "internet",
					"--autoconnect")
				err = cmd.Run()
				if err != nil {
					log.Printf("Starting %s failed: %v\n", dev, err)
				}
			default:
				panic("Unknown msg type")
			}
		}
		finished <- true
	}()

	<-finished
}
