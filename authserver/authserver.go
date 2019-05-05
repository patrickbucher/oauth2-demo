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

const accessTokenLifetime = "30s"

type tokenConstraint struct {
	expires time.Time
	scope   string
}

var issuedTokens = map[string]tokenConstraint{
	// access_token: {expiration, scope}
}

type AuthForm struct {
	CallbackURL string
	ClientID    string
}

func main() {
	http.HandleFunc("/authorization", handleAuthorization)
	http.HandleFunc("/token", handleToken)
	http.HandleFunc("/accesscheck", handleAccessCheck)
	http.HandleFunc("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "lock.ico")
	})
	info("listening on port 8443")
	http.ListenAndServe("0.0.0.0:8443", nil)
}

func handleAuthorization(w http.ResponseWriter, r *http.Request) {
	info("call /authorization")
	if r.Method == "GET" {
		info("show authorization form")
		showAuthorizationForm(w, r)
	} else if r.Method == "POST" {
		info("process authorization request")
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
		info(message)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	callbackURL, err := url.Parse(callbackRawURL)
	if err != nil {
		message := fmt.Sprintf("parse %s: %v", callbackRawURL, err)
		info(message)
		http.Error(w, message, http.StatusBadRequest)
		return
	}
	clientID := r.URL.Query().Get("client_id")
	loginForm := AuthForm{callbackURL.String(), clientID}
	info("callback URL: %s", callbackURL.String())
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
		info("client %s is unauthorized for  %s", clientID, username)
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	callbackURL, err := url.Parse(r.FormValue("callback_url"))
	if err != nil {
		info("missing field callback_url in form")
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
	info("redirect to %s", redirectURL.String())
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

func handleToken(w http.ResponseWriter, r *http.Request) {
	info("call /token")
	if r.Method != "POST" {
		httpCode := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	authHeader := r.Header.Get("Authorization")
	authHeaderParams := strings.Fields(authHeader)
	if len(authHeaderParams) != 2 || authHeaderParams[0] != "Basic" {
		info("Authorization header missing 'Basic' field")
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	sentCredentials, err := base64.RawURLEncoding.DecodeString(authHeaderParams[1])
	if err != nil {
		info("unable to base64 decode Authorization header %s", authHeaderParams[1])
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	credentials := strings.Split(string(sentCredentials), ":")
	if len(credentials) != 2 {
		info("credentials must be of the form client_id:client_secret, but is %s", credentials)
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	clientID, clientSecret := credentials[0], credentials[1]
	if secret, ok := clientCredentials[clientID]; !ok || secret != clientSecret {
		info("client %s is not authorized", clientID)
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	if err := r.ParseForm(); err != nil {
		info("error parsing form %v", err)
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	grantType := r.FormValue("grant_type")
	if grantType != "authorization_code" {
		info("grantType %s not supported", grantType)
		httpCode := http.StatusBadRequest
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	authCode := r.PostFormValue("authorization_code")
	username, ok := authorizationCodes[authCode]
	if !ok {
		info("authorization code %s invalid", authCode)
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	authorized := false
	for _, authorizedClient := range authorizedScopes[username] {
		if authorizedClient == clientID {
			authorized = true
			break
		}
	}
	if !authorized {
		info("client %s is not authorized for user %s", clientID, username)
		httpCode := http.StatusUnauthorized
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	} else {
		// code has been used once
		delete(authorizationCodes, authCode)
	}
	token := commons.Base64RandomString(32)
	accessToken := commons.AccessToken{AccessToken: token, TokenType: "Bearer"}
	duration, _ := time.ParseDuration(accessTokenLifetime)
	expiresAt := time.Now().Add(duration)
	issuedTokens[token] = tokenConstraint{expires: expiresAt, scope: username}
	w.Header().Add("Content-Type", "application/json")
	payload, _ := json.Marshal(accessToken)
	w.Write(payload)
}
func handleAccessCheck(w http.ResponseWriter, r *http.Request) {
	info("call /accesscheck")
	if r.Method != "POST" {
		httpCode := http.StatusMethodNotAllowed
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	accessToken := r.FormValue("access_token")
	scope := r.FormValue("scope")
	constraints, ok := issuedTokens[accessToken]
	if !ok {
		info("invalid access token %s", accessToken)
		httpCode := http.StatusForbidden
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	if scope != constraints.scope {
		info("invalid scope %s for access token %s", scope, accessToken)
		httpCode := http.StatusForbidden
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	if constraints.expires.Before(time.Now()) {
		info("access token expired at %v", constraints.expires)
		httpCode := http.StatusForbidden
		http.Error(w, http.StatusText(httpCode), httpCode)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func info(format string, args ...interface{}) {
	message := format
	if len(args) > 0 {
		message = fmt.Sprintf(format, args)
	}
	log.Println("[authserver]", message)
}
