package libpeerproxy

import (
    //"net"
    "log"
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

    log.Println("!!!!!!=================GET A PING!=================!!!!!!")
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

func (p *ProxyServerRPC) AskForContacts(request AskContactsRequest, reply *AskContactsResult) error {
    var length int
    if request.Number < len(p.proxyServer.ContactList.Contacts){
        length = request.Number
    } else {
        length = len(p.proxyServer.ContactList.Contacts)
    }
    for i := 0; i < length; i++ {
        contact := p.proxyServer.ContactList.Contacts[i]
        reply.Nodes = append(reply.Nodes, contact)
    }

    reply.Sender = p.proxyServer.SelfContact

    // update sender in ContactList
    p.proxyServer.ContactList.UpdateContactWithoutLatency(&request.Sender)

    return nil
}
