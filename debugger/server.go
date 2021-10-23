package go86

import (
	"net"
	"strconv"
	"strings"
	"unicode"

	log "github.com/golang/glog"
)

type GdbSession struct {
	conn net.Conn
}

func (s GdbSession) HandleRequests(request chan DebuggerRequest) {

	defer s.conn.Close()
	buf := make([]byte, 1024)

	for {
		n, err := s.conn.Read(buf)
		if err != nil || n == 0 {
			log.Infoln("Got error or zero read; exiting GDB Socket Reader")
			break
		}
		data := strings.TrimSpace(string(buf[:n]))
		log.Infof("Got: '%s'", data)
		switch unicode.ToUpper(rune(data[0])) {
		case 'C':
			request <- DebuggerRequest{Cmd: CONTINUE, Data: data}
		case 'D':
			request <- DebuggerRequest{Cmd: DETACH, Data: data}
		case 'H':
			request <- DebuggerRequest{Cmd: HEARTBEAT, Data: data}
		case 'I':
			request <- DebuggerRequest{Cmd: INFO, Data: data}
		case 'Q':
			request <- DebuggerRequest{Cmd: HALT, Data: data}
		case 'S':
			request <- DebuggerRequest{Cmd: STEP, Data: data}
		default:
			request <- DebuggerRequest{Cmd: UNKNOWN_COMMAND, Data: data}
		}
	}
}

func (s GdbSession) HandleResponses(response chan DebuggerResponse) {

	for r := range response {
		log.Infof("Got Debugger Response: '%s'", r)
		s.conn.Write([]byte("+\r\n"))
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

func Listen(port int, dbgtype string, request chan DebuggerRequest, response chan DebuggerResponse) error {
	lsock, err := net.Listen("tcp4", ":"+strconv.Itoa(port))
	if err != nil {
		return err
	}
	defer lsock.Close()
	log.Infof("Listening on port %v", port)

	for {
		conn, err := lsock.Accept()
		if err != nil {
			log.Infof("Got error from Accept: %v", err)
			return err
		}

		if dbgtype == "gdb" {
			session := GdbSession{conn: conn}
			go session.HandleRequests(request)
			go session.HandleResponses(response)
		} else if dbgtype == "lame" {

		} else {
			panic("Unknown dbgtype: " + dbgtype)
		}
	}
}
