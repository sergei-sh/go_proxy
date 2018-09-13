package main

import(
    "fmt"
    "io"
    "io/ioutil"
    "net/http"
    s "strings"
    )

func min(left int, right int) int {
    if left < right {
        return left
    } else {
        return right
    }
}

// Handle HTTP GET method
// Pass response from origin
// dbJob will be set if there is a file to be saved to cache
func HandleHttp(dbJob **DbJob, responseW http.ResponseWriter, request *http.Request, l Logger) {

    url := request.URL.String()

    const dispWrap = 120
    l.Log("G", url[:min(dispWrap, len(url))])

    if isPicture(url) {
        l.Log("")
        pic := DbGet(url)
        if pic == nil {
            if CacheOnly {
                http.Error(responseW, "", http.StatusNotFound)
                return
            }
            //Else - continue handling the request
        } else {
            l.Log("GC")
            // Its not clear from the ResponseWriter source comments whether this
            // can be omitted
            responseW.Header().Add("Content-Lentgth", fmt.Sprintf("%s", len(*pic)))
            _, err := responseW.Write(*pic) // will write headers as well 
            if err != nil {
                l.Log("GCX")
            }
            return
        }
    }

    origResponse, err := http.DefaultTransport.RoundTrip(request)
    if err != nil {
        http.Error(responseW, "Bad Gateway", http.StatusBadGateway)
        l.Log("GX", err.Error())
        return 
    }

    defer origResponse.Body.Close()
    copyHeader(responseW.Header(), origResponse.Header)
    responseW.WriteHeader(origResponse.StatusCode)

    respBodyT := io.TeeReader(origResponse.Body, responseW)

    l.Log("G", "responding")
    byPic, err := ioutil.ReadAll(respBodyT)
    if err != nil {
        l.Log("GX", err.Error())
        return 
    }

    if isPicture(url) {
        l.Log("GCS", "saving",  url)
        *dbJob = &DbJob{url, &byPic}
    }
}

// Header is map[string][] string
func copyHeader(dest, source http.Header) {
    for name, values := range source {
        dest.Set(name, values[0])
        //fmt.Println(name, values)
    }
}

func isPicture(url string) bool {
    return s.HasSuffix(url, ".jpg")  || s.HasSuffix(url, ".png")
}
