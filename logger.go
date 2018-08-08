package main

import(
    "log"
    )

//LOGGING PARAMETERS
const logOptions int = ^log.Ldate | log.Ltime | log.Lmicroseconds

type Logger struct {
    routine_id int
}

func LogInit() {
    log.SetFlags(logOptions)
}

func (l Logger) Log(short string, msg ...interface{}) {
    if 0 == len(msg) {
        log.Printf(" %d-%s ", l.routine_id, short)
    } else {
        msg = append([]interface{}{l.routine_id, short}, msg...)
        log.Printf(" %d-%s: %s ", msg...)
    }
}



