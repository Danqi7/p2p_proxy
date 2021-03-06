package main

import (
    "fmt"
    "net"
    "rpc"
)

const (
    rpcAddr             = "localhost:7890"
    proxyServerAddr     = "localhost:3128"
    proxyRouterAddr     = "localhost:3129"
)

type Contact struct {
    Host      net.IP
    Port    int
}

type ProxyServer struct {
    SlefContact     Contact
    ContactList     []Contacts
}

type ProxyServerRPC struct {
	proxyServer *ProxyServer
}

func NewProxyServer() *ProxyServer {
    laddr = rpcAddr
	p := new(ProxyServer)
    p.ContactList = make([]Contact, 0)

    // Set up RPC server
	// NOTE: ProxyServerRPC is just a wrapper around ProxyServer. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&ProxyServerRPC{p})
	hostname, port, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}

	s.HandleHTTP(rpc.DefaultRPCPath+port,
		rpc.DefaultDebugPath+port)

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Run RPC server forever.
	go http.Serve(l, nil)
	// Add self contact
	hostname, port, _ = net.SplitHostPort(l.Addr().String())
	port_int, _ := strconv.Atoi(port)
	ipAddrStrings, err := net.LookupHost(hostname)
	var host net.IP
	for i := 0; i < len(ipAddrStrings); i++ {
		host = net.ParseIP(ipAddrStrings[i])
		if host.To4() != nil {
			break
		}
	}
	p.SelfContact = Contact{host, port_int}

    // Every ProxyServer serves as a proxy at addr proxyServerAddr
    go p.ServeAsProxy()

    // Every ProxyServer peer also serve as a proxyRouter,
    // only for routing requests of itself
    go p.ServerAsProxyRouter()

	return p
}

// peer serve as a proxy and hancdles incoming connection
func (p *ProxyServer) ServeAsProxy() {
	listener, err := net.Listen("tcp", proxyServerAddr)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Proxy listening at address: %v", proxyServerAddr)
	id := 0
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Panic(err)
		}
		id += 1
		log.Println("Proxy new client: ", id)
		go p.HandleClient(conn)
	}
}

// Every ProxyServer peer has a proxyRouter
// that will randomly decide to go from itself or go from other proxies
// the router listens on default proxyRouterAddr 3129
// For user configuration, the peer need to configure its network proxy to the proxyRouter
func (p *ProxyServer) ServerAsProxyRouter() {
	router, err := net.Listen("tcp", proxyRouterAddr)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Router listening at address: %v", proxyRouterAddr)
	id := 0
	for {
		conn, err := router.Accept()
		if err != nil {
			log.Panic(err)
		}
		id += 1
		log.Println("Router new client: ", id)
		go p.RouteClient(conn)
	}
}


func (p *ProxyServer) HandleClient(client net.Conn) {
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

func (p *ProxyServer) RouteClient(client net.Conn) {
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


// ========= RPCs ==========//
func (p *ProxyServerRPC) Ping(ping *string, pong *string) {
    
}

// ========= RPCs ==========//
func (p *ProxyServer) PingPeers() {
	length := list.Len()
	for i := 0; i < length; i++ {
		peer := list[i]
		res := p.ping_peer(peer.address)

		if res != 0 {
			heap.Remove(&(p.list), peer.index)
		}
	}

	return
}

func (p *ProxyServer) DoPing(contact Contact) int {
    address := contact.IP.String() + ":" + strconv.Itoa(contact.Port)
    path := rpc.DefaultRPCPath + strconv.Itoa(contact.Port)

    client, err := rpc.DialHTTPPath("tcp", address, path)
    if err != nil {
        log.Fatal("Dialing: ", err, address)
    }

	ping := "ping"
	pong := ""

	err = client.Call("ProxyServerRPC.Ping", &ping, &pong)
	if err!= nil {
			log.Fatal("ProxyServerRPC.Ping", err)
			return 1
	}

	if pong == "Active"{
		return 0
	}

	return 1

}
