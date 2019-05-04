package main

import (
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
	http.HandleFunc("/gossip/", handleGossip)
	log.Println("resource listening on port 8000")
	http.ListenAndServe("0.0.0.0:8000", nil)
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	scope, err := extractScope(r.URL.Path)
	if err != nil {
		errCode := http.StatusNotFound
		http.Error(w, err.Error(), errCode)
		return
	}
	log.Println("/gossip/" + scope)
	params := r.URL.Query()
	remoteHost := params.Get("host")
	remotePort := params.Get("port")
	clientID := params.Get("client_id")
	state := params.Get("state")
	authHeader := r.Header.Get("Authorization")
	accessToken, err := extractAccessToken(authHeader)
	if accessToken == "" || err != nil {
		redirectURL, err := buildRedirectURL(remoteHost, remotePort, scope, clientID, state)
		if err != nil {
			errCode := http.StatusInternalServerError
			http.Error(w, http.StatusText(errCode), errCode)
			return
		}
		w.Header().Add("WWW-Authenticate", "bearer")
		w.Header().Add("Location", redirectURL.String())
		log.Println("redirect to", redirectURL.String())
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	// TODO validate accessToken against authserver, then continue
	response, err := json.Marshal(gossip[scope])
	if err != nil {
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	w.Write(response)
}

func extractScope(resource string) (string, error) {
	paths := strings.Split(resource, "/") // ["", "gossip", "[scope]"]
	if len(paths) < 3 {
		return "", fmt.Errorf("resource '%s' must be /gossip/[scope]", resource)
	}
	scope := paths[2]
	if _, ok := gossip[scope]; !ok {
		return "", fmt.Errorf("no gossip found for %s", scope)
	}
	return scope, nil
}

func extractAccessToken(authorizationHeader string) (string, error) {
	fields := strings.Fields(authorizationHeader)
	if len(fields) != 2 || fields[0] != "Bearer" {
		return "", fmt.Errorf(`form must be "Bearer [access_token]"`)
	}
	return fields[1], nil
}

func buildRedirectURL(host, port, scope, clientID, state string) (*url.URL, error) {
	callbackRawURL := fmt.Sprintf("http://%s:%s/callback/%s?state=%s",
		host, port, scope, state)
	callbackURL, err := url.Parse(callbackRawURL)
	if err != nil {
		return nil, fmt.Errorf("parse URL %s: %v", callbackRawURL, err)
	}
	redirectRawURL := "http://" + authHost + "/authorization?callback_url=" +
		url.QueryEscape(callbackURL.String()) + "&client_id=" + clientID
	return url.Parse(redirectRawURL)
}
