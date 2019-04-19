package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
)

var gossip = map[string][]string{
	"alice": {
		"Oreos are made out of sand.",
		"Bob stinks.",
	},
	"bob": {
		"Larry Ellison would like to be bought by Oracle",
		"Alice has a crush on me.",
	},
	"mallory": {
		"Obama is to blame for climate change.",
		"There's something going on between Alice and Bob.",
	},
}

func main() {
	http.HandleFunc("/gossip/", func(w http.ResponseWriter, r *http.Request) {
		paths := strings.Split(r.URL.Path, "/") // ["", "gossip", "[username]"]
		if len(paths) != 3 {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		username := paths[2]
		if _, ok := gossip[username]; !ok {
			http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
			return
		}
		log.Println("/gossip/" + username)
		// TODO: read access_token from "Authorization" header
		accessToken := r.URL.Query().Get("access_token")
		if accessToken == "" {
			w.Header().Add("WWW-Authenticate", "bearer")
			// TODO: forward to authserver
			return
		}
		// TODO: validate accessToken against authserver
		response, _ := json.Marshal(gossip[username])
		w.Write(response)
	})
	http.ListenAndServe("0.0.0.0:8000", nil)
}
