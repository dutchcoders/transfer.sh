package ghost

import (
	"log"
)

// Logging function, defaults to Go's native log.Printf function. The idea to use
// this instead of a *log.Logger struct is that it can be set to any of log.{Printf,Fatalf, Panicf},
// but also to more flexible userland loggers like SeeLog (https://github.com/cihub/seelog).
// It could be set, for example, to SeeLog's Debugf function. Any function with the
// signature func(fmt string, params ...interface{}).
var LogFn = log.Printf
