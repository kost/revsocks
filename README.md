[![Build Status](https://travis-ci.org/kost/revsocks.png)](https://travis-ci.org/kost/revsocks)
[![CircleCI](https://circleci.com/gh/kost/revsocks.svg?style=svg)](https://circleci.com/gh/kost/revsocks)
[![Codacy Badge](https://api.codacy.com/project/badge/Grade/3c687bcd445e4a828c914e4e2384196e)](https://www.codacy.com/manual/kost/revsocks?utm_source=github.com&amp;utm_medium=referral&amp;utm_content=kost/revsocks&amp;utm_campaign=Badge_Grade)

revsocks
========

Reverse socks5 tunneler with SSL/TLS and proxy support (without proxy authentication and with basic/NTLM proxy authentication)
Based on https://github.com/brimstone/rsocks and https://github.com/llkat/rsockstun

Features
========

- Single executable (thanks to Go!)
- Linux/Windows/Mac/BSD support
- Encrypted communication with TLS
- Support for proxies (without authentication or with basic/NTLM proxy authentication)
- Automatic SSL/TLS certificate generation if not specified

Modes of operation
==================

- server = locally listening socks server
- client = client which connects back to server

Usage
------
```
Usage:
1) Start on VPS: revsocks -listen :8443 -socks 127.0.0.1:1080 -pass SuperSecretPassword
2) Start on client: revsocks -connect clientIP:8443 -pass SuperSecretPassword
3) Connect to 127.0.0.1:1080 on the VPS with any socks5 client.
4) Enjoy. :]

Add params:
 -proxy 1.2.3.4:3128 - connect via proxy
 -proxyauth Domain/username:password  - proxy creds
 -proxytimeout 2000 - server and clients will wait for 2000 msec for proxy connections... (Sometime it should be up to 4000...)
 -useragent "Internet Explorer 9.99" - User-Agent used in proxy connection (sometimes it is usefull)
 -pass Password12345 - challenge password between client and server (if not match - server reply 301 redirect)
 -recn - reconnect times number. Default is 3. If 0 - infinite reconnection
 -rect - time delay in secs between reconnection attempts. Default is 30
```

Requirements
============

- Go 1.4 or higher
- Few external Go modules (yamux, go-socks5 and go-ntlmssp)

Compile and Installation
=====

Linux VPS
- install Golang: apt install golang
```sh
export GOPATH=~/go
go get github.com/hashicorp/yamux
go get github.com/armon/go-socks5
go get github.com/kost/go-ntlmssp
go build
```
launch:
```sh
./revsocks -listen :8443 -socks 127.0.0.1:1080 -pass Password1234
```

Windows client:
- download and install golang
```sh
go get github.com/hashicorp/yamux
go get github.com/armon/go-socks5
go get github.com/kost/go-ntlmssp
go build
```
optional: to build as Windows GUI: go build -ldflags -H=windowsgui
optional: to compress exe - use any exe packer, ex: UPX
launch:
```sh
revsocks -connect clientIP:8443 -pass Password1234
```

or with proxy and user agent:
```sh
revsocks -connect clientIP:8443 -pass Password1234 -proxy proxy.domain.local:3128 -proxyauth Domain/userpame:userpass -useragent "Mozilla 5.0/IE Windows 10"
```

Client connects to server and send agentpassword to authorize on server. If server does not receive agentpassword or reveive wrong pass from client (for example if spider or client browser connects to server ) then it send HTTP 301 redirect code to www.microsoft.com

Generate self-signed certificate with openssl:

```sh
openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes
```
