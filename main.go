package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"golang.org/x/net/dns/dnsmessage"
)

func main() {
	listen := flag.String("listen", "127.0.0.1:5300", "address to listen on")
	upstream := flag.String("upstream", "1.1.1.1:53", "upstream DNS server")
	flag.Parse()

	addr, err := net.ResolveUDPAddr("udp", *listen)
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()
	fmt.Println("listening on", conn.LocalAddr())

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	go func() {
		<-ctx.Done()
		conn.SetReadDeadline(time.Now())
	}()

	var wg sync.WaitGroup
	buf := make([]byte, 4096)
	for {
		n, from, err := conn.ReadFromUDP(buf)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			log.Println("read error:", err)
			continue
		}
		query := make([]byte, n)
		copy(query, buf[:n])

		wg.Add(1)
		go func() {
			defer wg.Done()
			handle(conn, from, query, *upstream)
		}()
	}

	log.Println("shutting down, waiting for queries...")
	wg.Wait()
	log.Println("bye")
}

func handle(conn *net.UDPConn, from *net.UDPAddr, msg []byte, upstream string) {
	q, err := parseQuestion(msg)
	if err != nil {
		log.Println("parse error:", err)
	} else {
		log.Printf("%s → %s (%s)", from.IP, q.Name, q.Type)
	}

	resp, err := exchange(upstream, msg)
	if err != nil {
		log.Println("exchange error:", err)
		return
	}

	_, err = conn.WriteToUDP(resp, from)
	if err != nil {
		log.Println("write error:", err)
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

func parseQuestion(b []byte) (dnsmessage.Question, error) {
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
