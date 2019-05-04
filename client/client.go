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

func handleCallback(w http.ResponseWriter, r *http.Request) {
	// TODO: extract callback_url and client_secret
	// TODO: get an access_token from the auth server
	// TODO: how to get that URL again? client does not know the auth server,
	// must store somehow from the first redirect before following it along...?
	// resource issues redirectURL (to auth server) and callbackURL (this handler)
	// so authServer must attach its coordinates upon redirecting back to me
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

func getGossipTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("gossip").Parse(string(htmlTemplate)))
}
