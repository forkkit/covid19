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
var countryTemplate = template.Must(template.ParseFiles("templates/country.html"))
var whoStats []who.RawData

func homepage(w http.ResponseWriter, r *http.Request) {
	log.Println(r.RemoteAddr, r.Method, r.URL)
	tmpl.Execute(w, nil)
}

func countryPage(w http.ResponseWriter, r *http.Request) {
	loc, ok := r.URL.Query()["loc"]
	if !ok {
		http.NotFound(w, r)
		return
	}

	countryTemplate.Execute(w, loc[0])
}

func country(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Data []who.RawData `json:"data"`
	}
	loc, ok := r.URL.Query()["loc"]
	if ok {
		payload.Data = who.Country(whoStats, loc[0])
	}
	json.NewEncoder(w).Encode(payload)
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
	http.HandleFunc("/country.json", country)
	http.HandleFunc("/country.html", countryPage)
	http.HandleFunc("/favicon.ico", favicon)

	http.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("listening on localhost:8080")
	log.Fatal(http.ListenAndServe(":8080", nil))

}
