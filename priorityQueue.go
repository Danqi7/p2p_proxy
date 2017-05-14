package main
import (
	"container/heap"
	 "fmt"
	//"strconv"
)

var capacity int = 3


func main() {
	// Some items and their priorities.
	items := map[string]int64{
		"peer1": 2, "peer2": 5, "pee3": 4,
	}

	// Create a priority queue, put the items in it, and
	// establish the priority queue (heap) invariants.
	pq := make(PriorityQueue, capacity)
	i := 0
	for addr, lat := range items {
		pq[i] = &contact{
			address:    addr,
			latency: lat,
			index:    i,
		}
		i++
	}
	heap.Init(&pq)

	//Insert a new item and then modify its priority.
	item := &contact{
		address:    "peer 4",
		latency: 1,
	}
	heap.Push(&pq, item)
	//pq.update(item, item.value, 5)

	// Take the items out; they arrive in decreasing priority order.
	for pq.Len() > 0 {
		item := heap.Pop(&pq).(*contact)
		fmt.Printf("%d%f", item.address, item.latency)
	}
}




type contact struct{
	address string
	latency int64
	index int
}

type PriorityQueue[] *contact

func(pq PriorityQueue) Len() int {return len(pq)}

func(pq PriorityQueue) Less(i,j int) bool{
	return pq[i].latency > pq[j].latency
}

func(pq PriorityQueue) Swap(i,j int){
	pq[i], pq[j] = pq[j], pq[i]
    pq[i].index = i
    pq[j].index = j
}

func (pq *PriorityQueue) Push(x interface{}) {
	length := pq.Len()
	item := x.(*contact)
	if length == capacity{
		//fmt.Printf("length is %d", length)
		// max := (*pq)[0]
		//fmt.Printf("max has latency %f", max.latency)
		if item.latency < (*pq)[0].latency{
			_ = heap.Pop(pq).(*contact)
			new_len := len(*pq)
    		item.index = new_len
    		*pq = append(*pq, item)
		}
	}else{
		item.index = length
		*pq = append(*pq, item)
	}
        
}

func (pq *PriorityQueue) Pop() interface{} {
    old := *pq
    n := len(old)
    item := old[n-1]
    item.index = -1 // for safety
    *pq = old[0 : n-1]
    return item
}





// func(pq *PriorityQueue) insert


