package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"os"
	"time"

	"github.com/frioux/amygdala/internal/middleware"
	"github.com/frioux/amygdala/internal/notes"
	"github.com/frioux/amygdala/internal/twilio"
)

var (
	dropboxAccessToken, myCell string
)

var twilioAuthToken, twilioURL []byte

func init() {
	rand.Seed(time.Now().UnixNano())

	dropboxAccessToken = os.Getenv("DROPBOX_ACCESS_TOKEN")
	if dropboxAccessToken == "" {
		panic("dropbox token is missing")
	}

	myCell = os.Getenv("MY_CELL")
	if myCell == "" {
		myCell = "+15555555555"
	}

	twilioAuthToken = []byte(os.Getenv("TWILIO_AUTH_TOKEN"))
	if len(twilioAuthToken) == 0 {
		twilioAuthToken = []byte("xyzzy")
	}

	twilioURL = []byte(os.Getenv("TWILIO_URL"))
	if len(twilioURL) == 0 {
		twilioURL = []byte("http://localhost:8080/twilio")
	}
}

var port int

func init() {
	flag.IntVar(&port, "port", 8080, "port to listen on")
}

func main() {
	flag.Parse()
	cl := &http.Client{}

	http.Handle("/twilio", middleware.Adapt(receiveSMS(cl, dropboxAccessToken),
		middleware.Log(os.Stdout),
	))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func receiveSMS(cl *http.Client, tok string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			io.WriteString(rw, "Couldn't Parse Form")
			return
		}

		if ok, err := twilio.CheckMAC(twilioAuthToken, twilioURL, r); err != nil || !ok {
			rw.WriteHeader(403)
			return
		}

		if r.Form.Get("From") != myCell {
			rw.WriteHeader(http.StatusForbidden)
			io.WriteString(rw, "Wrong Cell\n")
			return
		}

		message := r.Form.Get("Body")
		if message == "" {
			rw.WriteHeader(http.StatusBadRequest)
			io.WriteString(rw, "No Message\n")
			return
		}

		resp, err := notes.Dispatch(cl, tok, message)
		if err != nil {
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}

		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("Content-Type", "text/plain")

		io.WriteString(rw, resp+"\n")
	}
}
