
package main

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
)

type Pxy struct {}

func (p *Pxy) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	fmt.Printf("Received request %s %s %s\n", req.Method, req.Host, req.RemoteAddr)

	transport :=  http.DefaultTransport

	// step 1
	outReq := new(http.Request)
	*outReq = *req // this only does shallow copies of maps

	if clientIP, _, err := net.SplitHostPort(req.RemoteAddr); err == nil {
		if prior, ok := outReq.Header["X-Forwarded-For"]; ok {
			clientIP = strings.Join(prior, ", ") + ", " + clientIP
		}
		outReq.Header.Set("X-Forwarded-For", clientIP)
	}

	// step 2
	res, err := transport.RoundTrip(outReq)
	if err != nil {
		rw.WriteHeader(http.StatusBadGateway)
		return
	}

	// step 3
	for key, value := range res.Header {
		for _, v := range value {
			rw.Header().Add(key, v)
		}
	}

	rw.WriteHeader(res.StatusCode)
	io.Copy(rw, res.Body)
	res.Body.Close()
}

func main() {
	fmt.Println("Serve on :8080")
	http.Handle("/", &Pxy{})
	http.ListenAndServe("0.0.0.0:8080", nil)
}




// package main

// import(
// 	//"io"
// 	"net/http"
// 	"log"
// )

// var(
// 	listenPort = "8080"
// 	listenIP = "localhost"
// )

// func main() {
// 	proxyserver := NewProxyServer(listenIP, listenPort)
// 	log.Fatal(proxyserver.ListenAndServe())


// 	// &http.Server{
// 	// 	Addr: ":8000",
// 	// 	Handler: &myHandler,
// 	// }
	
// 	// log.Fatal(proxyserver.ListenAndServe())
// 	// log.Println("proxy server launched")
// }


// func NewProxyServer(listenIP string, listenPort string) *http.Server {
// 	listen := listenIP + ":" + listenPort
// 	return &http.Server{
// 		Addr:           listen,
// 		Handler:        &myHandler{},
// 		// ReadTimeout:    10 * time.Second,
// 		// WriteTimeout:   10 * time.Second,
// 		// MaxHeaderBytes: 1 << 20,
// 	}
// }

// type myHandler struct{
// 	//transport http.Transport
// }

// func (h *myHandler) ServeHTTP(w http.ResponseWriter, r *http.Request){
// 	if r.Method == "CONNECT"{
// 		h.HttpsHandler(w,r)
// 	}else{
// 		h.HttpHandler(w,r)
// 	}
// }

// func(h *myHandler) HttpsHandler(w http.ResponseWriter, r *http.Request){
// 	log.Println("using HTTPs Handler")
// 	log.Println(r.Method, r.URL.Host)
// 	//client_host := r.RemoteAddr.String()

// }

// func(h *myHandler) HttpHandler(w http.ResponseWriter, r *http.Request){
// 	log.Println("using HTTP Handler")
// 	log.Println(r.Method, r.URL.Host)
// }
