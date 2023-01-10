package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"strings"
	"time"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Version   int    `json:"version"`
	Type      string `json:"type"`
	Tailnet   string `json:"tailnet"`
	Message   string `json:"message"`
	Data      Data   `json:"data"`
}
type Data struct {
	NodeID     string `json:"nodeID,omitempty"`
	DeviceName string `json:"deviceName,omitempty"`
	ManagedBy  string `json:"managedBy,omitempty"`
	Actor      string `json:"actor,omitempty"`
	URL        string `json:"url,omitempty"`
}

type TSLogs []LogEntry

// Push represents a message sent to the Pushover api
type Push struct {
	Token     string    `json:"token"`
	User      string    `json:"user"`
	Message   string    `json:"message"`
	Device    string    `json:"device,omitempty"`
	Title     string    `json:"title"`
	URL       string    `json:"url"`
	URLTitle  string    `json:"url_title,omitempty"`
	Priority  int       `json:"priority,omitempty"`
	Sound     string    `json:"sound,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

// PushResponse is a response from the Pushover api
type PushResponse struct {
	Status  int      `json:"status,omitempty"`
	Request string   `json:"request,omitempty"`
	User    string   `json:"user,omitempty"`
	Errors  []string `json:"errors,omitempty"`
}

type apiHandler struct{}

var (
	appToken  = os.Getenv("PUSHOVER_TOKEN")
	userToken = os.Getenv("PUSHOVER_USER")
	potsToken = os.Getenv("POTS_TOKEN")
	client    = *http.DefaultClient
)

func sendPush(p *Push) error {
	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(p); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://api.pushover.net/1/messages.json", buf)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	var resBody PushResponse
	if err = json.NewDecoder(res.Body).Decode(&resBody); err != nil {
		return err
	}

	if len(resBody.Errors) > 0 {
		return fmt.Errorf("%s", strings.Join(resBody.Errors, ", "))
	}

	return nil
}

func (apiHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != "POST" {
		http.NotFound(w, req)
		return
	}

	token := path.Base(req.URL.Path)
	if token != potsToken {
		http.NotFound(w, req)
		return
	}

	decoder := json.NewDecoder(req.Body)
	var tsLogs TSLogs
	err := decoder.Decode(&tsLogs)
	if err != nil {
		http.Error(w, "invalid data", http.StatusBadRequest)
		return
	}

	for _, t := range tsLogs {
		var push = &Push{
			Token:     appToken,
			User:      userToken,
			Timestamp: time.Now(),
			Title:     t.Message,
			Message:   t.Type,
			URL:       t.Data.URL,
		}
		err := sendPush(push)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	listen := flag.String("listen", ":8888", "listen string")
	flag.Parse()

	mux := http.NewServeMux()
	mux.Handle("/api/", apiHandler{})
	mux.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		log.Println(req.URL.Path)
		if req.URL.Path != "/" {
			http.NotFound(w, req)
			return
		}
		fmt.Fprintf(w, `<a href="https://suah.dev/pots">suah.dev/pots</a>`)
	})

	s := &http.Server{
		Addr:    *listen,
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())
}
