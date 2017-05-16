package main

import (
	"fmt"
	"libpeerproxy"
	//"net"
)

func main() {
	p := libpeerproxy.NewProxyServer()
	host, err := libpeerproxy.ExternalIP()
	if err != nil {
		fmt.Println("ExternalIP :", err)
	}
	port := "7890"
	addr := host + ":" + port
	contact := libpeerproxy.Contact{host, port, addr, -1}
	p.DoPing(contact)

	//TODO: need to periodically update ContactList

	// Every ProxyServer serves as a proxy at addr proxyServerAddr
	go p.ServeAsProxy()

	// Every ProxyServer peer also serve as a proxyRouter,
	// only for routing requests of itself
	p.ServerAsProxyRouter()
}
