package main

import (
    "fmt"
    "log"
    "net"
    "net/url"
    "bytes"
    "strings"
    "io"
    "math/rand"
    "time"
)

// a list of avaiable proxies, need to decide how to get
// the global list with p2p later
var PROXY_LIST = [2]string{"self", "192.168.1.102:3128"}

func main() {
    // every peer serves as an proxy on its port 3128
    go serveAsProxy()

    // for every peer, the router listens on 3300
    routerPort := ":3300"
    router, err := net.Listen("tcp", ":3300")
    if err != nil {
        log.Panic(err)
    }

    log.Printf("Router listening at port: %v", routerPort)
    id := 0
    for {
        conn, err := router.Accept()
        if err != nil {
            log.Panic(err)
        }
        id += 1
        log.Println("Router new client: ", id)
        go routeClient(conn)
    }
}

func routeClient(client net.Conn) {
    if client == nil {
        return
    }
    defer client.Close()

    // get destination host and address
    var b [1024]byte
    _, err := client.Read(b[:])
    if err != nil {
        log.Println(err)
        return
    }

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

    // decide whether to go from peer itself
    // or go from the real proxy
    t := time.Now()
    var seed int64 = int64(t.Second())
    rand.Seed(seed)
    proxyIndex := rand.Intn(len(PROXY_LIST))
    // index is 0, go from peer itself
    log.Println("index: ", proxyIndex)
    if proxyIndex == 0 {
        forwardFromItself(client, method, address, b)
    } else {
        // otherwise go from the real proxy
        forwardFromProxy(PROXY_LIST[proxyIndex], client, method, address, b)
    }
}

func forwardFromItself(client net.Conn, method string, address string, b [1024]byte) {
    if client == nil {
        return
    }
    log.Println("forwardFromItself: ", method, address)

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
        server.Write(b[:])
	}

    //send the content back to client
	io.Copy(client, server)
}

func forwardFromProxy(proxyString string, client net.Conn, method string, address string, b [1024]byte) {
    //dial the proxy
    log.Println("forwardFromProxy proxy: ", proxyString)
    server, err := net.Dial("tcp", proxyString)
    if err != nil {
        log.Println("forwardFromProxy Error: ", err)
        return
    }

    if method == "CONNECT" {
        server.Write(b[:])
        go io.Copy(server, client)
    } else {
        server.Write(b[:])
    }

    //send the content back to client
    io.Copy(client, server)
}


// peer serve as a proxy and hancdles incoming connection
func serveAsProxy() {
    port := ":3128"
    listener, err := net.Listen("tcp", ":3128")
    if err != nil {
        log.Panic(err)
    }

    log.Printf("Proxy listening at port: %v", port)
    id := 0
    for {
        conn, err := listener.Accept()
        if err != nil {
            log.Panic(err)
        }
        id += 1
        log.Println("Proxy new client: ", id)
        go handleClient(conn)
    }
}

func handleClient(client net.Conn) {
    if client == nil {
        return
    }
    defer client.Close()
    log.Println("handleClient: reading client's request")

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
