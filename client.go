package main

import (
    "fmt"
    "net/http"
    "net/url"
    "os"
    "log"
    "net/http/httputil"
)

func main() {
    if len(os.Args) != 2 {
        fmt.Println("Usage: ", os.Args[0], "http://host:port/page")
        os.Exit(1)
    }

    url, err := url.Parse(os.Args[1])
    checkError(err)

    client := &http.Client{}
    log.Println(url.String())
    request, err := http.NewRequest("GET", url.String(), nil)

    dump, _ := httputil.DumpRequest(request, false)
    fmt.Println(string(dump))

    // only accept UTF-8
    request.Header.Add("Accept-Charset", "UTF-8;q=1, ISO-8859-1;q=0")
    checkError(err)

    response, err := client.Do(request)
    if response.Status != "200 OK" {
        fmt.Println(response.Status)
        os.Exit(2)
    }

    var buf [512]byte
    reader := response.Body
    fmt.Println("got body")
    for {
        n, err := reader.Read(buf[0:])
        fmt.Print(string(buf[0:n]))
        if err != nil {
            fmt.Println(err)
            os.Exit(0)
        }
    }

    os.Exit(0)
}

func checkError(err error) {
    if err != nil {
        fmt.Println("Fatal error ", err.Error())
        os.Exit(1)
    }
}
