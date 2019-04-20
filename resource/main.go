package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
)

const authHost = "localhost:8443"

var gossip = map[string][]string{
	"alice": {
		"Oreos are made out of sand.",
		"Bob stinks.",
	},
	"bob": {
		"Larry Ellison would like to be bought by Oracle.",
		"Alice has a crush on me.",
	},
	"mallory": {
		"Obama is to blame for climate change.",
		"There's something going on between Alice and Bob.",
	},
}

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "resource.ico")
	})
	http.HandleFunc("/gossip/", handleGossip)
	log.Println("resource listening on port 8000")
	http.ListenAndServe("0.0.0.0:8000", nil)
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	username, err := extractUsername(r.URL.Path)
	if err != nil {
		errCode := http.StatusNotFound
		http.Error(w, err.Error(), errCode)
		return
	}
	log.Println("/gossip/" + username)
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		redirectURL, err := buildRedirectURL(r, username)
		if err != nil {
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		w.Header().Add("WWW-Authenticate", "bearer")
		w.Header().Add("Location", redirectURL.String())
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	// TODO: extract accessToken ("Authorization: Bearer [accessToken")
	// TODO validate accessToken against authserver, then continue
	response, err := json.Marshal(gossip[username])
	if err != nil {
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	w.Write(response)
}

func extractUsername(resource string) (string, error) {
	paths := strings.Split(resource, "/") // ["", "gossip", "[username]"]
	if len(paths) != 3 {
		return "", fmt.Errorf("resource '%s' must be /gossip/[username]", resource)
	}
	username := paths[2]
	if _, ok := gossip[username]; !ok {
		return "", fmt.Errorf("no gossip found for %s", username)
	}
	return username, nil
}

func buildRedirectURL(r *http.Request, username string) (*url.URL, error) {
	// TODO: is this id really needed?
	id := base64RandomString(32)
	callbackRawURL := "http://" + r.Host + "/callback/" + username + "?id=" + id
	callbackURL, err := url.Parse(callbackRawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL %s: %v", callbackRawURL, err)
	}
	redirectRawURL := "http://" + authHost + "/authorization?callback_url=" +
		url.QueryEscape(callbackURL.String())
	return url.Parse(redirectRawURL)
}

func base64RandomString(nBytes uint) string {
	data := make([]byte, nBytes)
	rand.Read(data)
	return base64.RawURLEncoding.EncodeToString(data)
}
