package go86

import (
	"fmt"
	"net"
	"strconv"
	"strings"
	"unicode"

	log "github.com/golang/glog"
)

type GdbSession struct {
	conn net.Conn
}

// Gets the packet data or returns an error if the bytes of the packet
// are not valid.
func (S GdbSession) GetPacketData(raw []byte, n int) (string, error) {
	data := strings.TrimSpace(string(raw[:n]))
	idx := strings.IndexByte(data, '$')
	if idx < 0 {
		return "", fmt.Errorf("invalid packet, missing '$', line: '%v'", data)
	}
	data = data[idx+1:]
	idx = strings.IndexByte(data, '#')
	if idx < 0 {
		return "", fmt.Errorf("invalid packet, missing checksum: '%v', idx: '%v', split: '%v'", data, idx, data[idx+1:])
	}
	checksum, err := strconv.ParseInt(data[idx+1:], 16, 16)
	if err != nil {
		return "", fmt.Errorf("invalid packet, missing checksum: '%v', idx: '%v', split: '%v'", data, idx, data[idx+1:])
	}
	data = data[:idx]
	if len(data) == 0 {
		return "", fmt.Errorf("got empty packet")
	}

	if expChksum := CreateChecksum(data); expChksum != int(checksum) {
		return "", fmt.Errorf("checksum did not match. got: '%v', expected: '%v'", checksum, expChksum)
	}
	return data, nil
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
		data, err := s.GetPacketData(buf, n)
		if err != nil {
			log.Error(err.Error())
			continue
		}
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
		log.Infof("GDB: Got Debugger Response: '%v'", r)
		s.conn.Write([]byte("+\r\n"))
	}

}

func CreateChecksum(cmd string) (chksum int) {
	for _, c := range cmd {
		chksum += int(c)
	}
	chksum &= 255
	return
}

func CreatePacket(cmd string) string {
	chksum := 0
	for _, c := range cmd {
		chksum += int(c)
	}
	chksum &= 255
	return "$" + cmd + "#" + strconv.Itoa(chksum)
}
