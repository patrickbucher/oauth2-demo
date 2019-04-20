package main

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var credentials = map[string]string{
	// "username": "password"
	"alice":   "topsecret",
	"bob":     "1234",
	"mallory": "70p53cr37",
}

var clients = map[string]string{
	// "client_id": "client_secret"
}

var authorizedClients = map[string][]string{
	// "username": {"client_id1", "client_id2", ...}
	"alice":   {},
	"bob":     {},
	"mallory": {},
}

type accessToken struct {
	clientId string    `json:"client_id"`
	username string    `json:"username"`
	expires  time.Time `json:"expires"`
	tokenId  string    `json:"token_id"`
}

var issuedTokens = map[string]accessToken{
	// "clientId:username" : accessToken (store in map for fast lookup)
}

type LoginForm struct {
	CallbackURL string
}

func main() {
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "lock.ico")
	})
	http.HandleFunc("/authorization", func(w http.ResponseWriter, r *http.Request) {
		if r.Method == "GET" {
			template := getLoginTemplate("auth.html")
			callbackEscapedURL := r.URL.Query().Get("callback_url")
			callbackRawURL, err := url.QueryUnescape(callbackEscapedURL)
			if err != nil {
				message := fmt.Sprintf("unescape %s: %v", callbackRawURL, err)
				http.Error(w, message, http.StatusBadRequest)
				return
			}
			callbackURL, err := url.Parse(callbackRawURL)
			if err != nil {
				message := fmt.Sprintf("parse %s: %v", callbackRawURL, err)
				http.Error(w, message, http.StatusBadRequest)
				return
			}
			loginForm := LoginForm{callbackURL.String()}
			log.Println("callback URL", callbackURL.String())
			template.Execute(w, loginForm)
			// TODO: could username be pre-filled?
			return
		}
		if r.Method != "POST" {
			httpCode := http.StatusMethodNotAllowed
			http.Error(w, http.StatusText(httpCode), httpCode)
			return
		}
		// TOOD
		// form params: username, password, client_id
		// check if credentials[username] == password
		// retrieve client_secret for client_id in clients map...
		// ... or issue new client_secret and store it in clients map:
		// clients[client_id] = client_secret
		// authorize client to user: authorizedClients[username] = client_id
	})
	http.HandleFunc("/token", func(w http.ResponseWriter, r *http.Request) {
		// check if clients[client_id] == client_secret
		// check if authorizedClients[username] contains client_id
		// issue new accessToken and store it in issuedTokens
		// - tokenId are some random, base64 encoded bytes
		// - serialize accessToken as JSON
	})
	http.HandleFunc("/accesscheck", func(w http.ResponseWriter, r *http.Request) {
		// convert token from base64 string to JSON string
		// unmarshal JSON structure to accessToken struct
		// build lookup key clientId:username
		// retrieve all access tokens for this client/user combination
		// check for each key if tokenId is matching
		// if so, check if the token is not expired yet
		// if so, return status 200
		// otherwise, return status 403 (or better: 404?)
	})
	log.Println("auth server listening on port 8443")
	http.ListenAndServe("0.0.0.0:8443", nil)
}

func getLoginTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("login").Parse(string(htmlTemplate)))
}
