package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"time"

	"strconv"
	"strings"
)

var agentpassword string

type AppOptions struct {
	usetls          bool
	verify          bool
	usewebsocket    bool
	useragent       string
	envproxy        bool
	debug           bool
	autocert        string
	recn            int
	rect            int
	optquiet        bool
	proxyauthstring string
	optproxytimeout string
	optdnslisten    string
	optdnsdelay     string
	optdnsdomain    string
	listen          string
	certificate     string
	socks           string
	connect         string
	proxy           string
	agentpassword   string
}

var CurOptions AppOptions

func main() {
	flag.StringVar(&CurOptions.listen, "listen", "", "listen port for receiver address:port")
	flag.StringVar(&CurOptions.certificate, "cert", "", "certificate file")
	flag.StringVar(&CurOptions.socks, "socks", "127.0.0.1:1080", "socks address:port")
	flag.StringVar(&CurOptions.connect, "connect", "", "connect address:port (or https://address:port for ws)")
	flag.StringVar(&CurOptions.proxy, "proxy", "", "use proxy address:port for connecting (or http://address:port for ws)")
	flag.StringVar(&CurOptions.optdnslisten, "dnslisten", "", "Where should DNS server listen")
	flag.StringVar(&CurOptions.optdnsdelay, "dnsdelay", "", "Delay/sleep time between requests (200ms by default)")
	flag.StringVar(&CurOptions.optdnsdomain, "dns", "", "DNS domain to use for DNS tunneling")
	flag.StringVar(&CurOptions.optproxytimeout, "proxytimeout", "", "proxy response timeout (ms)")
	flag.StringVar(&CurOptions.proxyauthstring, "proxyauth", "", "proxy auth Domain/user:Password")
	flag.StringVar(&CurOptions.autocert, "autocert", "", "use domain.tld and automatically obtain TLS certificate")
	flag.StringVar(&CurOptions.useragent, "agent", "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko", "User agent to use")
	flag.StringVar(&CurOptions.agentpassword, "pass", "", "Connect password")
	flag.BoolVar(&CurOptions.optquiet, "q", false, "Be quiet - do not display output")
	flag.IntVar(&CurOptions.recn, "recn", 3, "reconnection limit")
	flag.IntVar(&CurOptions.rect, "rect", 30, "reconnection delay")
	flag.BoolVar(&CurOptions.debug, "debug", false, "display debug info")
	flag.BoolVar(&CurOptions.usetls, "tls", false, "use TLS for connection")
	flag.BoolVar(&CurOptions.usewebsocket, "ws", false, "use websocket for connection")
	flag.BoolVar(&CurOptions.verify, "verify", false, "verify TLS connection")
	version := flag.Bool("version", false, "version information")

	flag.Usage = func() {

		fmt.Printf("revsocks - reverse socks5 server/client by kost %s (%s)\n", Version, CommitID)
		fmt.Println("")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Usage (standard tcp):")
		fmt.Println("1) Start on the client: revsocks -listen :8080 -socks 127.0.0.1:1080 -pass test -tls")
		fmt.Println("2) Start on the server: revsocks -connect client:8080 -pass test -tls")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the client with any socks5 client.")
		fmt.Println("Usage (dns):")
		fmt.Println("1) Start on the DNS server: revsocks -dns example.com -dnslisten :53 -socks 127.0.0.1:1080")
		fmt.Println("2) Start on the target: revsocks -dns example.com -pass <paste-generated-key>")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the DNS server with any socks5 client.")
	}

	flag.Parse()

	setFlags := make(map[string]bool)
	flag.Visit(func(f *flag.Flag) {
		setFlags[f.Name] = true
	})

	flag.VisitAll(func(f *flag.Flag) {
		if !setFlags[f.Name] {
			envName := "REVSOCKS_" + strings.ToUpper(f.Name)
			if envVal, ok := os.LookupEnv(envName); ok {
				f.Value.Set(envVal)
			}
		}
	})

	if CurOptions.optquiet {
		log.SetOutput(ioutil.Discard)
	}

	if *version {
		fmt.Printf("revsocks - reverse socks5 server/client %s (%s)\n", Version, CommitID)
		os.Exit(0)
	}

	if CurOptions.agentpassword == "" {
		agentpassword = RandString(64)
		log.Printf("No password specified. Generated password is %s", agentpassword)
	} else {
		agentpassword = CurOptions.agentpassword
	}

	if CurOptions.listen != "" {
		log.Println("Starting to listen for clients")
		if CurOptions.optproxytimeout != "" {
			opttimeout, _ := strconv.Atoi(CurOptions.optproxytimeout)
			proxytout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytout = time.Millisecond * 1000
		}

		//listenForSocks(*listen, *certificate)
		if CurOptions.usewebsocket {
			log.Fatal(listenForWebsocketAgents(CurOptions.usetls, CurOptions.listen, CurOptions.socks, CurOptions.certificate, CurOptions.autocert))
		} else {
			log.Fatal(listenForAgents(CurOptions.usetls, CurOptions.listen, CurOptions.socks, CurOptions.certificate, CurOptions.autocert))
		}
	}

	if CurOptions.connect != "" {

		if CurOptions.optproxytimeout != "" {
			opttimeout, _ := strconv.Atoi(CurOptions.optproxytimeout)
			proxytimeout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytimeout = time.Millisecond * 1000
		}

		if CurOptions.proxyauthstring != "" {
			if strings.Contains(CurOptions.proxyauthstring, "/") {
				domain = strings.Split(CurOptions.proxyauthstring, "/")[0]
				username = strings.Split(strings.Split(CurOptions.proxyauthstring, "/")[1], ":")[0]
				password = strings.Split(strings.Split(CurOptions.proxyauthstring, "/")[1], ":")[1]
			} else {
				username = strings.Split(CurOptions.proxyauthstring, ":")[0]
				password = strings.Split(CurOptions.proxyauthstring, ":")[1]
			}
			log.Printf("Using domain %s with %s:%s", domain, username, password)
		} else {
			domain = ""
			username = ""
			password = ""
		}

		//log.Fatal(connectForSocks(*connect,*proxy))
		if CurOptions.recn > 0 {
			for i := 1; i <= CurOptions.recn; i++ {
				log.Printf("Connecting to the far end. Try %d of %d", i, CurOptions.recn)
				if CurOptions.usewebsocket {
					WSconnectForSocks(CurOptions.verify, CurOptions.connect, CurOptions.proxy)
				} else {
					error1 := connectForSocks(CurOptions.usetls, CurOptions.verify, CurOptions.connect, CurOptions.proxy)
					log.Print(error1)
				}
				log.Printf("Sleeping for %d sec...", CurOptions.rect)
				tsleep := time.Second * time.Duration(CurOptions.rect)
				time.Sleep(tsleep)
			}

		} else {
			for {
				log.Printf("Reconnecting to the far end... ")
				if CurOptions.usewebsocket {
					WSconnectForSocks(CurOptions.verify, CurOptions.connect, CurOptions.proxy)
				} else {
					error1 := connectForSocks(CurOptions.usetls, CurOptions.verify, CurOptions.connect, CurOptions.proxy)
					log.Print(error1)
				}
				log.Printf("Sleeping for %d sec...", CurOptions.rect)
				tsleep := time.Second * time.Duration(CurOptions.rect)
				time.Sleep(tsleep)
			}
		}

		log.Fatal("Ending...")
	}

	if CurOptions.optdnsdomain != "" {
		dnskey := CurOptions.agentpassword
		if CurOptions.agentpassword == "" {
			dnskey = GenerateKey()
			log.Printf("No password specified, generated following (recheck if same on both sides): %s", dnskey)
		}
		if len(dnskey) != 64 {
			fmt.Fprintf(os.Stderr, "Specified key of incorrect size for DNS (should be 64 in hex)\n")
			os.Exit(1)
		}
		if CurOptions.optdnslisten != "" {
			ServeDNS(CurOptions.optdnslisten, CurOptions.optdnsdomain, CurOptions.socks, dnskey, CurOptions.optdnsdelay)
		} else {
			DnsConnectSocks(CurOptions.optdnsdomain, dnskey, CurOptions.optdnsdelay)
		}
		log.Fatal("Ending...")
	}

	flag.Usage()
	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address\n")
	os.Exit(1)
}
