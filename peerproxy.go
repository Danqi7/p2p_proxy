package main

import (
	"fmt"
	"libpeerproxy"
	"log"
	"os"
	"time"
	//"strconv"
)

// const UPDATETIME = 3600 // default per hour an update

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
	fmt.Println("My own IP Address: " + host)

	host = os.Args[1]
	port := "7890"
	proxyPort := "3128"
	addr := host + ":" + port
	proxyAddr := host + ":" + proxyPort
	contact := libpeerproxy.Contact{host, port, proxyPort, addr, proxyAddr, -1}
	go p.DoPing(contact)

	// periodically update ContactList
	updateCh := make(chan bool)
	// var updateTime time.Duration
	// if len(os.Args) > 2 {
	// 	updateTime, _ = strconv.Atoi(os.Args[2])
	// } else {
	// 	updateTime = UPDATETIME
	// }

	go func() {
		for {
			time.Sleep(60 * time.Second) // customized for testing purpose
			updateCh <- true
		}
	}()

	go func() {
		for {
			update := <-updateCh
			if update == true {
				p.ContactList.PrintContactList()

				err := p.DoUpdateContactList()
				if err != nil {
					log.Println("Error DoUpdateContactList: ", err.Error())
				}
			}
		}
	}()

	// Every ProxyServer serves as a proxy at addr proxyServerAddr
	go p.ServeAsProxy()

	// Every ProxyServer peer also serve as a proxyRouter,
	// only for routing requests of itself
	p.ServerAsProxyRouter()
}
