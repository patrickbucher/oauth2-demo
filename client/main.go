package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
)

type GossipOutput struct {
	User   string
	Gossip []string
}

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
	// TODO: bring in after authN/authZ
	accessToken := r.FormValue("access_token")
	resp, err := client.Get("http://localhost:8000/gossip/" + username +
		"?access_token=" + accessToken)
	if err != nil {
		fmt.Println(err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther {
		redirectURL := resp.Header.Get("Location")
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return
	}
	if resp.StatusCode != http.StatusOK {
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
