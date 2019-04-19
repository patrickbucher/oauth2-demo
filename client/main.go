package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
)

type GossipOutput struct {
	User   string
	Gossip []string
}

func main() {
	htmlTemplate, err := ioutil.ReadFile("gossip.html")
	if err != nil {
		panic("error reading gossip.html template")
	}
	gossipTemplate := template.Must(template.New("gossip").Parse(string(htmlTemplate)))

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	http.HandleFunc("/gossip", func(w http.ResponseWriter, r *http.Request) {
		username := r.FormValue("username")
		if username == "" {
			status := http.StatusBadRequest
			http.Error(w, http.StatusText(status), status)
			return
		}
		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}
		accessToken := r.FormValue("accessToken") // TODO: bring in somehow
		resp, err := client.Get("http://localhost:8000/gossip/" + username +
			"?access_token=" + accessToken)
		if err != nil {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return
		}
		decoder := json.NewDecoder(resp.Body)
		var gossip []string
		if err := decoder.Decode(&gossip); err != nil {
			fmt.Println(err)
			status := http.StatusInternalServerError
			http.Error(w, http.StatusText(status), status)
			return
		}
		output := GossipOutput{username, gossip}
		gossipTemplate.Execute(w, output)
	})
	http.ListenAndServe("0.0.0.0:1234", nil)
}
