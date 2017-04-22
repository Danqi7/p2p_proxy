package main

import (
    "fmt"
    "log"
    "net"
    "net/url"
    "bytes"
    "strings"
    "io"
)

func main() {
    port := ":3128"
    listener, err := net.Listen("tcp", ":3128")
    if err != nil {
        log.Panic(err)
    }

    log.Printf("Listening at port: %v", port)
    id := 0
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Panic(err)
        }
        log.Println("New client: ", id)
        go handleClient(conn)
    }
}

func handleClient(client net.Conn) {
    if client == nil {
        return
    }
    defer client.Close()
    log.Println("reading client's request")

    //get host and port
    var b [1024]byte
    n, err := client.Read(b[:])
    if err != nil {
        log.Println(err)
        return
    }

    log.Printf("%s", b[:n])

    var method, host, address string
    fmt.Sscanf(string(b[:bytes.IndexByte(b[:], '\n')]), "%s%s", &method, &host)
    hostPortURL, err := url.Parse(host)
	if err != nil {
		log.Println(err)
		return
	}

	if hostPortURL.Opaque == "443" { //https
		address = hostPortURL.Scheme + ":443"
	} else { //http
		if strings.Index(hostPortURL.Host, ":") == -1 { //host defualt port
			address = hostPortURL.Host + ":80"
		} else {
			address = hostPortURL.Host
		}
	}

    log.Println("method, host, address: ", method, host, address)

	//forward the request
	server, err := net.Dial("tcp", address)
	if err != nil {
		log.Println(err)
		return
	}

	if method == "CONNECT" {
		fmt.Fprint(client, "HTTP/1.1 200 Connection established\r\n\r\n")
        go io.Copy(server, client)
	} else {
        server.Write(b[:n])
	}

    //send the content back to client
	io.Copy(client, server)
}
