package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/patrickbucher/oauth2-demo/commons"
)

const (
	resourceHost = "localhost:8000"
	clientID     = "gossip_client"
	clientSecret = "43897dfa-c910-4d3c-9851-5328cf49467d"
)

type GossipOutput struct {
	User   string
	Gossip []string
}

var pendingRequests = map[string]bool{}

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gossip.ico")
	})
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/callback", handleCallback)
	log.Println("client listening on port 1234")
	http.ListenAndServe("0.0.0.0:1234", nil)
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	if username == "" {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	state := commons.Base64RandomString(16)
	getGossipURL := fmt.Sprintf("http://%s/gossip/%s?host=%s&port=%d&client_id=%s&state=%s",
		resourceHost, username, "localhost", 1234, clientID, state)
	log.Println(getGossipURL)
	get, err := http.NewRequest("GET", getGossipURL, nil)
	if err != nil {
		log.Printf("create GET request to %s: %v", getGossipURL, err)
		httpCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(httpCode), httpCode)
	}
	resp, err := client.Do(get)
	pendingRequests[state] = true
	if err != nil {
		log.Println(err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther {
		redirectURL := resp.Header.Get("Location")
		log.Printf("forwarded to %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}

	// TODO: this down here is for later...
	// get.Header.Add("Authorization", "Bearer "+accessToken)

	if resp.StatusCode != http.StatusOK {
		log.Printf("unexpected status code %d", resp.StatusCode)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	gossipTemplate := getGossipTemplate("gossip.html")
	decoder := json.NewDecoder(resp.Body)
	var gossip []string
	if err := decoder.Decode(&gossip); err != nil {
		fmt.Println(err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	output := GossipOutput{username, gossip}
	gossipTemplate.Execute(w, output)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	scope, err := commons.ExtractPathElement(r.URL.Path, uint(1))
	if err != nil {
		log.Printf("extract first path element of %s: %v\n", r.URL.Path, err)
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	authHost := r.URL.Query().Get("auth_host")
	authPort := r.URL.Query().Get("auth_host")
	state := r.URL.Query().Get("state")
	if _, ok := pendingRequests[state]; !ok {
		log.Println("state", state, "not in pending requests")
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	log.Println("ready to request access_token from authserver", scope, authHost, authPort)
	// TODO: get an access_token from the auth server
}

func getGossipTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("gossip").Parse(string(htmlTemplate)))
}
