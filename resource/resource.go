package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"

	"github.com/patrickbucher/oauth2-demo/commons"
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
	clientId := r.URL.Query().Get("client_id")
	authHeader := r.Header.Get("Authorization")
	accessToken, err := extractAccessToken(authHeader)
	if accessToken == "" || err != nil {
		redirectURL, err := buildRedirectURL(r, username, clientId)
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
	// TODO: extract accessToken ("Authorization: Bearer [accessToken]")
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
	if len(paths) < 3 {
		return "", fmt.Errorf("resource '%s' must be /gossip/[username]", resource)
	}
	username := paths[2]
	if _, ok := gossip[username]; !ok {
		return "", fmt.Errorf("no gossip found for %s", username)
	}
	return username, nil
}

func extractAccessToken(authorizationHeader string) (string, error) {
	fields := strings.Fields(authorizationHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return "", fmt.Errorf(`form must be "Bearer [access_token]"`)
	}
	return fields[1], nil
}

func buildRedirectURL(r *http.Request, username, clientId string) (*url.URL, error) {
	// TODO: is this id really needed?
	id := commons.Base64RandomString(32)
	callbackRawURL := "http://" + r.Host + "/callback/" + username + "?id=" + id
	callbackURL, err := url.Parse(callbackRawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL %s: %v", callbackRawURL, err)
	}
	// TODO: build up URL proprely (not through string concatenation)
	redirectRawURL := "http://" + authHost + "/authorization?callback_url=" +
		url.QueryEscape(callbackURL.String()) + "&client_id=" + clientId
	return url.Parse(redirectRawURL)
}
