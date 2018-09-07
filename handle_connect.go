package main

import(
    "io"
    "net"
    "net/http"
    "sync"
    )

// Handle HTTP CONNECT method
// Set up a tunnel with the request addressee
func HandleConnect(responseW http.ResponseWriter, request *http.Request, l Logger) {
    l.Log("J00")
    hijacker, ok := responseW.(http.Hijacker)
    if !ok {
        http.Error(responseW, "Hijacking not supported", http.StatusInternalServerError)
        l.Log("JY", "Hijacking not supported")
        return
    }

    origConn, err := net.DialTimeout("tcp", request.Host, ConstTimeOut)
    if err != nil {
        http.Error(responseW, err.Error(), http.StatusBadGateway)
        l.Log("JRX", err)
        return
    }

    responseW.WriteHeader(http.StatusOK)

    clientConn, _, err := hijacker.Hijack()
    if err != nil {
        http.Error(responseW, err.Error(), http.StatusInternalServerError)
        l.Log("JY", err)
        return
    }


    l.Log("J0")

    //Now then, client is informed on OK tunnelling so can start transfer
    
    defer clientConn.Close()
    defer origConn.Close() 

    var wg sync.WaitGroup

    transfer := func(dst io.WriteCloser, src io.ReadCloser) {
        wg.Add(1)
        go func(dst io.WriteCloser, src io.ReadCloser) {
            io.Copy(dst, src)
            wg.Done()
        } (dst, src)
    }

    //?Test G working
    transfer(clientConn, origConn)
    transfer(origConn, clientConn)

    l.Log("J")

    wg.Wait()
}
