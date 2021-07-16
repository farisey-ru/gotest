package nl_kobj

import (
	"github.com/farisey-ru/gotest/regexp_ext"
	"golang.org/x/sys/unix"
	"strconv"
	"strings"
)

//import "os"
//import "log"

type NlKobjSock struct {
	fd  int
	buf []byte
	re  regexp_ext.RegexpArray
}

type NlKobjEventer interface {
	Event() uint
	Path() string
}

const (
	NLKEV_UNKNOWN = iota
	NLKEV_ADD
	NLKEV_REMOVE
	NLKEV_BIND
	NLKEV_UNBIND
	NLKEV_CHANGE
)

var ev_map = map[string]uint{
	"add":    NLKEV_ADD,
	"remove": NLKEV_REMOVE,
	"bind":   NLKEV_BIND,
	"unbind": NLKEV_UNBIND,
	"change": NLKEV_CHANGE,
}

type NlKobjEv struct {
	событие      uint
	путь_в_sysfs string
	Subsys       string
	Devtype      string
	Driver       string
	Product      [3]uint
	Type         [3]uint
	Interface    [3]uint
}

func (ev *NlKobjEv) Event() uint {
	return ev.событие
}

func (ev *NlKobjEv) Path() string {
	return ev.путь_в_sysfs
}

func Subscribe(rcvbuf int, expr []string) (*NlKobjSock, error) {
	re, err := regexp_ext.CompileExpr(expr)
	if err != nil {
		return nil, err
	}

	fd, err := unix.Socket(unix.AF_NETLINK,
		unix.SOCK_RAW,
		unix.NETLINK_KOBJECT_UEVENT)
	if err != nil {
		return nil, err
	}

	unix.SetsockoptInt(fd, unix.SOL_SOCKET, unix.SO_RCVBUF, rcvbuf)

	sk := &NlKobjSock{
		fd:  fd,
		buf: make([]byte, rcvbuf),
		re:  *re,
	}

	sanl := &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: 1,
		Pid:    uint32(unix.Getpid()),
	}

	if err = unix.Bind(sk.fd, sanl); err != nil {
		unix.Close(sk.fd)
		return nil, err
	}

	return sk, nil
}

func (sk *NlKobjSock) Close() error {
	err := unix.Close(sk.fd)
	return err
}

func (sk *NlKobjSock) MatchPath(s string) bool {
	return sk.re.MatchString(s)
}

func three(arr *[3]uint, str string, base int) {
	parts := strings.SplitN(str, "/", 3)
	for i, x := range parts {
		u, err := strconv.ParseUint(x, base, 32)
		if err != nil {
			panic(err)
		}
		arr[i] = uint(u)
	}
}

func (sk *NlKobjSock) Receive() (*NlKobjEv, error) {
	for {
		n, _, err := unix.Recvfrom(sk.fd, sk.buf, 0)
		if err != nil {
			return nil, err
		}

		// "add@devices/path/in/sysfs\0KEY=VALUE\0KEY=VALUE\0..."
		all := strings.Split(string(sk.buf[0:n]), "\000")
		head := all[0]
		parts := strings.Split(head, "@")
		if !sk.MatchPath(parts[1]) {
			continue
		}

		ev := &NlKobjEv{
			событие:      ev_map[parts[0]],
			путь_в_sysfs: parts[1],
		}

		for _, token := range all[1:] {
			tk := strings.SplitN(token, "=", 2)
			switch tk[0] {
			case "SUBSYSTEM":
				ev.Subsys = tk[1]
			case "DEVTYPE":
				ev.Devtype = tk[1]
			case "DRIVER":
				ev.Driver = tk[1]
			case "PRODUCT":
				three(&ev.Product, tk[1], 16)
			case "TYPE":
				three(&ev.Type, tk[1], 10)
			case "INTERFACE":
				three(&ev.Interface, tk[1], 10)
			}
		}

		return ev, nil
	}
}
