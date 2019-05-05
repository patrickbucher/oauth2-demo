package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/patrickbucher/oauth2-demo/commons"
)

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

var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var authserverHost = os.Getenv("AUTHSERVER_HOST")
var authserverPort = os.Getenv("AUTHSERVER_PORT")

func main() {
	http.HandleFunc("/gossip/", handleGossip)
	info("listening on port %s", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
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
	info("call /gossip/%s", scope)
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
		info("redirect to %s", redirectURL.String())
		w.WriteHeader(http.StatusSeeOther)
		return
	}
	if validToken := validate(accessToken, scope); !validToken {
		info("invalid access token %s for scope %s", accessToken, scope)
		errCode := http.StatusForbidden
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	response, err := json.Marshal(gossip[scope])
	if err != nil {
		info("error marshalling gossip of %s: %v", scope, err)
		errCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(errCode), errCode)
		return
	}
	info("returning gossip of scope %s", scope)
	w.Write(response)
}

func validate(accessToken, scope string) bool {
	bodyParams := url.Values{}
	bodyParams.Set("access_token", accessToken)
	bodyParams.Set("scope", scope)
	encodedBody := bodyParams.Encode()
	authEndpoint := fmt.Sprintf("http://%s:%s/accesscheck", authserverHost, authserverPort)
	post, err := http.NewRequest("POST", authEndpoint, strings.NewReader(encodedBody))
	if err != nil {
		info("error creating POST request for token check: %v", err)
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
		return "", errors.New(`form must be "Bearer [access_token]"`)
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
	redirectRawURL := fmt.Sprintf("http://%s:%s/authorization?callback_url=%s&client_id=%s",
		authserverHost, authserverPort, url.QueryEscape(callbackURL.String()), clientID)
	return url.Parse(redirectRawURL)
}

func info(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args)
	}
	log.Println("[resource]", message)
}
