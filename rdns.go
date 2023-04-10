package main

import (
	"log"

	"github.com/kost/dnstun"

	socks5 "github.com/armon/go-socks5"
)

func GenerateKey() string {
	return dnstun.GenerateKey()
}

func DnsConnectSocks(targetDomain string, encryptionKey string, dnsdelay string) error {
	server, err := socks5.New(&socks5.Config{})
	if err != nil {
		log.Printf("Error socks5.new:  %v", err)
		return err
	}
	dt:=dnstun.NewDnsTunnel(targetDomain,encryptionKey)
	if dnsdelay != "" {
		err=dt.SetDnsDelay(dnsdelay)
		if err != nil {
			log.Printf("Error setting delay:  %v", err)
			return err
		}
	}
	for {
		session, err := dt.DnsClient()
		if err != nil {
			log.Printf("Error yamux transport:  %v", err)
			return err
		}
		for {
			stream, err := session.Accept()
			log.Println("Accepting stream")
			if err != nil {
				log.Printf("Error accepting stream:  %v", err)
				break
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
}

func ServeDNS (dnslisten string, DnsDomain string, clients string, enckey string, dnsdelay string) error {
	dt:=dnstun.NewDnsTunnel(DnsDomain,enckey)
	if dnsdelay != "" {
		err:=dt.SetDnsDelay(dnsdelay)
		if err != nil {
			log.Printf("Error parsing DNS delay/sleep duration %s: %v", dnsdelay, err)
			return err
		}
	}
	dt.DnsServer(dnslisten, clients)
	err:=dt.DnsServerStart()
	if err != nil {
		log.Printf("Error starting DNS server %s: %v", DnsDomain, err)
		return err
	}
	return nil
}

