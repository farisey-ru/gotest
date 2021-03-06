package lte_listener

import (
	nl "github.com/farisey-ru/gotest/nl_kobj"
	"github.com/farisey-ru/gotest/regexp_ext"
	"io/ioutil"
	"os"
	"path"
	"strconv"
)

type LteMsg struct {
	ev           nl.NlKobjEv
	numEndpoints uint   // .../bNumEndpoints
	sysfs_iface  string // .../interface
	device       string // "ttyUSB3"
}

// do not forget the final slash
const DevMountPoint string = "/dev/"

type devmap map[string]string

type Lte struct {
	sk         nl.NlKobjSock
	re_drivers regexp_ext.RegexpArray
	devs       devmap
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

// -> "/dev/ttyUSB3"
func (msg *LteMsg) Device() string {
	return DevMountPoint + msg.device
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
		devs:       make(devmap),
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

	/* Q: How could we stop this?
	 * A: just below, near 'return'
	 */
	go func() {
		defer func() {
			close(out)
		}()

		for {
			ev, err := lte.sk.Receive()
			if err != nil {
				/* exit point. See the detailed
				 * comment in NlKobjSock.Receive()
				 */
				return
			}

			msg := LteMsg{
				ev: *ev,
			}

			switch ev.Event() {
			case nl.NLKEV_UNBIND:
				if len(lte.devs[ev.Path()]) == 0 {
					// ignore the detaching of unknown
					continue
				}
				msg.device = lte.devs[ev.Path()]
				//add nothing else to the msg

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

				_, msg.device = path.Split(ev.Path())
				lte.devs[ev.Path()] = msg.device
				if _, err := os.Stat(msg.Device()); err != nil {
					// node not exists
					continue
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
