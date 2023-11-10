package main

import (
	"crypto/tls"
	"fmt"
	"log"
	"net"

	"errors"

	"context"
	"net/url"

	"bufio"
	"bytes"
	"encoding/base64"
	socks5 "github.com/armon/go-socks5"
	"github.com/hashicorp/yamux"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	ntlmssp "github.com/kost/go-ntlmssp"

	"nhooyr.io/websocket"
)

var encBase64 = base64.StdEncoding.EncodeToString
var decBase64 = base64.StdEncoding.DecodeString
var username string
var domain string
var password string
var connectproxystring string
var useragent string
var proxytimeout = time.Millisecond * 1000 //timeout for proxyserver response

func WSconnectForSocks(verify bool, address string, proxy string) error {
	// Define the proxy URL and WebSocket endpoint URL
	proxyURL := proxy   // Change this to your proxy URL
	wsURL := address // Change this to your WebSocket endpoint

	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Printf("Error setting up socks server: %v", err)
		return err
	}
	httpClient := &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !verify},
		},
	}

	if proxy != "" {
	ntlmssp.NewNegotiateMessage(domain, "")

	// Create an HTTP client that authenticates via NTLMSSP
	negmsg, err := ntlmssp.NewNegotiateMessage(domain, "")
	if err != nil {
		log.Printf("Error getting domain negotiate message: %v", err)
		return err
	}
	httpClient = &http.Client{
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: !verify},
			Proxy: http.ProxyURL(mustParseURL(proxyURL)),
			ProxyConnectHeader: http.Header{
				"Proxy-Authorization": []string{string(negmsg)},
			},
		},
	}

	// resp, err := http.Get(wsURL)
	// resp, err := httpClient.Get(wsURL)
	req, err := http.NewRequest("GET", wsURL, nil)
	if err != nil {
		log.Printf("error creating http request to %s: %s\n", wsURL, err)
		return err
	}

	req.Header.Set("User-Agent", useragent)

	resp, err := httpClient.Do(req)
	if err != nil {
		log.Printf("error making http request to %s: %s\n", wsURL, err)
		return err
	}

	if resp.StatusCode == 200 {
		log.Printf("No proxy auth required. Will make standard request", resp.StatusCode)
	} else if resp.StatusCode == 407 {
		ntlmchall := resp.Header.Get("Proxy-Authenticate")
		log.Printf("Got following challenge: %s", ntlmchall)
		if strings.Contains(ntlmchall, "NTLM") {
			ntlmchall = ntlmchall[5:]
			challengeMessage, errb := base64.StdEncoding.DecodeString(ntlmchall)
			if errb != nil {
				log.Printf("Error getting base64 decode of challengde: %v", errb)
				return errb
			}
			authenticateMessage, erra := ntlmssp.ProcessChallenge(challengeMessage, username, password)
			if erra != nil {
				log.Printf("Error getting auth message for challenge: %v", erra)
			}
			authMessage := fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticateMessage))
			httpClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(mustParseURL(proxyURL)),
					ProxyConnectHeader: http.Header{
						"Proxy-Authorization": []string{string(authMessage)},
					},
				},
			}
		} else if strings.Contains(ntlmchall, "Basic") {
			authCombo := fmt.Sprintf("%s:%s", username, password)
			authMessage := fmt.Sprintf("Basic %s", base64.StdEncoding.EncodeToString([]byte(authCombo)))
			httpClient = &http.Client{
				Transport: &http.Transport{
					Proxy: http.ProxyURL(mustParseURL(proxyURL)),
					ProxyConnectHeader: http.Header{
						"Proxy-Authorization": []string{authMessage},
					},
				},
			}
		} else {
			log.Printf("Unknown proxy challenge: %s", ntlmchall)
			return errors.New("Unknown proxy challenge")
		}
	} else {
		log.Printf("Unknown http response code: %d", resp.StatusCode)
	}

	}

	// Connect to the WebSocket endpoint via the proxy
	wconn, _, err := websocket.Dial(context.Background(), wsURL, &websocket.DialOptions{
		HTTPClient:   httpClient,
		HTTPHeader:   http.Header{"User-Agent": []string{useragent}, "Accept-Language": []string{agentpassword}, "Sec-WebSocket-Protocol": []string{"chat"}},
		Subprotocols: []string{"chat"},
	})
	if err != nil {
		fmt.Println("Error connecting to the WebSocket:", err)
		return err
	}
	defer wconn.Close(websocket.StatusInternalError, "Connection closed")

	nc_over_ws := websocket.NetConn(context.Background(), wconn, websocket.MessageBinary)

	session, erry := yamux.Server(nc_over_ws, nil)
	if erry != nil {
		fmt.Println("Error creating yamux server:", err)
		return erry
	}

       for {
		stream, err := session.Accept()
		if err != nil {
			fmt.Println("Error accepting stream:", err)
			continue
		}
		log.Println("Accepted stream")
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Printf("Error serving: %v", err)
			}
		}()
	}
	return nil
}

func mustParseURL(u string) *url.URL {
	parsedURL, err := url.Parse(u)
	if err != nil {
		panic(err)
	}
	return parsedURL
}

func connectviaproxy(proxyaddr string, connectaddr string) net.Conn {
	socksdebug := CurOptions.debug
	connectproxystring = ""
	if (username != "") && (password != "") && (domain != "") {
		negotiateMessage, errn := ntlmssp.NewNegotiateMessage(domain, "")
		if errn != nil {
			log.Println("NEG error")
			log.Println(errn)
			// return nil
		}
		log.Print(negotiateMessage)
		negheader := fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateMessage))

		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: " + useragent +
			"\r\nProxy-Authorization: " + negheader +
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"

	} else {
		connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
			"\r\nUser-Agent: " + useragent +
			"\r\nProxy-Connection: Keep-Alive" +
			"\r\n\r\n"
	}

	if socksdebug {
		log.Print(connectproxystring)
	}

	conn, err := net.Dial("tcp", proxyaddr)
	if err != nil {
		// handle error
		log.Printf("Error connect: %v", err)
	}
	conn.Write([]byte(connectproxystring))

	time.Sleep(proxytimeout) //Because socket does not close - we need to sleep for full response from proxy

	resp, err := http.ReadResponse(bufio.NewReader(conn), &http.Request{Method: "CONNECT"})
	status := resp.Status

	if socksdebug {
		log.Print(status)
		log.Print(resp)
	}

	if (resp.StatusCode == 200) || (strings.Contains(status, "HTTP/1.1 200 ")) ||
		(strings.Contains(status, "HTTP/1.0 200 ")) {
		log.Print("Connected via proxy. No auth required")
		return conn
	}

	if socksdebug {
		log.Print("Checking proxy auth")
	}
	if resp.StatusCode == 407 {
		log.Print("Got Proxy status code (407)")
		ntlmchall := resp.Header.Get("Proxy-Authenticate")
		log.Print(ntlmchall)
		if strings.Contains(ntlmchall, "NTLM") {
			if socksdebug {
				log.Print("Got NTLM challenge:")
				log.Print(ntlmchall)
			}

			/*
				negstring:= fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(negotiateMessage))
				connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
					"\r\nUser-Agent: "+useragent+
					"\r\nProxy-Authorization: " + negstring +
					"\r\nProxy-Connection: Keep-Alive" +
					"\r\n\r\n"
			*/

			ntlmchall = ntlmchall[5:]
			if socksdebug {
				log.Print("NTLM challenge:")
				log.Print(ntlmchall)
			}
			challengeMessage, errb := decBase64(ntlmchall)
			if errb != nil {
				log.Println("BASE64 Decode error")
				log.Println(errb)
				return nil
			}
			authenticateMessage, erra := ntlmssp.ProcessChallenge(challengeMessage, username, password)
			if erra != nil {
				log.Println("Process challenge error")
				log.Println(erra)
				return nil
			}

			authMessage := fmt.Sprintf("NTLM %s", base64.StdEncoding.EncodeToString(authenticateMessage))

			//log.Print(authenticate)
			connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
				"\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko" +
				"\r\nProxy-Authorization: " + authMessage +
				"\r\nProxy-Connection: Keep-Alive" +
				"\r\n\r\n"
		} else if strings.Contains(ntlmchall, "Basic") {
			if socksdebug {
				log.Print("Got Basic challenge:")
			}
			var authbuffer bytes.Buffer
			authbuffer.WriteString(username)
			authbuffer.WriteString(":")
			authbuffer.WriteString(password)

			basicauth := encBase64(authbuffer.Bytes())

			//log.Print(authenticate)
			connectproxystring = "CONNECT " + connectaddr + " HTTP/1.1" + "\r\nHost: " + connectaddr +
				"\r\nUser-Agent: Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko" +
				"\r\nProxy-Authorization: Basic " + basicauth +
				"\r\nProxy-Connection: Keep-Alive" +
				"\r\n\r\n"
		} else {
			log.Print("Unknown authentication")
			return nil
		}
		log.Print("Connecting to proxy")
		log.Print(connectproxystring)
		conn.Write([]byte(connectproxystring))

		//read response
		bufReader := bufio.NewReader(conn)
		conn.SetReadDeadline(time.Now().Add(proxytimeout))
		statusb, _ := ioutil.ReadAll(bufReader)

		status = string(statusb)

		//disable socket read timeouts
		conn.SetReadDeadline(time.Now().Add(100 * time.Hour))

		if (resp.StatusCode == 200 || strings.Contains(status, "HTTP/1.1 200 ") ||
		strings.Contains(status, "HTTP/1.0 200 ")) {
			log.Print("Connected via proxy")
			return conn
		}
		log.Printf("Not Connected via proxy. Status:%v", status)
		return nil

	}
	log.Print("Not connected via proxy")
	conn.Close()
	return nil
}

func connectForSocks(tlsenable bool, verify bool, address string, proxy string) error {
	var session *yamux.Session
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		return err
	}

	conf := &tls.Config{
		InsecureSkipVerify: !verify,
	}

	var conn net.Conn
	var connp net.Conn
	var newconn net.Conn
	//var conntls tls.Conn
	//var conn tls.Conn
	if proxy == "" {
		log.Println("Connecting to far end")
		if tlsenable {
			conn, err = tls.Dial("tcp", address, conf)
		} else {
			conn, err = net.Dial("tcp", address)
		}
		if err != nil {
			return err
		}
	} else {
		log.Println("Connecting to proxy ...")
		connp = connectviaproxy(proxy, address)
		if connp != nil {
			log.Println("Proxy successfull. Connecting to far end")
			if tlsenable {
				conntls := tls.Client(connp, conf)
				err := conntls.Handshake()
				if err != nil {
					log.Printf("Error connect: %v", err)
					return err
				}
				newconn = net.Conn(conntls)
			} else {
				newconn = connp
			}
		} else {
			log.Println("Proxy NOT successfull. Exiting")
			return nil
		}
	}

	log.Println("Starting client")
	if proxy == "" {
		conn.Write([]byte(agentpassword))
		//time.Sleep(time.Second * 1)
		session, err = yamux.Server(conn, nil)
	} else {

		//log.Print(conntls)
		newconn.Write([]byte(agentpassword))
		time.Sleep(time.Second * 1)
		session, err = yamux.Server(newconn, nil)
	}
	if err != nil {
		return err
	}

	for {
		stream, err := session.Accept()
		log.Println("Accepting stream")
		if err != nil {
			return err
		}
		log.Println("Passing off to socks5")
		go func() {
			err = server.ServeConn(stream)
			if err != nil {
				log.Println(err)
			}
		}()
	}
}
