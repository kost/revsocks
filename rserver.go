package main

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"bufio"
	"github.com/hashicorp/yamux"
	"strconv"
	"strings"
	"time"
)

var proxytout = time.Millisecond * 1000 //timeout for wait magicbytes

// listen for agents
func listenForAgents(tlslisten bool, address string, clients string, certificate string) error {
	var err, erry error
	var cer tls.Certificate
	var session *yamux.Session
	var sessions []*yamux.Session
	var ln net.Listener

	log.Printf("Will start listening for clients on %s", clients)
	if tlslisten {
		log.Printf("Listening for agents on %s using TLS", address)
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
		config := &tls.Config{Certificates: []tls.Certificate{cer}}
		ln, err = tls.Listen("tcp", address, config)
	} else {
		log.Printf("Listening for agents on %s", address)
		ln, err = net.Listen("tcp", address)
	}
	if err != nil {
		log.Printf("Error listening on %s: %v", address, err)
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
		agentstr := conn.RemoteAddr().String()
		log.Printf("[%s] Got a connection from %v: ", agentstr, conn.RemoteAddr())
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
			log.Printf("[%s] Got Client from %s", agentstr, conn.RemoteAddr())
			conn.SetReadDeadline(time.Now().Add(100 * time.Hour))
			//Add connection to yamux
			session, erry = yamux.Client(conn, nil)
			if erry != nil {
				log.Printf("[%s] Error creating client in yamux for %s: %v", agentstr, conn.RemoteAddr(), erry)
				continue
			}
			sessions = append(sessions, session)
			go listenForClients(agentstr, listenstr[0], portnum+portinc, session)
			portinc = portinc + 1
		}
	}
	return nil
}

// Catches local clients and connects to yamux
func listenForClients(agentstr string, listen string, port int, session *yamux.Session) error {
	var ln net.Listener
	var address string
	var err error
	portinc := port
	for {
		address = fmt.Sprintf("%s:%d", listen, portinc)
		log.Printf("[%s] Waiting for clients on %s", agentstr, address)
		ln, err = net.Listen("tcp", address)
		if err != nil {
			log.Printf("[%s] Error listening on %s: %v", agentstr, address, err)
			portinc = portinc + 1
		} else {
			break
		}
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("[%s] Error accepting on %s: %v", agentstr, address, err)
			return err
		}
		if session == nil {
			log.Printf("[%s] Session on %s is nil", agentstr, address)
			conn.Close()
			continue
		}
		log.Printf("[%s] Got client. Opening stream for %s", agentstr, conn.RemoteAddr())

		stream, err := session.Open()
		if err != nil {
			log.Printf("[%s] Error opening stream for %s: %v", agentstr, conn.RemoteAddr(), err)
			return err
		}

		// connect both of conn and stream

		go func() {
			log.Printf("[%s] Starting to copy conn to stream for %s", agentstr, conn.RemoteAddr())
			io.Copy(conn, stream)
			conn.Close()
			log.Printf("[%s] Done copying conn to stream for %s", agentstr, conn.RemoteAddr())
		}()
		go func() {
			log.Printf("[%s] Starting to copy stream to conn for %s", agentstr, conn.RemoteAddr())
			io.Copy(stream, conn)
			stream.Close()
			log.Printf("[%s] Done copying stream to conn for %s", agentstr, conn.RemoteAddr())
		}()
	}
}
