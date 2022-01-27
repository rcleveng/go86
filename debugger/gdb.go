package go86

import (
	"fmt"
	"net"
	"strconv"
	"strings"

	log "github.com/golang/glog"
)

type GdbSession struct {
	conn     net.Conn
	threadId int64
	needAck  bool
}

// Gets the packet data or returns an error if the bytes of the packet
// are not valid.
func (S GdbSession) GetPacketData(raw []byte, n int) (string, error) {
	log.Infof("GetPacketData: size: %d; raw: '%v'", n, raw[:n])
	data := strings.TrimSpace(string(raw[:n]))
	idx := strings.IndexByte(data, '$')
	if idx < 0 {
		return "", fmt.Errorf("invalid packet, missing '$', line: '%v'", data)
	}
	data = data[idx+1:]
	idx = strings.IndexByte(data, '#')
	if idx < 0 {
		return "", fmt.Errorf("invalid packet, missing checksum, no '#': '%v', idx: '%v', split: '%v'", data, idx, data[idx+1:])
	}
	cs := data[idx+1:]
	// Remove any overlapping '+' acknowledgement
	cs = strings.TrimRight(cs, "+")
	checksum, err := strconv.ParseInt(cs, 16, 16)
	if err != nil {
		return "", fmt.Errorf("invalid packet, missing checksum, parseint failed: '%v', idx: '%v', split: '%v'", data, idx, data[idx+1:])
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

func (s GdbSession) HandleQueryRequest(request chan DebuggerRequest, data string) {
	req := DebuggerRequest{
		Cmd:        INFO,
		RawCmdText: data,
	}
	if strings.HasPrefix(data, "qSupported:") {
		sup := strings.Split(data[11:], ";")
		log.Infof("Supported: '%V'", sup)
		request <- req
		return
	}
	log.Info("Query Request: ", data)
	switch data {
	case "qTStatus":
		s.WriteAck()
		s.WriteResponse("") // tnotrun:0?? didn't work
	default:
		log.Info("Unknown Query Request: ", data)
		s.WriteAck()
		s.WriteResponse("")
	}
}

func (s *GdbSession) WriteResponse(data string) {
	packet := CreatePacket(data)
	if s.needAck {
		// Prepend an ACK
		packet = "+" + packet
		s.needAck = false
	}
	log.Infof("GdbSession::WriteResponse: '%s', ack: %v", packet, s.needAck)
	s.conn.Write([]byte(packet))
}

func (s *GdbSession) WriteAck() {
	log.Infoln("GdbSession::WriteAck.")
	s.needAck = true
}

func (s GdbSession) HandleVRequest(request chan DebuggerRequest, data string) {
	s.WriteAck()
	switch data {
	case "vMustReplyEmpty":
		// The ‘vMustReplyEmpty’ is used as a feature test to check how
		// gdbserver handles unknown packets, it is important that this packet
		// be handled in the same way as other unknown ‘v’ packets.
	}
	s.WriteResponse("")
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
		if n == 1 && buf[0] == '+' {
			// skip over + responses, we don't need to worry about them.
			continue
		}
		data, err := s.GetPacketData(buf, n)
		if err != nil {
			log.Errorf("Error '%v' in packet '%s'", err, buf)
			continue
		}
		log.Infof("Got: '%s'", data)
		switch data[0] {
		case '?':
			// TODO: don't be lame, return a real value
			request <- DebuggerRequest{Cmd: STOP_REASON, RawCmdText: "?", Data: data}
		case 'q':
			s.HandleQueryRequest(request, data)
		case 'v':
			s.HandleVRequest(request, data)
		case 'H':
			s.WriteAck()
			if len(data) < 3 {
				log.Warningf("Malformed Hg requst: '%v'", data)
				s.WriteResponse("")
				continue
			}
			threadId, err := strconv.ParseInt(data[2:], 16, 16)
			if err != nil {
				log.Warningf("Malformed Hg requst, bad thread ID: '%v'", data)
				s.WriteResponse("")
				continue
			}
			s.threadId = threadId
			s.WriteResponse("OK")
		default:
			request <- DebuggerRequest{Cmd: UNKNOWN_COMMAND, RawCmdText: data, Data: data}
		}
	}
}

func (s GdbSession) HandleResponses(response chan DebuggerResponse) {
	log.Info("GdbSession::HandleResponses")
	for r := range response {
		log.Infof("GDB: Got Debugger Response: '%v': cmd:'%s'; all '%v'", r.Type, r.RawCmdText, r)
		s.WriteAck()
		switch r.Cmd {
		case STEPPED, INFO, HEARTBEAT:
			//
		case CONTINUE, HALT, STOP_REASON:
			s.WriteResponse(fmt.Sprintf("S%02d", r.Signal))
		default:
			log.Warningln("Unhandled response type: ", r.Type)
			s.WriteResponse("")
		}

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
	return fmt.Sprintf("$%s#%02x", cmd, chksum)
}
