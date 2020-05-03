package calls

import (
	"log"
	"net/http"
	"os"

	"github.com/BTBurke/twiml"
)

func CallForward(w http.ResponseWriter, r *http.Request) {
	var vr twiml.VoiceRequest
	if err := twiml.Bind(&vr, r); err != nil {
		log.Printf("error decoding request %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	log.Printf("Incoming call from %s", vr.From)

	response := twiml.NewResponse()
	response.Add(&twiml.Say{
		Text: "Please enter your pin code, followed by the pound key"})

	response.Add(&twiml.Gather{
		Input:  "dtmf",
		Action: "forward_verify",
	})

	b, err := response.Encode()
	if err != nil {
		log.Printf("error encoding body %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("error writing response %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func CallForwardVerify(w http.ResponseWriter, r *http.Request) {
	var vr twiml.RecordActionRequest
	if err := twiml.Bind(&vr, r); err != nil {
		log.Printf("error decoding request %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	if vr.Digits != os.Getenv("CALL_FORWARD_SECRET") {
		log.Printf("invalid call forwarding secret %q", vr.Digits)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
	}

	response := twiml.NewResponse()
	response.Add(&twiml.Say{
		Text: "Please enter the number you are calling, followed by the pound key"})

	response.Add(&twiml.Gather{
		Input:  "dtmf",
		Action: "forward_complete",
	})

	b, err := response.Encode()
	if err != nil {
		log.Printf("error encoding body %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("error writing response %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func CallForwardComplete(w http.ResponseWriter, r *http.Request) {
	var vr twiml.RecordActionRequest
	if err := twiml.Bind(&vr, r); err != nil {
		log.Printf("error decoding request %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	response := twiml.NewResponse()
	response.Add(&twiml.Dial{
		Number:   vr.Digits,
		Action:   "forward_error",
		CallerID: vr.To,
	})

	b, err := response.Encode()
	if err != nil {
		log.Printf("error encoding body %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("error writing response %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}

func CallForwardError(w http.ResponseWriter, r *http.Request) {
	var vr twiml.DialActionRequest
	if err := twiml.Bind(&vr, r); err != nil {
		log.Printf("error decoding request %v", err)
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)

		return
	}

	response := twiml.NewResponse()

	switch vr.DialCallStatus {
	case "completed":
		response.Add(&twiml.Say{
			Text: "call successful, thank you",
		})
	case "busy":
		response.Add(&twiml.Say{
			Text: "the number was busy, please try again later",
		})
	case "no-answer":
		response.Add(&twiml.Say{
			Text: "the number did not answer, please try again later",
		})
	case "failed":
		response.Add(&twiml.Say{
			Text: "the attempt to dial failed, this likely means the number we dialed was not a valid number",
		})
	default:
		log.Printf("unknown call status %q", vr.DialCallStatus)
	}

	b, err := response.Encode()
	if err != nil {
		log.Printf("error encoding body %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	if _, err := w.Write(b); err != nil {
		log.Printf("error writing response %v", err)
		http.Error(w, http.StatusText(http.StatusBadGateway), http.StatusBadGateway)

		return
	}

	w.Header().Set("Content-Type", "application/xml")
}
