package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"

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

var accessTokens = map[string]string{
	// "username":"access_token"
}

var pendingRequests = map[string]string{
	// "state":"username"
}

var redirected = errors.New("redirected")

var log = commons.Logger("client")

func main() {
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/callback/", handleCallback)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gossip.ico")
	})
	log("listening on port 1234")
	http.ListenAndServe("0.0.0.0:1234", nil)
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	log("call /gossip")
	username := r.FormValue("username")
	if username == "" {
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	output, err := requestGossip(username, w, r)
	if err != nil {
		return
	}
	gossipTemplate := getGossipTemplate("gossip.html")
	gossipTemplate.Execute(w, output)
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	log("call /callback")
	scope, err := commons.ExtractPathElement(r.URL.Path, 1)
	if err != nil {
		log("extract first path element of %s: %v", r.URL.Path, err)
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	authHost := r.URL.Query().Get("auth_host")
	authPort := r.URL.Query().Get("auth_port")
	authCode := r.URL.Query().Get("auth_code")
	state := r.URL.Query().Get("state")
	username, ok := pendingRequests[state]
	if !ok {
		log("state %s not in pending requests", state)
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	log("request access_token for scope %s from authserver %s:%s with authCode %s",
		scope, authHost, authPort, authCode)
	requestTokenRawURL := fmt.Sprintf("http://%s:%s/token", authHost, authPort)
	bodyParams := url.Values{}
	bodyParams.Set("grant_type", "authorization_code")
	bodyParams.Set("authorization_code", authCode)
	encodedBody := bodyParams.Encode()
	post, err := http.NewRequest("POST", requestTokenRawURL, strings.NewReader(encodedBody))
	if err != nil {
		log("error building POST request: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	authHeader := base64.RawURLEncoding.EncodeToString([]byte(clientID + ":" + clientSecret))
	post.Header.Add("Authorization", "Basic "+authHeader)
	post.Header.Add("Accept", "application/json")
	post.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	post.Header.Add("Content-Length", strconv.Itoa(len(encodedBody)))
	client := &http.Client{}
	resp, err := client.Do(post)
	if err != nil {
		log("error executing POST request: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log("getting access token: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	var token commons.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log("unmarshal access token: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	log("received token %s", token)
	accessTokens[username] = token.AccessToken
	gossip, err := requestGossip(username, w, r)
	if err != nil {
		// TODO HTTP status 403: remove access token from list, retry authorization
		log("error requesting gossip: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	gossipTemplate := getGossipTemplate("gossip.html")
	gossipTemplate.Execute(w, gossip)
}

func requestGossip(username string, w http.ResponseWriter, r *http.Request) (*GossipOutput, error) {
	client := &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	state := commons.Base64RandomString(16)
	getGossipURL := fmt.Sprintf("http://%s/gossip/%s?host=%s&port=%d&client_id=%s&state=%s",
		resourceHost, username, "localhost", 1234, clientID, state)
	get, err := http.NewRequest("GET", getGossipURL, nil)
	if err != nil {
		log("create GET request to %s: %v", getGossipURL, err)
		httpCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(httpCode), httpCode)
		return nil, err
	}
	if accessToken, ok := accessTokens[username]; ok {
		get.Header.Add("Authorization", "Bearer "+accessToken)
	}
	resp, err := client.Do(get)
	pendingRequests[state] = username
	if err != nil {
		log("perform GET request to %s: %v", getGossipURL, err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther {
		redirectURL := resp.Header.Get("Location")
		log("forwarded to %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, redirected
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected status code %d", resp.StatusCode)
		log(msg)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, errors.New(msg)
	}
	decoder := json.NewDecoder(resp.Body)
	var gossip []string
	if err := decoder.Decode(&gossip); err != nil {
		log("error decoding JSON response: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	return &GossipOutput{username, gossip}, nil
}

func getGossipTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("gossip").Parse(string(htmlTemplate)))
}
