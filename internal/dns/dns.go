package dns

import (
	"errors"
	"fmt"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func Exchange(addr string, msg []byte) ([]byte, error) {
	conn, err := net.Dial("udp", addr)
	if err != nil {
		return nil, fmt.Errorf("error connecting to server: %w", err)
	}
	defer conn.Close()
	conn.SetReadDeadline(time.Now().Add(2 * time.Second))

	_, err = conn.Write(msg)
	if err != nil {
		return nil, fmt.Errorf("error writing to server: %w", err)
	}

	resp := make([]byte, 4096)
	n, err := conn.Read(resp)
	if err != nil {
		return nil, fmt.Errorf("error reading from server: %w", err)
	}
	return resp[:n], nil
}

func ParseQuestion(b []byte) (dnsmessage.Question, error) {
	var msg dnsmessage.Message
	err := msg.Unpack(b)
	if err != nil {
		return dnsmessage.Question{}, fmt.Errorf("error unpacking query: %w", err)
	}

	if len(msg.Questions) == 0 {
		return dnsmessage.Question{}, errors.New("no questions in packet")
	}
	return msg.Questions[0], nil
}
