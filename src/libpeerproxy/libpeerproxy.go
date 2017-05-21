package libpeerproxy

import (
    "fmt"
    "net"
	"log"
	"net/http"
    "net/rpc"
	//"strconv"
	"bytes"
	"io"
	"math/rand"
	"net/url"
	"strings"
	"errors"
	"time"
)

const (
    RPCAddr             = ":7890"
    ProxyServerAddr     = ":3128"
    ProxyRouterAddr     = ":3129"
	k					= 5
	DefaultLatency		= -1
)

// a list of avaiable proxies, need to decide how to get
// the global list with p2p later
var PROXY_LIST = [2]string{"self", "10.105.99.145:3128"}

type ProxyServer struct {
    SelfContact     Contact
    ContactList     *ContactList
}

func NewProxyServer() *ProxyServer {
    laddr := RPCAddr
	p := new(ProxyServer)
    p.ContactList = new(ContactList)
	p.ContactList.Init(k)

    // Set up RPC server
	// NOTE: ProxyServerRPC is just a wrapper around ProxyServer. This type includes
	// the RPC functions.

	s := rpc.NewServer()
	s.Register(&ProxyServerRPC{p})
	_, rpcPort, err := net.SplitHostPort(laddr)
	if err != nil {
		return nil
	}

	s.HandleHTTP(rpc.DefaultRPCPath+rpcPort,
		rpc.DefaultDebugPath+rpcPort)

	l, err := net.Listen("tcp", laddr)
	if err != nil {
		log.Fatal("Listen: ", err)
	}

	// Run RPC server forever.
	go http.Serve(l, nil)
	// Add self contact
	host, err := ExternalIP()
	if err != nil {
		log.Println(err)
	}

	// port := strconv.Atoi(ProxyServerAddr[1:])
	address := host + ProxyServerAddr
    rpcAddress := host + RPCAddr
	p.SelfContact = Contact{host, rpcPort, "3128", rpcAddress, address, DefaultLatency}


	// Every ProxyServer serves as a proxy at addr proxyServerAddr
	// go p.ServeAsProxy()
	//
	// // Every ProxyServer peer also serve as a proxyRouter,
	// // only for routing requests of itself
	// go p.ServerAsProxyRouter()

	return p
}

// peer serve as a proxy and hancdles incoming connection
func (p *ProxyServer) ServeAsProxy() {
	listener, err := net.Listen("tcp", ProxyServerAddr)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Proxy listening at address: %v", ProxyServerAddr)
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
	router, err := net.Listen("tcp", ProxyRouterAddr)
	if err != nil {
		log.Panic(err)
	}

	log.Printf("Router listening at address: %v", ProxyRouterAddr)
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
    // TODO: read the length of ContactList need to be atomic
	proxyIndex := rand.Intn(len(p.ContactList.Contacts))
    proxyIndex -= 1
    p.PrintContactList()
    // proxyIndex := 0
	// index is 0, go from peer itself
	// log.Println("~~~~~~~index: ", proxyIndex)
	if proxyIndex == -1 {
		p.ForwardFromItself(client, method, address, b)
	} else {
		// otherwise go from the real proxy
		p.ForwardFromProxy(p.ContactList.Contacts[proxyIndex].ProxyAddr, client, method, address, b)
	}
}
func (p *ProxyServer) ForwardFromProxy(proxyString string, client net.Conn, method string, address string, b [1024]byte) {
	//dial the proxy
	log.Println("forwardFromProxy proxy: ", proxyString)
	server, err := net.Dial("tcp", proxyString)
	if err != nil {
		log.Println("forwardFromProxy Error: ", err)
        // TODO: remove Error producing peer
        // p.RemoveContact
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

func (p *ProxyServer) ForwardFromItself(client net.Conn, method string, address string, b [1024]byte) {
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

// TODO: Ask every contact in ContactList to return its own ContactList
//Remove non-responding contact
func (p *ProxyServer) DoUpdateContactList() error {
    shortlist := make([]Contact, 0)
	for _, contact := range p.ContactList.Contacts {
		nodes, err := p.AskForContacts(contact, k)

        if err != nil {
            // remove this contact
            p.ContactList.RemoveContact(&contact)
            continue
        }

        for _, node := range nodes {
            shortlist = append(shortlist, node)
        }
	}

    index := 0
    for len(p.ContactList.Contacts) < p.ContactList.Capacity && index < len(shortlist){
        // add node from shortlist
        c := shortlist[index]
        p.ContactList.UpdateContactWithoutLatency(&c)

        index += 1
    }

	return nil
}

// TODO: Ask n number of contacts from input contact
func (p *ProxyServer) AskForContacts(c Contact, n int) ([]Contact, error) {
	address := c.RPCAddr
    path := rpc.DefaultRPCPath + c.RPCPort

    client, err := rpc.DialHTTPPath("tcp", address, path)
    if err != nil {
        log.Fatal("Dialing: ", err, address)
    }

    start := time.Now()
	request := new(AskContactsRequest)
	request.Sender = p.SelfContact
	request.Number = n
	reply := new(AskContactsResult)
	err = client.Call("ProxyServerRPC.AskForContacts", request, &reply)
	if err != nil {
			return nil, errors.New("Failed to call RPC AskForContacts on address: " + address)
	}

    // update contact
	elapse := time.Since(start)
	duration := int64(elapse/time.Millisecond)
	update := reply.Sender

	p.ContactList.UpdateContactWithLatency(&update, duration)

    // log.Println("AskForContacts:", reply.Nodes)
	return reply.Nodes, nil
}

func (p *ProxyServer) DoPing(contact Contact) error {
    address := contact.RPCAddr
    path := rpc.DefaultRPCPath + contact.RPCPort

    client, err := rpc.DialHTTPPath("tcp", address, path)
    if err != nil {
        log.Fatal("Dialing: ", err, address)
    }

	pingMsg := new(PingMessage)
	pingMsg.Sender = p.SelfContact
	pingMsg.Msg = "ping"

	//Dial RPC and compute the latency
	var pongMsg PongMessage
	start := time.Now()

	err = client.Call("ProxyServerRPC.Ping", pingMsg, &pongMsg)
	if err != nil {
			return errors.New("Failed to dial address: " + address)
	}

	if pongMsg.Msg != "pong" {
		return errors.New("Wrong pong message: " + pongMsg.Msg)
	}

	// update contact
	elapse := time.Since(start)
	duration := int64(elapse/time.Millisecond)
	update := pongMsg.Sender
	p.ContactList.UpdateContactWithLatency(&update, duration)

	return nil
}

func (p *ProxyServer) PrintContactList() {
    log.Println("=========Current ContactList=========")
    for _, con := range p.ContactList.Contacts {
        log.Println(con.ProxyAddr)
    }
    log.Println("=========Current ContactList=========")
}
