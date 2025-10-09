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

// Push represents a message sent to the Gotify api
type Push struct {
	Message  string `json:"message"`
	Title    string `json:"title"`
	Priority int    `json:"priority"`
}

// PushResponse is a response from the Gotify api
type PushResponse struct {
	AppID      int       `json:"appid,omitempty"`
	Date       time.Time `json:"date,omitempty"`
	Error      string    `json:"error,omitempty"`
	ErrorCode  int       `json:"errorCode,omitempty"`
	ErrorDescr string    `json:"errorDescription",omitempty"`
}

type apiHandler struct{}

var (
	appToken  = os.Getenv("PUSHOVER_TOKEN")
	potsToken = os.Getenv("POTS_TOKEN")
	client    = *http.DefaultClient
)

func sendPush(p *Push) error {
	buf := new(bytes.Buffer)

	if err := json.NewEncoder(buf).Encode(p); err != nil {
		return err
	}

	req, err := http.NewRequest("POST", "https://notify.otter-alligator.ts.net/message", buf)
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

	if resBody.Error != "" {
		return fmt.Errorf("%s", strings.Join([]string{
			fmt.Sprintf("Error Code: %d", resBody.ErrorCode),
			resBody.Error,
			resBody.ErrorDescr,
		}, ", "))
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
			Title:   t.Message,
			Message: t.Type,
		}
		err := sendPush(push)
		if err != nil {
			log.Println(err)
		}
	}
}

func main() {
	port := flag.Int("port", 8888, "port to listen on")
	ip := flag.String("ip", "127.0.0.1", "ip to listen on")
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
		Addr:    fmt.Sprintf("%s:%d", *ip, *port),
		Handler: mux,
	}

	log.Fatal(s.ListenAndServe())
}
