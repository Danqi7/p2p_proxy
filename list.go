package main

import (
	"container/heap"
	// "strings"
	"net/rpc"
	//"fmt"
	//"strconv"
	"log"
	"time"
)

var K int = 3
var NumOfContacts int = 5


type proxy_server struct{
	ip string
	port int
	list PriorityQueue
}

func NewProxyServer(ipaddr string, portnum int) *proxy_server {
	p := new(proxy_server)
	p.ip = ipaddr
	p.port = portnum
	p.list = make(PriorityQueue, NumOfContacts)
	
	return p
}

func (p *proxy_server) ping_peers(list PriorityQueue){
	length := list.Len()
	for i:= 0; i< length; i++{
		peer := list[i]
		res := p.ping_peer(peer.address)

		if res != 0{
			heap.Remove(&(p.list), peer.index)
		}
	}
	return

}

func (p *proxy_server) ping_peer(peerstr string) int{
	client, err := rpc.DialHTTP("tcp", peerstr)

	if err != nil{
		log.Fatal("DialHTTP", err)
		return 1
	}

	ping := "ping"
	pong := ""

	err = client.Call("proxy_server.ping", &ping, &pong)
	if err!= nil{
			log.Fatal("proxy_server ping:", err)
			return 1
	}

	if pong == "Active"{
		return 0
	} 

	return 1

}

func (p *proxy_server) ping (ping *string, pong *string){
	*pong = "Active"
	return
}




func (p *proxy_server) AskContacts(list PriorityQueue){
	length := list.Len()
	for i:= 0; i< length; i++{
		peer := list[i]
		p.Ask(peer)
	}

	return
}


func (p *proxy_server) Ask(peer *contact){
	peerstr := peer.address
	client, err := rpc.DialHTTP("tcp", peerstr)

	if err != nil{
		log.Fatal("DialHTTP", err)
		heap.Remove(&(p.list), peer.index)
		return
	}

	request := K
	response := make([]string, K)

	err = client.Call("proxy_server.GetContacts", &request, &response)

	if err!= nil{
		log.Fatal("proxy_server GetContacts:", err)
		return
	} 

	for _, contact := range response{
		p.update(contact)
	} 

}

func (p *proxy_server) update(peerstr string){
	start := time.Now()
	result := p.ping_peer(peerstr)
	if result != 0{
		return
	}
	elapse := time.Since(start)
	duration := int64(elapse/time.Millisecond)
	new_peer := &contact{address:peerstr, latency: duration}
	heap.Push(&(p.list), new_peer)
	
}