package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/frioux/amygdala/internal/notes"
	"golang.org/x/crypto/bcrypt"
)

var (
	dropboxAccessToken, myCell string
)

var pass []byte

func init() {
	dropboxAccessToken = os.Getenv("DROPBOX_ACCESS_TOKEN")
	if dropboxAccessToken == "" {
		panic("dropbox token is missing")
	}

	myCell = os.Getenv("MY_CELL")
	if myCell == "" {
		panic("cell is missing")
	}

	pass = []byte(os.Getenv("TWILIO_PASSWORD"))
	if len(pass) == 0 {
		panic("password is missing")
	}
}

var port int

func init() {
	flag.IntVar(&port, "port", 8080, "port to listen on")
}

func main() {
	flag.Parse()
	cl := &http.Client{}

	http.HandleFunc("/twilio", twilio(cl, dropboxAccessToken))

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%d", port), nil))
}

func twilio(cl *http.Client, tok string) http.HandlerFunc {
	return func(rw http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			io.WriteString(rw, "Couldn't Parse Form")
			return
		}

		if bcrypt.CompareHashAndPassword(pass, []byte(r.Form.Get("Authorization"))) != nil {
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

		if err := notes.Todo(cl, tok, message); err != nil {
			rw.WriteHeader(http.StatusInternalServerError)

			panic(err)
		}

		rw.WriteHeader(http.StatusOK)
		rw.Header().Set("Content-Type", "application/xml")
		io.WriteString(rw, `<?xml version="1.0" encoding="UTF-8"?>
			<Response><Message><Body>got em.</Body></Message></Response>`)
	}
}
