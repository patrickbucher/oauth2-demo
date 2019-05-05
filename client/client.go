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

func main() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "gossip.ico")
	})
	http.HandleFunc("/gossip", handleGossip)
	http.HandleFunc("/callback/", handleCallback)
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
	output, err := requestGossip(username, w, r)
	if err != nil {
		return
	}
	gossipTemplate := getGossipTemplate("gossip.html")
	gossipTemplate.Execute(w, output)
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
	log.Println(getGossipURL)
	get, err := http.NewRequest("GET", getGossipURL, nil)
	if err != nil {
		log.Printf("create GET request to %s: %v", getGossipURL, err)
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
		log.Println(err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode == http.StatusSeeOther {
		redirectURL := resp.Header.Get("Location")
		log.Printf("forwarded to %s", redirectURL)
		http.Redirect(w, r, redirectURL, http.StatusSeeOther)
		return nil, redirected
	}

	if resp.StatusCode != http.StatusOK {
		msg := fmt.Sprintf("unexpected status code %d", resp.StatusCode)
		log.Printf(msg)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, errors.New(msg)
	}
	decoder := json.NewDecoder(resp.Body)
	var gossip []string
	if err := decoder.Decode(&gossip); err != nil {
		fmt.Println(err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return nil, err
	}
	return &GossipOutput{username, gossip}, nil
}

func handleCallback(w http.ResponseWriter, r *http.Request) {
	scope, err := commons.ExtractPathElement(r.URL.Path, 1)
	if err != nil {
		log.Printf("extract first path element of %s: %v\n", r.URL.Path, err)
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
		log.Println("state", state, "not in pending requests")
		status := http.StatusBadRequest
		http.Error(w, http.StatusText(status), status)
		return
	}
	log.Println("ready to request access_token from authserver", scope, authHost, authPort, authCode)
	requestTokenRawURL := fmt.Sprintf("http://%s:%s/token", authHost, authPort)
	bodyParams := url.Values{}
	bodyParams.Set("grant_type", "authorization_code")
	bodyParams.Set("authorization_code", authCode)
	encodedBody := bodyParams.Encode()
	post, err := http.NewRequest("POST", requestTokenRawURL, strings.NewReader(encodedBody))
	log.Println("body", bodyParams.Encode())
	if err != nil {
		log.Printf("building POST request: %v\n", err)
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
		log.Printf("executing POST request: %v\n", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Printf("getting access token: %d (%s)\n", resp.StatusCode, http.StatusText(resp.StatusCode))
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	var token commons.AccessToken
	if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
		log.Printf("unmarshal access token: %v\n", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	log.Println("received token", token)
	accessTokens[username] = token.AccessToken
	gossip, err := requestGossip(username, w, r)
	if err != nil {
		// TODO: 403: remove access token from list, retry authorization
		log.Println("request gossip", err)
		status := http.StatusInternalServerError
		http.Error(w, http.StatusText(status), status)
		return
	}
	gossipTemplate := getGossipTemplate("gossip.html")
	gossipTemplate.Execute(w, gossip)
}

func getGossipTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("gossip").Parse(string(htmlTemplate)))
}
