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

	"context"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
)

var proxytout = time.Millisecond * 1000 //timeout for wait magicbytes

type agentHandler struct {
	mu    sync.Mutex
	listenstr string // listen string for clients
	portnext int // next port for listen
	timeout time.Duration
	sessions []*yamux.Session // all sessions
	// agentstr string // connecting agent combo (IP:port)
}

func (h *agentHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var session *yamux.Session
	var erry error

	agentstr := r.RemoteAddr
	log.Printf("[%s] Got HTTP request (%s):  %s", agentstr, r.Method, r.URL.String())

	if r.Header.Get("Upgrade") != "websocket" {
		w.Header().Set("Location", "https://www.microsoft.com/")
		w.WriteHeader(http.StatusFound) // Use 302 status code for redirect
		// fmt.Fprintf(w, "OK")
		return
	}

	c, err := websocket.Accept(w, r, nil)
	if err != nil {
		log.Printf("[%s] Error upgrading to socket (%s):  %v", agentstr, r.RemoteAddr, err)
		http.Error(w, "Bad request - Go away!", 500)
		return
	}
	defer c.CloseNow()

	if h.timeout>0 {
		_, cancel := context.WithTimeout(r.Context(), time.Second*60)
		defer cancel()
	}

	nc_over_ws := websocket.NetConn(context.Background(), c, websocket.MessageBinary)

	//Add connection to yamux
	session, erry = yamux.Client(nc_over_ws, nil)
	if erry != nil {
		log.Printf("[%s] Error creating client in yamux for (%s): %v", agentstr, r.RemoteAddr, erry)
		http.Error(w, "Bad request - Go away!", 500)
		return
	}
	h.sessions = append(h.sessions, session)
	h.mu.Lock()
	listenport := h.portnext
	h.portnext = h.portnext + 1
	h.mu.Unlock()
	listenForClients(agentstr, h.listenstr, listenport, session)

	c.Close(websocket.StatusNormalClosure, "")
}

func setupHTTP (tlslisten bool, address string, clients string, certificate string) error {
	var cer tls.Certificate
	var err error
	log.Printf("Will start listening for clients on %s", clients)
	var listenstr = strings.Split(clients, ":")
	portnum, errc := strconv.Atoi(listenstr[1])
	if errc != nil {
		log.Printf("Error converting listen str %s: %v", clients, errc)
	}

	aHandler := &agentHandler{
		portnext: portnum,
		listenstr: listenstr[0],
	}
	server := &http.Server{
		Addr:    address, // e.g. ":8443"
		Handler: aHandler,
	}
	if tlslisten {
		log.Printf("Listening for websocket agents on %s using TLS", address)
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
		// config := &tls.Config{Certificates: []tls.Certificate{cer}}
		server.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cer},
		}
	}

	if tlslisten {
		err = server.ListenAndServeTLS("", "")
	} else {
		err = server.ListenAndServe()
	}

	return nil
}

func listenForWebsocketAgents(tlslisten bool, address string, clients string, certificate string) error {
	return setupHTTP(tlslisten, address, clients, certificate)
}

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
