package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"bufio"
	"time"
	//"encoding/hex"
	"github.com/hashicorp/yamux"
	"strings"
	"strconv"
)

var proxytout = time.Millisecond * 1000 //timeout for wait magicbytes
// Catches yamux connecting to us
func listenForSocks(address string, clients string, certificate string) error {
	var err, erry error
	var cer tls.Certificate
	var session *yamux.Session
	var sessions []*yamux.Session

	if certificate == "" {
		cer, err = getRandomTLS(2048)
		log.Println("No TLS certificate. Generated random one.")
	} else {
		cer, err = tls.LoadX509KeyPair(certificate+".crt", certificate+".key")
	}
	if err != nil {
		log.Println(err)
		return err
	}
	log.Printf("Listening for agents on %s", address)
	log.Printf("Will start listening for clients on %s", clients)
	config := &tls.Config{Certificates: []tls.Certificate{cer}}
	//ln, err := net.Listen("tcp", address)
	ln, err := tls.Listen("tcp", address, config)
	if err != nil {
		return err
	}
	var listenstr = strings.Split(clients, ":")
	portnum, errc := strconv.Atoi(listenstr[1])
	if errc != nil {
		log.Printf("Error converting listen str %s: %v", clients, errc)
	}
	portinc := 0
	for {
		conn, err := ln.Accept()
		conn.RemoteAddr()
		log.Printf("Got a SSL connection from %v: ", conn.RemoteAddr())
		if err != nil {
			fmt.Fprintf(os.Stderr, "Errors accepting!")
		}

		reader := bufio.NewReader(conn)

		//read only 64 bytes with timeout=1-3 sec. So we haven't delay with browsers
		conn.SetReadDeadline(time.Now().Add(proxytout))
		statusb := make([]byte, 64)
		_, _ = io.ReadFull(reader, statusb)

		//Alternatively  - read all bytes with timeout=1-3 sec. So we have delay with browsers, but get all GET request
		//conn.SetReadDeadline(time.Now().Add(proxytout))
		//statusb,_ := ioutil.ReadAll(magicBuf)

		//log.Printf("magic bytes: %v",statusb[:6])
		//if hex.EncodeToString(statusb) != magicbytes {
		if string(statusb)[:len(agentpassword)] != agentpassword {
			//do HTTP checks
			log.Printf("Received request: %v", string(statusb[:64]))
			status := string(statusb)
			if strings.Contains(status, " HTTP/1.1") {
				httpresonse := "HTTP/1.1 301 Moved Permanently" +
					"\r\nContent-Type: text/html; charset=UTF-8" +
					"\r\nLocation: https://www.microsoft.com/" +
					"\r\nServer: Apache" +
					"\r\nContent-Length: 0" +
					"\r\nConnection: close" +
					"\r\n\r\n"

				conn.Write([]byte(httpresonse))
				conn.Close()
			} else {
				conn.Close()
			}

		} else {
			//magic bytes received.
			//disable socket read timeouts
			log.Printf("Got Client from %s", conn.RemoteAddr())
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))
			listen4clients := fmt.Sprintf("%s:%d",listenstr[0],portnum+portinc)
			log.Printf("Built listen string %s", listen4clients)
			//Add connection to yamux
			session, erry = yamux.Client(conn, nil)
			if erry != nil {
				log.Printf("Error creating client in yamux for %s: %v", conn.RemoteAddr(), erry)
				continue
			}
			sessions=append(sessions,session)
			go listenForClients(listen4clients, session)
			portinc = portinc + 1
		}
	}
	return nil
}

// Catches clients and connects to yamux
func listenForClients(address string, session *yamux.Session) error {
	log.Printf("Waiting for clients on %s", address)
	ln, err := net.Listen("tcp", address)
	if err != nil {
		return err
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		// TODO dial socks5 through yamux and connect to conn

		if session == nil {
			conn.Close()
			continue
		}
		log.Printf("Got a client for %s", conn.RemoteAddr())

		log.Printf("Opening a stream for %s", conn.RemoteAddr())
		stream, err := session.Open()
		if err != nil {
			log.Printf("Error opening stream for %s: %v", conn.RemoteAddr(), err)
			return err
		}

		// connect both of conn and stream

		go func() {
			log.Printf("Starting to copy conn to stream for %s", conn.RemoteAddr())
			io.Copy(conn, stream)
			conn.Close()
		}()
		go func() {
			log.Printf("Starting to copy stream to conn for %s", conn.RemoteAddr())
			io.Copy(stream, conn)
			stream.Close()
			log.Printf("Done copying stream to conn for %s", conn.RemoteAddr())
		}()
	}
}
