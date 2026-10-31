package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

func main() {
	addr := &net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 5300}
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("listening on", conn.LocalAddr())

	buf := make([]byte, 4096)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			log.Println("read error:", err)
			continue
		}
		fmt.Printf("%d bytes from %s\n", n, from)

		resp, err := exchange("1.1.1.1:53", buf[:n])
		if err != nil {
			log.Println("exchange error:", err)
			continue
		}

		_, err = conn.WriteToUDP(resp, from)
		if err != nil {
			log.Println("write error:", err)
			continue
		}
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
