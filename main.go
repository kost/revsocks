package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"io/ioutil"

	"time"

	"strconv"
	"strings"
)

var agentpassword string
var socksdebug bool

func main() {

	listen := flag.String("listen", "", "listen port for receiver address:port")
	certificate := flag.String("cert", "", "certificate file")
	socks := flag.String("socks", "127.0.0.1:1080", "socks address:port")
	connect := flag.String("connect", "", "connect address:port")
	proxy := flag.String("proxy", "", "proxy address:port")
	optdnslisten := flag.String("dnslisten", "", "Where should DNS server listen")
	optdnsdelay := flag.String("dnsdelay", "", "Delay/sleep time between requests (200ms by default)")
	optdnsdomain := flag.String("dns", "", "DNS domain to use for DNS tunneling")
	optproxytimeout := flag.String("proxytimeout", "", "proxy response timeout (ms)")
	proxyauthstring := flag.String("proxyauth", "", "proxy auth Domain/user:Password ")
	optuseragent := flag.String("useragent", "", "User-Agent")
	optpassword := flag.String("pass", "", "Connect password")
	optquiet := flag.Bool("q",false,"Be quiet")
	recn := flag.Int("recn", 3, "reconnection limit")

	rect := flag.Int("rect", 30, "reconnection delay")
	fsocksdebug := flag.Bool("debug", false, "display debug info")
	version := flag.Bool("version", false, "version information")
	flag.Usage = func() {

		fmt.Println("revsocks - reverse socks5 server/client")
		fmt.Println("")
		flag.PrintDefaults()
		fmt.Println("")
		fmt.Println("Usage (standard tcp):")
		fmt.Println("1) Start on the client: revsocks -listen :8080 -socks 127.0.0.1:1080 -pass test")
		fmt.Println("2) Start on the server: revsocks -connect client:8080 -pass test")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the client with any socks5 client.")
		fmt.Println("Usage (dns):")
		fmt.Println("1) Start on the DNS server: revsocks -dns example.com -dnslisten :53 -socks 127.0.0.1:1080")
		fmt.Println("2) Start on the target: revsocks -dns example.com -pass <paste-generated-key>")
		fmt.Println("3) Connect to 127.0.0.1:1080 on the DNS server with any socks5 client.")
	}

	flag.Parse()

	if *optquiet {
		log.SetOutput(ioutil.Discard)
	}

	if *fsocksdebug {
		socksdebug = true
	}
	if *version {
		fmt.Println("revsocks - reverse socks5 server/client")
		os.Exit(0)
	}

	if *listen != "" {
		log.Println("Starting to listen for clients")
		if *optproxytimeout != "" {
			opttimeout, _ := strconv.Atoi(*optproxytimeout)
			proxytout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytout = time.Millisecond * 1000
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = RandString(64)
			log.Println("No password specified. Generated password is " + agentpassword)
		}

		//listenForSocks(*listen, *certificate)
		log.Fatal(listenForAgents(true, *listen, *socks, *certificate))
	}

	if *connect != "" {

		if *optproxytimeout != "" {
			opttimeout, _ := strconv.Atoi(*optproxytimeout)
			proxytimeout = time.Millisecond * time.Duration(opttimeout)
		} else {
			proxytimeout = time.Millisecond * 1000
		}

		if *proxyauthstring != "" {
			if strings.Contains(*proxyauthstring, "/") {
				domain = strings.Split(*proxyauthstring, "/")[0]
				username = strings.Split(strings.Split(*proxyauthstring, "/")[1], ":")[0]
				password = strings.Split(strings.Split(*proxyauthstring, "/")[1], ":")[1]
			} else {
				username = strings.Split(*proxyauthstring, ":")[0]
				password = strings.Split(*proxyauthstring, ":")[1]
			}
			log.Printf("Using domain %s with %s:%s", domain, username, password)
		} else {
			domain = ""
			username = ""
			password = ""
		}

		if *optpassword != "" {
			agentpassword = *optpassword
		} else {
			agentpassword = "RocksDefaultRequestRocksDefaultRequestRocksDefaultRequestRocks!!"
		}

		if *optuseragent != "" {
			useragent = *optuseragent
		} else {
			useragent = "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko"
		}
		//log.Fatal(connectForSocks(*connect,*proxy))
		if *recn > 0 {
			for i := 1; i <= *recn; i++ {
				log.Printf("Connecting to the far end. Try %d of %d", i, *recn)
				error1 := connectForSocks(true, *connect, *proxy)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}

		} else {
			for {
				log.Printf("Reconnecting to the far end... ")
				error1 := connectForSocks(true, *connect, *proxy)
				log.Print(error1)
				log.Printf("Sleeping for %d sec...", *rect)
				tsleep := time.Second * time.Duration(*rect)
				time.Sleep(tsleep)
			}
		}

		log.Fatal("Ending...")
	}

	if *optdnsdomain != "" {
		dnskey:=*optpassword
		if *optpassword == "" {
			dnskey=GenerateKey()
			log.Printf("No password specified, generated following (recheck if same on both sides): %s", dnskey)
		}
		if len(dnskey) != 64 {
			fmt.Fprintf(os.Stderr, "Specified key of incorrect size for DNS (should be 64 in hex)\n")
			os.Exit(1)
		}
		if *optdnslisten != "" {
			ServeDNS (*optdnslisten,*optdnsdomain,*socks, dnskey, *optdnsdelay)
		} else {
			DnsConnectSocks(*optdnsdomain, dnskey, *optdnsdelay)
		}
		log.Fatal("Ending...")
	}

	flag.Usage()
	fmt.Fprintf(os.Stderr, "You must specify a listen port or a connect address\n")
	os.Exit(1)
}
