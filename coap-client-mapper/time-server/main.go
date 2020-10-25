/*
This app is used as a server and you after building the server you can run it.
Then the client-mapper will interact with this server if the ip addres and port
numbers are correctly set up.
*/

package main

import (
	"bytes"
	"fmt"
	"log"
	"time"

	coap "github.com/plgd-dev/go-coap/v2"
	"github.com/plgd-dev/go-coap/v2/message"
	"github.com/plgd-dev/go-coap/v2/message/codes"
	"github.com/plgd-dev/go-coap/v2/mux"
)

// State contains the state of the resource

func loggingMiddleware(next mux.Handler) mux.Handler {
	return mux.HandlerFunc(func(w mux.ResponseWriter, r *mux.Message) {
		log.Printf("ClientAddress %v, %v\n", w.Client().RemoteAddr(), r.String())
		next.ServeCOAP(w, r)
	})
}

// written in an elaborated way to show how the header organization is setup
func replyState(w mux.ResponseWriter, r *mux.Message) {
	// Generate a random temperature reading between 14 and 38
	byteArray := fmt.Sprint(time.Now().UnixNano())
	log.Println("Time sent", byteArray)
	err := w.SetResponse(codes.Content, message.TextPlain, bytes.NewReader([]byte(byteArray)))
	if err != nil {
		log.Printf("cannot set response: %v", err)
	}
}

func common(w mux.ResponseWriter, r *mux.Message) {
	// check if the request method is GET
	if r.Message.Code == codes.GET {
		replyState(w, r)
	} else {
		msg := fmt.Sprintf("Wrong path and method combination.\n")
		err := w.SetResponse(codes.MethodNotAllowed, message.TextPlain, bytes.NewReader([]byte(msg)))
		if err != nil {
			log.Println("cant send the error response")
		}
	}
}

func main() {
	log.Println("Started accepting coap requests: ")
	r := mux.NewRouter()
	r.Use(loggingMiddleware)
	r.Handle("/time", mux.HandlerFunc(common))
	coap.ListenAndServe("udp", ":5683", r)
}
