package go86

import (
	"net"
	"strconv"

	"github.com/golang/glog"
	cpu "go86.org/go86/cpu"
)

type GdbSession struct {
}

func (s GdbSession) Run(conn net.Conn,
	request chan cpu.DebuggerRequest,
	response chan cpu.DebuggerResponse) {

}

func CreatePacket(cmd string) string {
	chksum := 0
	for _, c := range cmd {
		chksum += int(c)
	}
	chksum &= 255
	return "$" + cmd + "#" + strconv.Itoa(chksum)
}

func Listen(port int, request chan cpu.DebuggerRequest, response chan cpu.DebuggerResponse) error {
	lsock, err := net.Listen("tcp", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer lsock.Close()
	glog.Infof("Listening on port %v", port)

	conn, err := lsock.Accept()
	if err != nil {
		return err
	}

	session := GdbSession{}
	go session.Run(conn, request, response)

	return nil
}
