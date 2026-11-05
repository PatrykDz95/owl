package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/PatrykDz95/owl/internal/device"
	"github.com/PatrykDz95/owl/internal/dns"
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
	q, qErr := dns.ParseQuestion(msg)

	resp, err := dns.Exchange(upstream, msg)
	if err != nil {
		log.Println("exchange error:", err)
		return
	}

	_, err = conn.WriteToUDP(resp, from)
	if err != nil {
		log.Println("write error:", err)
	}

	macStr := "?"
	mac, err := device.LookupMAC(from.IP)
	if err == nil {
		macStr = fmt.Sprintf("%s, %s", mac, device.Vendor(mac))
	}

	if qErr != nil {
		log.Println("parse error:", qErr)
		return
	}
	log.Printf("%s (%s) → %s (%s)", from.IP, macStr, q.Name, q.Type)
}
