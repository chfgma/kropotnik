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
	"k8s.io/klog/klogr"
)

const (
	inboundTable = "inbound"
	greeting     = "https://storage.googleapis.com/artifacts.chfgma.appspot.com/assets/chad3.wav"
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
	log := klogr.New()

	var vr twiml.VoiceRequest
	if err := twiml.Bind(&vr, r); err != nil {
		log.Error(err, "error decoding request")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	log = log.WithValues("from", vr.From, "callSid", vr.CallSid)
	log.Info("Incoming call")

	response := twiml.NewResponse()
	response.Add(&twiml.Play{
		URL:    greeting,
		Digits: "w",
	})

	response.Add(&twiml.Record{
		Action:             "/record",
		TranscribeCallback: "/transcription",
	})

	b, err := response.Encode()
	if err != nil {
		log.Error(err, "error encoding body")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, "error writing response")
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
		log.Error(err, "erroring creating inbound")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}
}

func RecordingHandler(w http.ResponseWriter, r *http.Request) {
	log := klogr.New()

	var recording twiml.RecordActionRequest
	if err := twiml.Bind(&recording, r); err != nil {
		log.Error(err, "erroring receiving recording")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	log = log.WithValues("from", recording.From, "callSid", recording.CallSid)
	log.Info("Received recording", "url", recording.RecordingURL)

	record := &Inbound{
		Fields: InboundFields{
			TwilioID:  recording.CallSid,
			Number:    recording.From,
			Recording: recording.RecordingURL + ".mp3",
		},
	}

	if err := createOrUpdate(recording.CallSid, record); err != nil {
		log.Error(err, "erroring creating inbound")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	response := twiml.NewResponse()
	response.Add(&twiml.Hangup{})

	b, err := response.Encode()
	if err != nil {
		log.Error(err, "error encoding body")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, "error writing response")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func TranscriptionHandler(w http.ResponseWriter, r *http.Request) {
	log := klogr.New()

	var transcription twiml.TranscribeCallbackRequest
	if err := twiml.Bind(&transcription, r); err != nil {
		log.Error(err, "erroring receiving transcription")
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	log = log.WithValues("from", transcription.From, "callSid", transcription.CallSid)
	log.Info("Received transcription", "transcription", transcription.TranscriptionText)

	entry := Inbound{
		Fields: InboundFields{
			TwilioID:      transcription.CallSid,
			Number:        transcription.From,
			Recording:     transcription.RecordingURL + ".mp3",
			Transcription: transcription.TranscriptionText,
		},
	}

	if err := createOrUpdate(transcription.CallSid, &entry); err != nil {
		log.Error(err, "erroring creating inbound")
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
	}

	response := twiml.NewResponse()
	response.Add(&twiml.Hangup{})

	b, err := response.Encode()
	if err != nil {
		log.Error(err, "error encoding body")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Error(err, "error writing response")
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func NewClient() *airtable.Client {
	return &airtable.Client{
		APIKey: os.Getenv("AIRTABLE_API_KEY"),
		BaseID: os.Getenv("AIRTABLE_BASE_ID"),
	}
}

func createOrUpdate(callSid string, record *Inbound) error {
	inbound := NewClient().Table("inbound")
	var inbounds []Inbound

	err := retry.Do(
		func() error {
			if err := inbound.List(&inbounds, &airtable.Options{
				Filter:     fmt.Sprintf(`{Twilio ID} = %q`, callSid),
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
	)
	if err != nil {
		log.Printf("error listing inbound records %v for SID %q", err, callSid)
	}

	if len(inbounds) == 0 {
		log.Printf("could not preexisting entry for for SID %q", callSid)

		if err := inbound.Create(record); err != nil {
			return err
		}

		return nil
	}

	record.ID = inbounds[0].ID

	if err := inbound.Update(record); err != nil {
		return err
	}
	return nil
}
