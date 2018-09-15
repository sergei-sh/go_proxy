package main

import (
    "database/sql"
    _ "github.com/lib/pq"
    "fmt"
)

type DbJob struct{
    url string
    data *[]byte
}

const dbName string = "go_proxy"

var l Logger = NewLogger(-1) // DB Logger
var db *sql.DB

/* Open a DB connection and wait for incoming SAVE requests from the channel.
When an HTTP response with a picture is returned from origin, a SAVE job is added to the channel.
READ request are not pipelined, but just handled synchronously in DbGet.
The difference is because ServeHTTP should not return before a response is written (if not hijacked,
which would leadd too much into low-level). In this setting it makes no sense to pipeline picture READs. */
func DbWorker(serverH *MyHandler, user string, password string, host string, port int) {
    connStr := fmt.Sprintf("user=%s password=%s host=%s port=%d dbname=%s", user, password, host, port, dbName)
    var err error
    db, err = sql.Open("postgres", connStr)
    if err != nil {
        l.Log("DBX", err)
    }
    db = db
    l.Log("DB", "connected")

    defer func() {
        db.Close()
        l.Log("DB", "worker is exiting")
    }()

    // Sender will close the channel signalling stop
    for job := range serverH.dbChannel{
        select {
            case <- serverH.exitChannel:
                //Stop handling further writes. Generally deferred Close() is enough by itself,
                //but will wait for the already queued jobs to finish in this case

                //Cannot justs close dbChannel, because need to do this from SIGINT handler
                //Should not close channel from other routine than that sending to it
                return
            default:
                serverH.exitWg.Add(1)
                handleJob(job)
                serverH.exitWg.Done()
        }
        //TODO: enquire inserts are running seemingly concurrently while should be in a single thread
    }

    // 1. add done if ? DB full
}

func handleJob(job DbJob) {
    l.Log("DQ", "inserting ", job.url)

    _, err := db.Exec("insert into cached_files(url, data) values($1,$2)", job.url, job.data)
    if err != nil {
        l.Log("DQX", err)
        return
    }
    l.Log("DQ", "done")
}

func DbGet(url string) *[]byte {
    l.Log("DQ", "getting ")

    var data []byte
    err := db.QueryRow("select data from cached_files where url=$1", url).Scan(&data)
    if err != nil { //Not found
        l.Log("DQ", "not found")
        return nil
    }
    l.Log("DQ", "found ")
    return &data

}
