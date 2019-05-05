package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
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
	scope, err := commons.ExtractPathElement(r.URL.Path, 1) // /gossip/[username]
	if err != nil {
		errCode := http.StatusBadRequest
		http.Error(w, err.Error(), errCode)
		return
	}
	if _, ok := gossip[scope]; !ok {
		errCode := http.StatusNotFound
		http.Error(w, http.StatusText(errCode), errCode)
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
	if validToken := validate(accessToken, scope); !validToken {
		log.Println("access token invalid")
		errCode := http.StatusForbidden
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	log.Println("valid access token")
	response, err := json.Marshal(gossip[scope])
	if err != nil {
		log.Println("marshal gossip of", scope, err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	log.Println("returning gossip of", scope)
	w.Write(response)
}

func validate(accessToken, scope string) bool {
	bodyParams := url.Values{}
	bodyParams.Set("access_token", accessToken)
	bodyParams.Set("scope", scope)
	encodedBody := bodyParams.Encode()
	authEndpoint := "http://" + authHost + "/accesscheck"
	post, err := http.NewRequest("POST", authEndpoint, strings.NewReader(encodedBody))
	if err != nil {
		log.Println("error creating POST request for token check", err)
		return false
	}
	post.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	post.Header.Add("Content-Length", strconv.Itoa(len(encodedBody)))
	client := &http.Client{}
	resp, err := client.Do(post)
	if err != nil || resp.StatusCode != http.StatusOK {
		return false
	}
	return true
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
