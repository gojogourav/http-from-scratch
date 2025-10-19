package main

import (
	"bytes"
	"fmt"
	"io"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		var lineBuffer []byte

		for {
			buffer := make([]byte, 8)
			n, err := f.Read(buffer)
			if n > 0 {
				data := buffer[:n]
				for {
					index := bytes.IndexByte(data, '\n')
					if index == -1 {
						lineBuffer = append(lineBuffer, data...)

						break
					}

					//
					line := append(lineBuffer, data[:index]...)
					out <- string(line)

					lineBuffer = []byte{}
					data = data[index+1:]

				}

			}

			if err == io.EOF {
				if len(lineBuffer) > 0 {
					out <- string(lineBuffer)
					break

				}
				fmt.Printf("\nEOF reached")
				break
			}

		}

	}()
	return out
}
func main() {
	ln, err := net.Listen("tcp", ":42069")
	if err != nil {
		fmt.Println("Error : ", err)
		return
	}
	fmt.Printf("TCP listening at 42069")
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting:", err)
			continue
		}

		go func(c net.Conn) {
			defer c.Close()

			fmt.Printf("New client connected: %v\n", c.RemoteAddr())
			for line := range getLinesChannel(c) {
				fmt.Println("Received:", line)
			}

			fmt.Printf("Client disconnected: %v\n", c.RemoteAddr())
		}(conn)
	}

	// filepath := "messages.txt"

	// file, err := os.Open(filepath)
	// if err != nil {
	// 	fmt.Printf("Error reading file")
	// 	return
	// }

	// lines := getLinesChannel(file)
	// for line := range lines {
	// 	fmt.Println(line)
	// }
}
