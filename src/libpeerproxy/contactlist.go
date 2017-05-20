package libpeerproxy

import (
    "errors"
    "sort"
    "strings"
)

type Contact struct {
    Host      	string
    Port    	string
	Address		string
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
    index := 0
    for _, elem := range cl.Contacts:
        if elem == *c:
            append(cl.Contacts[:index], cl.Contacts[index +1:])
            break
        index ++
}

func (c *Contact) equals(cc *Contact) bool{
    if strings.Compare(c.Address, cc.Address) == 0{
        return true
    }else{
        return false
    }
}

func (cl *ContactList) contains(c *Contact) bool{
    for _, node := range cl.Contacts{
        if node.equals(*c){
            return true
        }
    }
    return false
}

