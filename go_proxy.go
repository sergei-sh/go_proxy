package main

import (
    "net/http"
    "fmt"
    "runtime"
    )

//STARTUP PARAMETERS
const port int = 8066
const workerThreads = 8

var routine_id int

/*Need to override default ServeMux. Otherwise won't handle CONNECT because it doesn't receive valid path
see https://echorand.me/dissecting-golangs-handlerfunc-handle-and-defaultservemux.html */
type MyHandler struct{}

func (h MyHandler) ServeHTTP(responseW http.ResponseWriter, request *http.Request){
    fmt.Println("starting ", request.URL)
    routine_id++
    var l Logger = Logger{routine_id} 
    switch request.Method {
        case http.MethodConnect:
            HandleConnect(responseW, request, l)
        default:
            HandleHttp(responseW, request, l)
    }
    fmt.Println("done ", request.URL)
}

func main() {
    runtime.GOMAXPROCS(3)
    LogInit()

    /* Using the application as an explicit HTTP/HTTPS proxy.
    This means the client is configured to use the proxy for both protocols.
    HTTPS requests are tunnelled through the proxy using CONNECT method.

    Another options are:
        1. MITM proxy with explicit HTTPS proxying. Client is set up to use the proxy.
            a) Handling HTTP. Same as above.
            b) Handling HTTPS.  CONNECT doesn't open a tunnel but rather TLS connection to the origin is made. The client
               then gets proxy generated certificated with the Common Name tampered to be as of origin server.
               The client traffic is decrypted and can be analyzed. The client need to trust the generated certficate.
        2. Transparent proxy. Client is not aware of the proxy. Using external mechanisms such as IP tables
           to redirect client requests to the proxy. 
            a) Handling HTTP. Same as above.
            b) Handling HTTPS. Need to somehow know the original request destination domain name
               (if not part of request). Need to generate the certificate simialar to 1-b. 
    - left for future implementation */               
    mh := MyHandler{}
    fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), mh))
}

