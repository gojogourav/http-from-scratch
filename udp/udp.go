package main

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
)

func getLinesChannel(conn *net.UDPConn) <-chan string {
	out := make(chan string)

	go func() {
		defer conn.Close()
		defer close(out)

		for {
			buffer := make([]byte, 1024)
			n, _, err := conn.ReadFromUDP(buffer)
			if err != nil {
				log.Println("Error reading udp : ", err)
				break
			}

			data := bytes.NewReader(buffer[:n])
			scanner := bufio.NewScanner(data)

			for scanner.Scan() {
				out <- scanner.Text()
			}

			if err := scanner.Err(); err != nil {
				log.Println("Scanner error:", err)
				break
			}
		}
	}()
	return out
}

func main() {
	addr, err := net.ResolveUDPAddr("udp", "localhost:8080")
	if err != nil {
		log.Fatalf("Failed to resolve UDP address: %v", err)
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Fatalf("Failed to listen on UDP: %v", err)
	}
	defer conn.Close()

	fmt.Println("UDP Server listening on localhost:8080...")

	lines := getLinesChannel(conn)
	for line := range lines {
		fmt.Println("Received:", line)
	}
}
