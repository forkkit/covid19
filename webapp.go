package main

import (
	"encoding/json"
	"github.com/amnonbc/covid19/who"
	"html/template"
	"log"
	"net/http"
	"time"
)

var tmpl = template.Must(template.ParseFiles("templates/top.html"))
var whoStats []who.RawData

func homepage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.Method, r.URL)
	tmpl.Execute(w, nil)
}

func stats(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Data []who.RawData `json:"data"`
	}

	payload.Data = who.Latest(whoStats)
	json.NewEncoder(w).Encode(payload)
}

func favicon(w http.ResponseWriter, r *http.Request) {
	http.ServeFile(w, r, "static/favicon.ico")
}

func updater() {
	for {
		ws, err := who.DownloadCSV(who.WhoSourceURL)
		if err != nil {
			log.Println(err)
		} else {
			whoStats = ws
		}
		time.Sleep(6 * time.Hour)
	}
}

func main() {
	go updater()
	http.HandleFunc("/", homepage)
	http.HandleFunc("/stats.json", stats)
	http.HandleFunc("/favicon.ico", favicon)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
