package main

import (
	"fmt"
	"log"
	"net"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func main() {
	q := dnsmessage.Message{
		Header: dnsmessage.Header{ID: 1234, RecursionDesired: true},
		Questions: []dnsmessage.Question{{
			Name:  dnsmessage.MustNewName("lg.com."),
			Type:  dnsmessage.TypeA,
			Class: dnsmessage.ClassINET,
		}},
	}
	packed, err := q.Pack()
	if err != nil {
		log.Fatal(err)
	}
	resp, err := exchange("127.0.0.1:5300", packed)
	if err != nil {
		log.Fatal(err)
	}
	ips, err := parseIPs(resp)
	if err != nil {
		log.Fatal(err)
	}
	for _, ip := range ips {
		fmt.Println(ip)
	}
}

func exchange(addr string, msg []byte) ([]byte, error) {
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

func parseIPs(resp []byte) ([]net.IP, error) {
	var msg dnsmessage.Message

	err := msg.Unpack(resp)
	if err != nil {
		return nil, fmt.Errorf("error unpacking response: %w", err)
	}
	var ips []net.IP
	for _, ans := range msg.Answers {
		switch body := ans.Body.(type) {
		case *dnsmessage.AResource:
			ips = append(ips, net.IP(body.A[:]))
		}
	}
	return ips, nil
}
