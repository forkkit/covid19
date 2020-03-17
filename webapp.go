package main

import (
	"encoding/json"
	"github.com/amnonbc/covid19/who"
	"github.com/gorilla/handlers"
	"html/template"
	"log"
	"net/http"
	"os"
	"sync"
	"time"
)

var tmpl = template.Must(template.ParseFiles("templates/top.html"))
var countryTemplate = template.Must(template.ParseFiles("templates/country.html"))

var lock sync.RWMutex
var whoStats []who.RawData

func homepage(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
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
		lock.RLock()
		ws := whoStats
		lock.RUnlock()
		payload.Data = who.Country(ws, loc[0])
	}
	json.NewEncoder(w).Encode(payload)
}

func stats(w http.ResponseWriter, r *http.Request) {
	var payload struct {
		Data []who.RawData `json:"data"`
	}
	lock.RLock()
	ws := whoStats
	lock.RUnlock()
	payload.Data = who.Latest(ws)
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
			lock.Lock()
			whoStats = ws
			lock.Unlock()
		}
		time.Sleep(time.Hour)
	}
}

func main() {
	go updater()
	r := http.NewServeMux()
	r.HandleFunc("/", homepage)
	r.HandleFunc("/stats.json", stats)
	r.HandleFunc("/country.json", country)
	r.HandleFunc("/country.html", countryPage)
	r.HandleFunc("/favicon.ico", favicon)
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	log.Println("listening on http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", handlers.LoggingHandler(os.Stdout, r)))

}
