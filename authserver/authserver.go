package main

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickbucher/oauth2-demo/commons"
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

var authorizationCodes = map[string]string{
	// "auth_code": "scope" (username)
}

var authorizedScopes = map[string][]string{
	// "scope": {"client_id1", "client_id2", ...}
	"alice":   {},
	"bob":     {},
	"mallory": {},
}

const accessTokenLifetime = "5m"

var issuedTokens = map[string]time.Time{
	// access_token: expiration time
}

type AuthForm struct {
	CallbackURL string
	ClientID    string
}

func main() {
	http.HandleFunc("/authorization", handleAuthorization)
	http.HandleFunc("/token", handleToken)
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

func handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		httpCode := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	authHeader := r.Header.Get("Authorization")
	authHeaderParams := strings.Fields(authHeader)
	if len(authHeaderParams) != 2 || authHeaderParams[0] != "Basic" {
		log.Println("Authorization header missing 'Basic' field")
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	sentCredentials, err := base64.RawURLEncoding.DecodeString(authHeaderParams[1])
	if err != nil {
		log.Println("unable to base64 decode Authorization header", authHeaderParams[1])
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	credentials := strings.Split(string(sentCredentials), ":")
	if len(credentials) != 2 {
		log.Println("credentials must be of the form client_id:client_secret, but is", credentials)
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	clientID, clientSecret := credentials[0], credentials[1]
	if secret, ok := clientCredentials[clientID]; !ok || secret != clientSecret {
		log.Println("client", clientID, "not authorized")
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	if err := r.ParseForm(); err != nil {
		log.Println("error parsing form", err)
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	grantType := r.FormValue("grant_type")
	log.Println("grantType:", grantType)
	if grantType != "authorization_code" {
		log.Println("grantType", grantType, "not supported")
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	authCode := r.PostFormValue("authorization_code")
	if username, ok := authorizationCodes[authCode]; !ok {
		log.Println("authorization code", authCode, "invalid")
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	} else {
		authorized := false
		for _, authorizedClient := range authorizedScopes[username] {
			if authorizedClient == clientID {
				authorized = true
				break
			}
		}
		if !authorized {
			log.Println("client", clientID, "is not authorized for", username)
			httpCode := http.StatusUnauthorized
			http.Error(w, http.StatusText(httpCode), httpCode)
			return
		} else {
			// code has been used once
			delete(authorizationCodes, authCode)
		}
	}
	token := commons.Base64RandomString(32)
	accessToken := commons.AccessToken{AccessToken: token, TokenType: "Bearer"}
	duration, _ := time.ParseDuration(accessTokenLifetime)
	expiresAt := time.Now().Add(duration)
	issuedTokens[token] = expiresAt
	w.Header().Add("Content-Type", "application/json")
	payload, _ := json.Marshal(accessToken)
	w.Write(payload)
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
	authCode := commons.Base64RandomString(16)
	authorizationCodes[authCode] = username
	params := url.Values{"auth_host": {"localhost"}, "auth_port": {"8443"}, "auth_code": {authCode}}
	redirectURL, err := url.Parse(fmt.Sprintf("%s&%s", callbackURL.String(), params.Encode()))
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
