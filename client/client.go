package main

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"

	"github.com/patrickbucher/oauth2-demo/commons"
)

const (
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

var host = os.Getenv("HOST")
var port = os.Getenv("PORT")
var resourceHost = os.Getenv("RESOURCE_HOST")
var resourcePort = os.Getenv("RESOURCE_PORT")

func main() {
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/callback/", handleCallback)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gossip.ico")
	})
	info("listening on port %s", port)
	http.ListenAndServe("0.0.0.0:"+port, nil)
}

func handleGossip(w http.ResponseWriter, r *http.Request) {
	info("call /gossip")
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
	info("call /callback")
	scope, err := commons.ExtractPathElement(r.URL.Path, 1)
	if err != nil {
		info("extract first path element of %s: %v", r.URL.Path, err)
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
		info("state %s not in pending requests", state)
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	info("request access_token for scope %s from authserver %s:%s with authCode %s",
		scope, authHost, authPort, authCode)
	requestTokenRawURL := fmt.Sprintf("http://%s:%s/token", authHost, authPort)
	bodyParams := url.Values{}
	bodyParams.Set("grant_type", "authorization_code")
	bodyParams.Set("authorization_code", authCode)
	encodedBody := bodyParams.Encode()
	post, err := http.NewRequest("POST", requestTokenRawURL, strings.NewReader(encodedBody))
	if err != nil {
		info("error building POST request: %v", err)
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
		info("error executing POST request: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		info("getting access token: %d (%s)", resp.StatusCode, http.StatusText(resp.StatusCode))
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	var token commons.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		info("unmarshal access token: %v", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	info("received token %s", token)
	accessTokens[username] = token.AccessToken
	gossip, err := requestGossip(username, w, r)
	if err != nil {
		// TODO HTTP status 403: remove access token from list, retry authorization
		info("error requesting gossip: %v", err)
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
	getGossipURL := fmt.Sprintf("http://%s:%s/gossip/%s?host=%s&port=%s&client_id=%s&state=%s",
		resourceHost, resourcePort, username, host, port, clientID, state)
	get, err := http.NewRequest("GET", getGossipURL, nil)
	if err != nil {
		info("create GET request to %s: %v", getGossipURL, err)
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
		info("perform GET request to %s: %v", getGossipURL, err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther {
		redirectURL := resp.Header.Get("Location")
		info("forwarded to %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, redirected
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected status code %d", resp.StatusCode)
		info(msg)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, errors.New(msg)
	}
	decoder := json.NewDecoder(resp.Body)
	var gossip []string
	if err := decoder.Decode(&gossip); err != nil {
		info("error decoding JSON response: %v", err)
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

func info(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args)
	}
	log.Println("[client]", message)
}
