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

var clientCredentials = map[string]string{
	// "client_id": "client_secret"
	"gossip_client": "43897dfa-c910-4d3c-9851-5328cf49467d",
}

var credentials = map[string]string{
	// "username": "password"
	"alice":   "topsecret",
	"bob":     "1234",
	"mallory": "70p53cr37",
}

var authorizedScopes = map[string][]string{
	// "scope": {"client_id1", "client_id2", ...}
	"alice":   {},
	"bob":     {},
	"mallory": {},
}

type accessToken struct {
	ClientID string    `json:"client_id"`
	Username string    `json:"username"`
	Expires  time.Time `json:"expires"`
	TokenID  string    `json:"token_id"`
}

var issuedTokens = map[string]accessToken{
	// "clientId:username" : accessToken (store in map for fast lookup)
}

type AuthForm struct {
	CallbackURL string
	ClientID    string
}

func main() {
	http.HandleFunc("/authorization", handleAuthorization)
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
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "lock.ico")
	})
	log.Println("auth server listening on port 8443")
	http.ListenAndServe("0.0.0.0:8443", nil)
}

func handleAuthorization(w http.ResponseWriter, r *http.Request) {
	if r.Method == "GET" {
		log.Println("show authorization form")
		showAuthorizationForm(w, r)
	} else if r.Method == "POST" {
		log.Println("process authorization request")
		processAuthorization(w, r)
	} else {
		httpCode := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
}

func showAuthorizationForm(w http.ResponseWriter, r *http.Request) {
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
	clientID := r.URL.Query().Get("client_id")
	loginForm := AuthForm{callbackURL.String(), clientID}
	log.Println("callback URL", callbackURL.String())
	template.Execute(w, loginForm)
}

func processAuthorization(w http.ResponseWriter, r *http.Request) {
	username := r.FormValue("username")
	password := r.FormValue("password")
	if realPassword, ok := credentials[username]; !ok ||
		password != realPassword {
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	clientID := r.FormValue("client_id")
	if _, ok := clientCredentials[clientID]; ok {
		// there is a secret, and owner authorizes the scope for this client,
		// but the client_secret hasn't been checked yet!
		authorizedScopes[username] = append(authorizedScopes[username], clientID)
	} else {
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	callbackURL, err := url.Parse(r.FormValue("callback_url"))
	if err != nil {
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	coordinates := url.Values{"auth_host": {"localhost"}, "auth_port": {"8443"}}
	redirectURL, err := url.Parse(fmt.Sprintf("%s&%s", callbackURL.String(), coordinates.Encode()))
	if err != nil {
		httpCode := http.StatusInternalServerError
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	log.Println("redirect to", redirectURL.String())
	w.Header().Add("Location", redirectURL.String())
	w.WriteHeader(http.StatusSeeOther)
}

func getLoginTemplate(file string) *template.Template {
	htmlTemplate, err := ioutil.ReadFile(file)
	if err != nil {
		panic("error reading template " + file)
	}
	return template.Must(template.New("login").Parse(string(htmlTemplate)))
}
