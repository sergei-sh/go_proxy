package main

import (
    "net/http"
    "fmt"
    "runtime"
    "time"
    )

//STARTUP PARAMETERS
const port int = 8066
const workerThreads = 3
const queueTimeOutMsec = 1000    

var routine_id int
var rl Logger = Logger{0} // Root logger

type proxyJob struct{
    responseWriter http.ResponseWriter
    request* http.Request
}

/*Need to override default ServeMux. Otherwise won't handle CONNECT because it doesn't receive valid path
see https://echorand.me/dissecting-golangs-handlerfunc-handle-and-defaultservemux.html */
type MyHandler struct{
    jobs chan proxyJob
}

func (h MyHandler) ServeHTTP(responseW http.ResponseWriter, request *http.Request){
    select {
        case h.jobs <- proxyJob{responseW, request}:
        // Let requests queue for small amount of time. If no workers available after waiting, drop a request
        case <- time.After(time.Duration(queueTimeOutMsec) * time.Millisecond):
            rl.Message("dropping ", *request.URL)
            
    }

}

/* Fan-in */
func (h MyHandler) workerSinkRoutine() {
    routine_id++
    var l Logger = Logger{routine_id} 

    for {
        select {
            case job := <-h.jobs:
                rl.Message("starting ", job.request.URL)
                rl.Log("starting ", job.request.URL, "abc", "cd")
                switch job.request.Method {
                    case http.MethodConnect:
                        HandleConnect(job.responseWriter, job.request, l)
                    default:
                        HandleHttp(job.responseWriter, job.request, l)
                }
                fmt.Println("done ", job.request.URL)
           //case: add stop condition if needed
        }
    }
}

func main() {
    /* Sets the number of goroutines actually running concurrently */
    runtime.GOMAXPROCS(3)
    LogInit()

    mh := MyHandler{make(chan proxyJob)}
    /* I would like to have a fixed number of routines as workers handling requests
    This achieves 2 goals (comapred to spawning a new routine for each request):
    -no overhead on spawning;
    -no resource exhaustion in case of some workers are blocked infinitely; */
    for i := 0; i < workerThreads; i++ {
        go mh.workerSinkRoutine()
    }

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
    fmt.Println(http.ListenAndServe(fmt.Sprintf(":%d", port), mh))
}

