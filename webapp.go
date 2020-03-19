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
var statsUpdated string

func homepage(w http.ResponseWriter, r *http.Request) {
	data := struct {
		Url        string
		UpdateTime string
	}{
		who.WhoSourceURL,
		statsUpdated,
	}
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	tmpl.Execute(w, data)
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
		ws, updated, err := who.DownloadCSV(who.WhoSourceURL)
		if err != nil {
			log.Println(err)
		} else {
			lock.Lock()
			whoStats = ws
			statsUpdated = updated
			lock.Unlock()
		}
		time.Sleep(time.Hour)
	}
}

func checkHTTPMethod(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func main() {
	go updater()
	h := setupHanders()
	log.Println("listening on http://localhost:8080/")
	s := &http.Server{
		Handler:      handlers.LoggingHandler(os.Stdout, h),
		ReadTimeout:  1 * time.Second,
		WriteTimeout: 10 * time.Second,
		Addr:         ":8080",
	}
	log.Fatalln(s.ListenAndServe())
}

func setupHanders() http.Handler {
	r := http.NewServeMux()
	r.HandleFunc("/", homepage)
	r.HandleFunc("/stats.json", stats)
	r.HandleFunc("/country.json", country)
	r.HandleFunc("/country.html", countryPage)
	r.HandleFunc("/favicon.ico", favicon)
	r.Handle("/static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))
	return checkHTTPMethod(r)
}
