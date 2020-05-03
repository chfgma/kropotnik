package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chfgma/kropotnik/calls"
	"github.com/chfgma/kropotnik/slack"
)

func main() {
	http.HandleFunc("/call", calls.CallHandler)
	http.HandleFunc("/record", calls.RecordingHandler)
	http.HandleFunc("/transcription", calls.TranscriptionHandler)
	http.HandleFunc("/basics", slack.BasicsCommand)

	http.HandleFunc("/forward", calls.CallForward)
	http.HandleFunc("/forward_verify", calls.CallForwardVerify)
	http.HandleFunc("/forward_complete", calls.CallForwardComplete)
	http.HandleFunc("/forward_error", calls.CallForwardError)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
