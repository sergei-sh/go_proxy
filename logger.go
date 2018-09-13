package main

import(
    "fmt"
    "log"
    s "strings"
    )

//LOGGING PARAMETERS
const logOptions int = ^log.Ldate | log.Ltime | log.Lmicroseconds

type Logger interface {
    Log(msgs ...interface{})
}

type logger struct {
    routineId int
    suppressCodes string
}

func NewLogger(routineId int) *logger {
    l := new(logger)
    l.routineId = routineId
    return l
}

func NewLoggerS(routineId int, suppressCodes string) *logger {
    l := new(logger)
    l.routineId = routineId
    l.suppressCodes = suppressCodes
    return l
}

func LogInit() {
    log.SetFlags(logOptions)
}

/* Log with routine id and short event code */
func (l logger) Log(msgs ...interface{}) {
        /* The first label is the message code (should be string). Codes starting from the same
        letter relate to the same control path describing some high high-level event 
        (like HTTPS tunnelling). It is also handy to be able to filter by some code's
        first letter */

        sMsg := msgs[0].(string)
        if len(sMsg) != 0 && s.Contains(l.suppressCodes, sMsg[0:1]) {
            return
        }

        outMsg := ""
        for _, msgIf := range msgs {
            msgStr := fmt.Sprintf("%s", msgIf)
            outMsg = s.Join([]string{outMsg, msgStr}, " ")
        }
        log.Printf(" %d: %s ", l.routineId, outMsg)
}



