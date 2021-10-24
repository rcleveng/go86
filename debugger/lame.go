package go86

import (
	"net"
	"strings"
	"unicode"

	log "github.com/golang/glog"
)

type LameSession struct {
	conn net.Conn
}

func (s LameSession) HandleRequests(request chan DebuggerRequest) {

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

func (s LameSession) HandleResponses(response chan DebuggerResponse) {

	for r := range response {
		log.Infof("Got Debugger Response: '%s'", r)
		s.conn.Write([]byte(r.Text))
	}

}
