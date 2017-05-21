package main

import (
	"fmt"
	"libpeerproxy"
	"time"
	"log"
	"os"
	//"net"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run peerproxy.go [peer host]")
		return
	}

	p := libpeerproxy.NewProxyServer()
	host, err := libpeerproxy.ExternalIP()
	if err != nil {
		fmt.Println("ExternalIP :", err)
	}
	fmt.Println(host)
	host = os.Args[1]
	port := "7890"
	proxyPort := "3128"
	addr := host + ":" + port
	proxyAddr := host + ":" + proxyPort
	contact := libpeerproxy.Contact{host, port, proxyPort, addr, proxyAddr, -1}
	go p.DoPing(contact)

	//TODO: need to periodically update ContactList
	updateCh := make(chan bool)
	go func() {
		for {
			time.Sleep(5 * time.Second) // set 5 seconds for testing
			updateCh <- true
		}
	}()

	go func() {
		for {
			update := <- updateCh
			if update == true {
				// log.Println("===========update============")
				err := p.DoUpdateContactList()
				if err != nil {
					log.Println("Error DoUpdateContactList: ", err.Error())
				}
				// log.Println("===========update============")
			}
		}
	}()

	// Every ProxyServer serves as a proxy at addr proxyServerAddr
	go p.ServeAsProxy()

	// Every ProxyServer peer also serve as a proxyRouter,
	// only for routing requests of itself
	p.ServerAsProxyRouter()
}
