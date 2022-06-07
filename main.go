package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/chfgma/kropotnik/calls"
)

func main() {
	http.HandleFunc("/call", calls.CallHandler)
	http.HandleFunc("/record", calls.RecordingHandler)
	http.HandleFunc("/transcription", calls.TranscriptionHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
