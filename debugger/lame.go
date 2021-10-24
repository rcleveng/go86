package go86

import (
	"net"
	"strconv"
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
		cmds := strings.Fields(data)
		if len(cmds) == 0 {
			log.Warningln("Got empty commands from lame debugger input.")
			continue
		}
		switch unicode.ToUpper(rune(cmds[0][0])) {
		case 'C':
			request <- DebuggerRequest{Cmd: CONTINUE, Data: data}
		case 'D':
			request <- DebuggerRequest{Cmd: DETACH, Data: data}
		case 'H':
			request <- DebuggerRequest{Cmd: HEARTBEAT, Data: data}
		case 'I':
			request <- DebuggerRequest{Cmd: INFO, Data: data}
		case 'M':
			if len(cmds) < 4 {
				log.Warningln("Invalid Memory command. Expected Mem Seg Off Len.")
				continue
			}
			seg, err := strconv.ParseInt(cmds[1], 16, 16)
			if err != nil {
				log.Warningf("Invalid Memory command [seg]. Expected Mem Seg Off Len: %v ", err)
				continue
			}
			off, err := strconv.ParseInt(cmds[2], 16, 16)
			if err != nil {
				log.Warningf("Invalid Memory command [off]. Expected Mem Seg Off Len: %v ", err)
				continue
			}
			length, err := strconv.ParseInt(cmds[3], 16, 16)
			if err != nil {
				log.Warningf("Invalid Memory command [len]. Expected Mem Seg Off Len: %v ", err)
				continue
			}
			m := DebuggerMemoryRequest{Seg: int(seg), Off: int(off), Length: int(length)}
			request <- DebuggerRequest{Cmd: MEMORY, Data: data, Mem: m}
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
		log.Infof("Lame: Got Debugger Response: '%v'", r)
		s.conn.Write([]byte(r.Text))
	}

}
