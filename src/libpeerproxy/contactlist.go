package libpeerproxy

import (
    "errors"
    "sort"
    "strings"
    "log"
)

type Contact struct {
    Host      	string
    RPCPort    	string
    ProxyPort   string
	RPCAddr 	string
    ProxyAddr   string
    Latency     int64
}

type ContactList struct {
    Contacts    []Contact
    Capacity    int
    sem         chan int // act as a semaphore/lock
}

func (cl *ContactList) Init(k int) {
    cl.Contacts = make([]Contact, 0, k)
    cl.Capacity = k
    cl.sem = make(chan int, 1)
}
//
func (cl *ContactList) UpdateContactWithoutLatency(c *Contact) error {
    return cl.UpdateContactWithLatency(c, DefaultLatency)
}

// Update contact with given latency
func (cl *ContactList) UpdateContactWithLatency(c *Contact, latency int64) error {
    cl.sem <- 1
    log.Println("UpdateContactWithLatency: ", c)
    // if contact already in ContactList, remove it and then re-insert
    if cl.Contains(c) {
        log.Println("Contains!!!UpdateContactWithLatency: ", c)
        <- cl.sem
        cl.RemoveContact(c)
        cl.UpdateContactWithLatency(c, latency)
        return nil
    }

    if len(cl.Contacts) >= cl.Capacity {
        <- cl.sem
        return errors.New("ContactList is full, disgarding the update request")
    }

    c.Latency = latency
    cl.Contacts = append(cl.Contacts, *c)

    // sort input contacts by latency
    // the closest contact at first
    sort.Slice(cl.Contacts, func(i, j int) bool {
        c1 := cl.Contacts[i]
        c2 := cl.Contacts[j]

        // -1 means latency is unknown, unknown is larger than known
        if c1.Latency == DefaultLatency {
            return false
        }
        if c2.Latency == DefaultLatency {
            return true
        }

        return c1.Latency < c2.Latency
    })

    <- cl.sem
    return nil
}

// TODO: Remove contact
func (cl *ContactList) RemoveContact(c *Contact) error {
    cl.sem <- 1
    found := 0

    for index, con := range cl.Contacts {
        if con.equals(c) {
            cl.Contacts = append(cl.Contacts[:index], cl.Contacts[index+1:]...)
            found = 1
            break
        }
    }

    if found == 0 {
        <- cl.sem
        return errors.New("Trying to remove a non-existen contact in ContactList!")
    }

    log.Println("RemoveContact: ", cl.Contacts)
    <-cl.sem
    return nil
}

func (c *Contact) equals(cc *Contact) bool{
    if strings.Compare(c.RPCAddr, cc.RPCAddr) == 0 {
        return true
    } else {
        return false
    }
}

func (cl *ContactList) Contains(c *Contact) bool{
    for _, node := range cl.Contacts {
        if node.equals(c){
            return true
        }
    }
    return false
}
