package commons

import (
	"fmt"
	"log"
)

func getLogger(name string) func(format string, args ...interface{}) {
	return func(format string, args ...interface{}) {
		message := format
		if len(args) > 0 {
			message = fmt.Sprintf(format, args)
		}
		log.Println("["+name+"]", message)
	}
}
