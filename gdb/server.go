package go86

import (
	"net"
	"strconv"
	"strings"
	"unicode"

	log "github.com/golang/glog"
	cpu "go86.org/go86/cpu"
)

type GdbSession struct {
}

func (s GdbSession) HandleRequests(conn net.Conn,
	request chan cpu.DebuggerRequest) {

	defer conn.Close()
	buf := make([]byte, 1024)

	for {
		n, err := conn.Read(buf)
		if err != nil || n == 0 {
			continue
		}
		data := strings.TrimSpace(string(buf[:n]))
		log.Infof("Got: '%s'", data)
		switch unicode.ToUpper(rune(data[0])) {
		case 'C':
			request <- cpu.DebuggerRequest{Cmd: cpu.CONTINUE, Data: data}
		case 'D':
			request <- cpu.DebuggerRequest{Cmd: cpu.DETACH, Data: data}
		case 'H':
			request <- cpu.DebuggerRequest{Cmd: cpu.HEARTBEAT, Data: data}
		case 'I':
			request <- cpu.DebuggerRequest{Cmd: cpu.INFO, Data: data}
		case 'Q':
			request <- cpu.DebuggerRequest{Cmd: cpu.HALT, Data: data}
		case 'S':
			request <- cpu.DebuggerRequest{Cmd: cpu.STEP, Data: data}
		default:
			request <- cpu.DebuggerRequest{Cmd: cpu.UNKNOWN_COMMAND, Data: data}
		}
	}
}

func (s GdbSession) HandleResponses(conn net.Conn,
	response chan cpu.DebuggerResponse) {

	for r := range response {
		log.Infof("Got Debugger Response: '%s'", r)
		conn.Write([]byte("+\r\n"))
	}

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
	lsock, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer lsock.Close()
	log.Infof("Listening on port %v", port)

	for {
		conn, err := lsock.Accept()
		if err != nil {
			return err
		}

		session := GdbSession{}
		go session.HandleRequests(conn, request)
		go session.HandleResponses(conn, response)
	}
}
