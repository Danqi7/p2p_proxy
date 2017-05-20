package libpeerproxy

import (
    //"net"
    "errors"
)

type ProxyServerRPC struct {
	proxyServer *ProxyServer
}

///////////////////////////////////////////////////////////////////////////////
// PING
///////////////////////////////////////////////////////////////////////////////
type PingMessage struct {
	Sender Contact
    Msg    string
}

type PongMessage struct {
	Sender Contact
    Msg    string
}
func (p *ProxyServerRPC) Ping(ping PingMessage, pong *PongMessage) error {
    if ping.Msg != "ping" {
        return errors.New("Wrong ping msg")
    }


    pong.Sender = p.proxyServer.SelfContact
    pong.Msg = "pong"

    // update sender in ContactList
    p.proxyServer.ContactList.UpdateContactWithoutLatency(&ping.Sender)

    return nil
}

///////////////////////////////////////////////////////////////////////////////
// ASK  FOR CONTACTS
///////////////////////////////////////////////////////////////////////////////
type AskContactsRequest struct {
	Sender Contact
    Number int
}

type AskContactsResult struct {
	Sender Contact
    Nodes  []Contact
}
//TODO:

func (p *ProxyServerRPC) AskForContacts(request AskContactsRequest, reply *AskContactsRequest) error {
    var length int
    if len(request.Number) < len(p.ContactList){
        length = len(request.Number)
    }else{
        length = len(p.ContactList)
    }
    for i=0; i < length; i++{
        contact := *(p.ContactList[i])
        reply.Nodes = append(reply.Nodes, contact)
    }

    return nil
}
