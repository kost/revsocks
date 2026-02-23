[![Codacy Badge](https://api.codacy.com/project/badge/Grade/3c687bcd445e4a828c914e4e2384196e)](https://www.codacy.com/manual/kost/revsocks?utm_source=github.com&utm_medium=referral&utm_content=kost/revsocks&utm_campaign=Badge_Grade)

# revsocks

Reverse socks5 tunneler with SSL/TLS and proxy support (without proxy authentication and with basic/NTLM proxy authentication)
Based on <https://github.com/brimstone/rsocks> and <https://github.com/llkat/rsockstun>

# Features

-   Single executable (thanks to Go!)
-   Linux/Windows/Mac/BSD support
-   Encrypted communication with TLS
-   DNS tunneling support (SOCKS5 over DNS)
-   Support for proxies (without authentication or with basic/NTLM proxy authentication)
-   Automatic SSL/TLS certificate generation if not specified

# Architecture

-   server = locally listening socks server
-   client = client which connects back to server

## Usage

### reverse TCP (TLS is enabled by default)

    Usage:
    1) Start on VPS: revsocks -listen :8443 -socks 127.0.0.1:1080 -pass SuperSecretPassword
    2) Start on client: revsocks -connect clientIP:8443 -pass SuperSecretPassword
    3) Connect to 127.0.0.1:1080 on the VPS with any socks5 client.
    4) Enjoy. :]

### reverse TCP without TLS encryption (Plaintext)

    Usage:
    1) Start on VPS: revsocks -listen :8443 -socks 127.0.0.1:1080 -pass SuperSecretPassword -tls=false
    2) Start on client: revsocks -connect clientIP:8443 -pass SuperSecretPassword -tls=false
    3) Connect to 127.0.0.1:1080 on the VPS with any socks5 client.
    4) Enjoy. :]

### reverse websocket with TLS encryption

    Usage:
    1) Start on VPS: `revsocks -listen :8443 -socks 127.0.0.1:1080 -pass SuperSecretPassword -ws`
    2) Start on client: `revsocks -connect https://clientIP:8443 -pass SuperSecretPassword -ws`
    3) Connect to 127.0.0.1:1080 on the VPS with any socks5 client.

### DNS tunnel

```sh
0) setup your domain records
1) Start on the DNS server: revsocks -dns example.com -dnslisten :53 -socks 127.0.0.1:1080 -pass 52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c64
2) Start on the target: revsocks -dns example.com -pass 52fdfc072182654f163f5f0f9a621d729566c74d10037c4d7bbb0407d1e2c64
3) Connect to 127.0.0.1:1080 on the DNS server with any socks5 client.
```

## Useful parameters

    Add params:
     -proxy 1.2.3.4:3128 - connect via proxy
     -proxyauth Domain/username:password  - proxy creds
     -proxytimeout 2000 - server and clients will wait for 2000 msec for proxy connections... (Sometime it should be up to 4000...)
     -useragent "Internet Explorer 9.99" - User-Agent used in proxy connection (sometimes it is usefull)
     -pass Password12345 - challenge password between client and server (if not match - server reply 301 redirect)
     -recn - reconnect times number. Default is 3. If 0 - infinite reconnection
     -rect - time delay in secs between reconnection attempts. Default is 30

## Options

Complete list of command line options. Any option can also be set via an environment variable by prefixing the option name with `REVSOCKS_` and converting it to uppercase (e.g., `-listen` becomes `REVSOCKS_LISTEN`). Command-line options have higher priority and will override environment variables.


```
  -agent string
    	User agent to use (default "Mozilla/5.0 (Windows NT 6.1; Trident/7.0; rv:11.0) like Gecko")
  -cert string
    	certificate file
  -connect string
    	connect address:port (or https://address:port for ws)
  -debug
    	display debug info
  -dns string
    	DNS domain to use for DNS tunneling
  -dnsdelay string
    	Delay/sleep time between requests (200ms by default)
  -dnslisten string
    	Where should DNS server listen
  -listen string
    	listen port for receiver address:port
  -pass string
    	Connect password
  -proxy string
    	use proxy address:port for connecting (or http://address:port for ws)
  -proxyauth string
    	proxy auth Domain/user:Password
  -proxytimeout string
    	proxy response timeout (ms)
  -q	Be quiet
  -recn int
    	reconnection limit (default 3)
  -rect int
    	reconnection delay (default 30)
  -socks string
    	socks address:port (default "127.0.0.1:1080")
  -tls
    	use TLS for connection
  -verify
    	verify TLS connection
  -version
    	version information
  -ws
    	use websocket for connection
```

# Requirements

-   Go 1.4 or higher
-   Few external Go modules (yamux, go-socks5 and go-ntlmssp)

# Compile and Installation

Linux VPS

-   install Golang: apt install golang make

```sh
make
```

launch:

```sh
./revsocks -listen :8443 -socks 127.0.0.1:1080 -pass Password1234
```

Windows client:

-   download and install golang

```sh
go get
go build
```

## Windows optional

optional: to build as Windows GUI:

```sh
go build -ldflags -H=windowsgui
```

You can also compress exe - just use any exe packer, ex: UPX

```sh
upx revsocks
```

## Usage examples

```sh
revsocks -connect clientIP:8443 -pass Password1234
```

or with proxy and user agent:

```sh
revsocks -connect clientIP:8443 -pass Password1234 -proxy proxy.domain.local:3128 -proxyauth Domain/userpame:userpass -useragent "Mozilla 5.0/IE Windows 10"
```

Client connects to server and send agentpassword to authorize on server. If server does not receive agentpassword or reveive wrong pass from client (for example if spider or client browser connects to server ) then it send HTTP 301 redirect code to www.microsoft.com

## Custom certificate

Generate self-signed certificate with openssl:

```sh
openssl req -new -x509 -keyout server.key -out server.crt -days 365 -nodes
```

## Debug

For debugging (especially DNS part):
```sh
go build -tags debug
```
