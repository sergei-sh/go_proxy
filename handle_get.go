package main

import(
    "fmt"
    "io"
    "net/http"
    )

// Handle HTTP GET method
// Pass response from origin
// Maybe implement caching
func HandleHttp(responseW http.ResponseWriter, request *http.Request, l Logger) {
    l.Log("G", request.URL)

    //http.Error(responseW, "In development", http.StatusNotFound)

    origResponse, err := http.DefaultTransport.RoundTrip(request)
    if err != nil {
        http.Error(responseW, "Bad Gateway", http.StatusBadGateway)
        l.Log("GX", err.Error())
        return
    }

    defer origResponse.Body.Close()
    //defer responseW.Close()
    copyHeader(responseW.Header(), origResponse.Header)
    fmt.Println("G response ", origResponse.StatusCode)
    responseW.WriteHeader(origResponse.StatusCode)
    io.Copy(responseW, origResponse.Body)

    return
}

// Header is map[string][] string
func copyHeader(dest, source http.Header) {
    for name, values := range source {
        dest.Set(name, values[0])
        //fmt.Println(name, values)
    }
}
