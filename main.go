package main

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/BTBurke/twiml"
	"github.com/avast/retry-go"
	"github.com/brianloveswords/airtable"
)

func main() {
	http.HandleFunc("/call", CallHandler)
	http.HandleFunc("/record", RecordingHandler)
	http.HandleFunc("/transcription", TranscriptionHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("listening on %s", port)
	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}

type Inbound struct {
	airtable.Record
	Fields InboundFields
}

type InboundFields struct {
	Number        string `json:"Phone number"`
	Recording     string `json:"Recording"`
	Transcription string `json:"Transcription"`
	TwilioID      string `json:"Twilio ID"`
}

func CallHandler(w http.ResponseWriter, r *http.Request) {
	var vr twiml.VoiceRequest
	if err := twiml.Bind(&vr, r); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	response := twiml.NewResponse()
	response.Add(&twiml.Say{
		Text: "Hello from Clinton Hill Fort Greene Mutual Aid. Please tell us why you are calling.",
	})

	response.Add(&twiml.Record{
		Action:             "/record",
		TranscribeCallback: "/transcription",
	})

	b, err := response.Encode()
	if err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	if _, err := w.Write(b); err != nil {
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", "application/xml")

	log.Printf("Incoming call from %s", vr.From)
}

func RecordingHandler(w http.ResponseWriter, r *http.Request) {
	var recording twiml.RecordActionRequest
	if err := twiml.Bind(&recording, r); err != nil {
		log.Printf("erroring receiving recording %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	airtableClient := airtable.Client{
		APIKey: os.Getenv("AIRTABLE_API_KEY"),
		BaseID: os.Getenv("AIRTABLE_BASE_ID"),
	}
	inbound := airtableClient.Table("inbound")

	if err := inbound.Create(&Inbound{
		Fields: InboundFields{
			TwilioID:  recording.CallSid,
			Number:    recording.From,
			Recording: recording.RecordingURL + ".mp3",
		},
	}); err != nil {
		log.Printf("erroring creating inbound %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}
}

func TranscriptionHandler(w http.ResponseWriter, r *http.Request) {
	var transcription twiml.TranscribeCallbackRequest
	if err := twiml.Bind(&transcription, r); err != nil {
		log.Printf("erroring receiving transcription %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	airtableClient := airtable.Client{
		APIKey: os.Getenv("AIRTABLE_API_KEY"),
		BaseID: os.Getenv("AIRTABLE_BASE_ID"),
	}
	inbound := airtableClient.Table("inbound")

	entry := Inbound{
		Fields: InboundFields{
			TwilioID:      transcription.CallSid,
			Number:        transcription.From,
			Recording:     transcription.RecordingURL + ".mp3",
			Transcription: transcription.TranscriptionText,
		},
	}

	var inbounds []Inbound

	if err := retry.Do(
		func() error {
			if err := inbound.List(&inbounds, &airtable.Options{
				Filter:     fmt.Sprintf(`{Twilio ID} = %q`, transcription.CallSid),
				MaxRecords: 1,
			}); err != nil {
				log.Println("retrying")
				return err
			}
			if len(inbounds) == 0 {
				log.Println("retrying")
				return errors.New("found no matching records")
			}
			return nil
		},
	); err != nil {
		log.Printf("error listing inbound records %v for SID %q", err, transcription.CallSid)
	}

	if len(inbounds) == 0 {
		log.Printf("could not preexisting entry for for SID %q", transcription.CallSid)

		if err := inbound.Create(&entry); err != nil {
			log.Printf("erroring creating inbound %v", err)
			http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

			return
		}

		return
	}

	entry.ID = inbounds[0].ID

	if err := inbound.Update(&entry); err != nil {
		log.Printf("erroring creating inbound %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}
}
