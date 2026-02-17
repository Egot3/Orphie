package handlers

import "log"

func AMQPErrorHandler(err error, msg string) {
	if err != nil {
		log.Panicf("%s: %s", msg, err)
	}
}
