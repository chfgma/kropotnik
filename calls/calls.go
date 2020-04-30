package calls

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

const (
	inboundTable = "inbound"
	greeting     = "Thank you for calling the Myrtle Avenue Brooklyn Partnership hotline. After the tone please tell us why you are calling and someone will call you back soon. You can hang up when you are done."
)

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

/*
* TODO: verify requests are really from Twilio (https://godoc.org/github.com/kevinburke/twilio-go#GetExpectedTwilioSignature)
 */
func CallHandler(w http.ResponseWriter, r *http.Request) {
	var vr twiml.VoiceRequest
	if err := twiml.Bind(&vr, r); err != nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	log.Printf("Incoming call from %s", vr.From)

	response := twiml.NewResponse()
	response.Add(&twiml.Say{
		Text: greeting,
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

	inbound := NewClient().Table(inboundTable)
	if err := inbound.Create(&Inbound{
		Fields: InboundFields{
			TwilioID: vr.CallSid,
			Number:   vr.From,
		},
	}); err != nil {
		log.Printf("erroring creating inbound %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}
}

func RecordingHandler(w http.ResponseWriter, r *http.Request) {
	var recording twiml.RecordActionRequest
	if err := twiml.Bind(&recording, r); err != nil {
		log.Printf("erroring receiving recording %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	log.Printf("Received recording from %s", recording.From)

	inbound := NewClient().Table("inbound")

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

	log.Printf("Received transcription from %s", transcription.From)

	inbound := NewClient().Table("inbound")

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

func NewClient() *airtable.Client {
	return &airtable.Client{
		APIKey: os.Getenv("AIRTABLE_API_KEY"),
		BaseID: os.Getenv("AIRTABLE_BASE_ID"),
	}
}
