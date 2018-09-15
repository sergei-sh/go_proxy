/* Use baidu.com for testing, because it supports HTTP among few now.
Search "pictures":
http://www.baidu.com/s?ie=utf-8&f=8&rsv_bp=1&rsv_idx=1&tn=baidu&wd=pictures&oq=pictures&rsv_pq=e2e8bdf50001d014&rsv_t=19633Qxvy3eAnbL1isnp5lejNCAdPRcF3xp%2FP2NV1P92tfb%2FVMC8%2FsClWEQ&rqlang=cn&rsv_enter=1&rsv_sug3=1&rsv_sug2=0&inputT=5&rsv_sug4=2626&rsv_jmp=slow
*/

package main

import (
    "net/http"
    "fmt"
    "runtime"
    "os"
    "os/signal"
    "syscall"
    "sync"
    )

/* STARTUP PARAMETERS */

//Don't query pictures. Send from the cache only.
const CacheOnly bool = false

const proxyPort int = 8066

const dbUser string = "serj"
const dbPassword string = "qwerty"
const dbHost string = "localhost"
const dbPort int = 5432
/* END */

var routineId int
var rl = NewLogger(0) // Root logger

/*Need to override default ServeMux. Otherwise won't handle CONNECT because it doesn't receive valid path
see https://echorand.me/dissecting-golangs-handlerfunc-handle-and-defaultservemux.html */
type MyHandler struct{
    dbChannel chan DbJob
    exitChannel chan bool
    exitWg sync.WaitGroup
}

func (h MyHandler) ServeHTTP(responseW http.ResponseWriter, request *http.Request){
    routineId++
    routine := NewLoggerS(routineId,"J") 

    select {
        case <- h.exitChannel: 
            rl.Log("ServeHTTP exiting")
            //This will send an empty response with code 200
            return
        default:
            h.exitWg.Add(1)            
            switch request.Method {
                case http.MethodConnect:
                    HandleConnect(responseW, request, routine)
                default:
                    var dbJob *DbJob
                    HandleHttp(&dbJob, responseW, request, routine)

                    if nil != dbJob {
                        routine.Log("sending ", request.URL)
                        h.dbChannel <- *dbJob
                    }
            }
            h.exitWg.Done()
            routine.Log("done")
    }
}

func main() {
    /* Sets the number of goroutines actually running concurrently */
    curMP := runtime.GOMAXPROCS(6)
    rl.Log("Current max proc:", curMP)
    LogInit()

    handler := MyHandler{
        dbChannel: make(chan DbJob), 
        exitChannel: make(chan bool),
        exitWg: sync.WaitGroup{},
    }

    go DbWorker(&handler, dbUser, dbPassword, dbHost, dbPort)

    server := http.Server{
        Addr: fmt.Sprintf(":%d", proxyPort),
        Handler: handler,
    }

    /*  By Ctrl-c:
    1. Stop creating additional DB jobs / cancel already received requests
    2. Wait running job finished
    3. Make server exit */
    sigC := make(chan os.Signal, 1)
    signal.Notify(sigC, os.Interrupt, syscall.SIGTERM)
    go func(){
            <-sigC
            close(handler.exitChannel)
            rl.Log("Waiting for requests/jobs to be processed")
            handler.exitWg.Wait()
            rl.Log("All done")
            server.Close()
    }()
    //TODO: enquire routines are not being waited as expected

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
    rl.Log(server.ListenAndServe().Error())
}

