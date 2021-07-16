package lte_listener

import (
	nl "github.com/farisey-ru/gotest/nl_kobj"
	"github.com/farisey-ru/gotest/regexp_ext"
	"io/ioutil"
	"os"
	"strconv"
)

type LteMsg struct {
	ev           nl.NlKobjEv
	numEndpoints uint   // .../bNumEndpoints
	sysfs_iface  string // .../interface
}

type Lte struct {
	sk         nl.NlKobjSock
	re_drivers regexp_ext.RegexpArray
}

func (msg *LteMsg) Event() uint {
	return msg.ev.Event()
}

func (msg *LteMsg) Path() string {
	return msg.ev.Path()
}

func (msg *LteMsg) NumEndpoints() uint {
	return msg.numEndpoints
}

func (msg *LteMsg) Interface() string {
	return msg.sysfs_iface
}

func Subscribe(rcvbuf int, expr_path []string, expr_driver []string) (*Lte, error) {
	re, err := regexp_ext.CompileExpr(expr_driver)
	if err != nil {
		return nil, err
	}

	sk, err := nl.Subscribe(rcvbuf, expr_path)
	if err != nil {
		return nil, err
	}

	ret := &Lte{
		sk:         *sk,
		re_drivers: *re,
	}
	return ret, nil
}

func (lte *Lte) Close() error {
	err := lte.sk.Close()
	return err
}

func (lte *Lte) MatchPath(s string) bool {
	return lte.sk.MatchPath(s)
}

func (lte *Lte) MatchDriver(s string) bool {
	return lte.re_drivers.MatchString(s)
}

func (lte *Lte) Listen() <-chan LteMsg {
	out := make(chan LteMsg, 4)

	go func() {
		defer func() {
			close(out)
		}()

		for {
			ev, err := lte.sk.Receive()
			if err != nil {
				return
			}

			msg := LteMsg{
				ev: *ev,
			}

			switch ev.Event() {
			case nl.NLKEV_UNBIND:
				//add nothing to the msg

			case nl.NLKEV_BIND:
				if !lte.MatchDriver(ev.Driver) {
					continue
				}

				var fn string

				fn = "/sys" + ev.Path() + "/bNumEndpoints"
				//dat, err := ioutil.ReadFile(fn) // os. instead of ioutil. since v1.16 !
				//if (err == nil) && !(len(dat) > 0) {
				if _, err := os.Stat(fn); err == nil {
					// must be if watching a correct path
					dat, _ := ioutil.ReadFile(fn)
					x, _ := strconv.Atoi(string(dat[:len(dat)-1]))
					msg.numEndpoints = uint(x)
				}

				fn = "/sys" + ev.Path() + "/interface"
				if _, err := os.Stat(fn); err == nil {
					dat, _ := ioutil.ReadFile(fn)
					msg.sysfs_iface = string(dat)
				}

				// add any stuff you need here

			default:
				// ignore all other event types
				continue
			}

			// cut the crap
			out <- msg
		}
	}()
	return out
}
