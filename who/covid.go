package who

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
)

const WhoSourceURL = "https://covid.ourworldindata.org/data/full_data.csv"
const nCols = 6

type RawData struct {
	Date        time.Time
	Location    string
	NewCases    int
	NewDeaths   int
	TotalCases  int
	TotalDeaths int
}

type fieldReader struct {
	headers []string
	rec     []string
	err     error
}

func (f *fieldReader) getField(item string) string {
	if f.err != nil {
		return ""
	}
	i := 0
	h := ""
	for i, h = range f.headers {
		if h == item {
			break
		}
	}
	if h != item {
		f.err = fmt.Errorf("No such column %q in CSV steam", item)
		return ""
	}
	if len(f.rec) <= i {
		f.err = fmt.Errorf("Too few fields in record")
	}
	return f.rec[i]
}

func (f *fieldReader) getIntField(item string) int {
	s := f.getField(item)
	if f.err != nil {
		return 0
	}
	if s == "" {
		return 0
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		f.err = fmt.Errorf("Error parsing CSV stream: bad format for %s, could not parse %q as an integer", item, s)
		return 0
	}
	return i
}

func ReadFullData(in io.Reader) ([]RawData, error) {
	var out []RawData
	r := csv.NewReader(in)
	r.FieldsPerRecord = nCols
	r.TrimLeadingSpace = true
	headers, err := r.Read()
	if err != nil {
		return out, err
	}
	for {
		rec, err := r.Read()
		if len(rec) == 0 {
			break
		}
		if err != nil {
			return out, err
		}

		f := fieldReader{headers: headers, rec: rec}
		tf := f.getField("date")
		date, err := time.Parse("2006-01-02", tf)
		if err != nil {
			return out, fmt.Errorf("Could not parse Date in CSV stream, %w", err)
		}
		location := f.getField("location")
		newCases := f.getIntField("new_cases")
		newDeaths := f.getIntField("new_deaths")
		totalCases := f.getIntField("total_cases")
		totalDeaths := f.getIntField("total_deaths")
		if f.err != nil {
			return out, f.err
		}

		out = append(out, RawData{
			date,
			location,
			newCases,
			newDeaths,
			totalCases,
			totalDeaths,
		})
	}
	return out, err
}

func DownloadCSV(url string) (o []RawData, updated string, err error) {
	r, err := http.Get(url)
	if err != nil {
		return o, "", err
	}
	if r.StatusCode != http.StatusOK {
		return o, "", fmt.Errorf("Status %q when downloading CSV file from %q", r.Status, url)
	}
	defer r.Body.Close()
	log.Println("GET", url, r.Status)
	updated = r.Header.Get("date")
	o, err = ReadFullData(r.Body)
	log.Println(len(o), "lines of data read for", len(Latest(o)), "countries.")
	return o, updated, err
}

func LastDate(r []RawData) time.Time {
	var last time.Time
	for _, x := range r {
		if x.Date.After(last) {
			last = x.Date
		}
	}
	return last
}

func Latest(r []RawData) []RawData {
	seen := make(map[string]bool, len(r))
	o := make([]RawData, 0, len(r))
	for i := len(r) - 1; i >= 0; i-- {
		if seen[r[i].Location] {
			continue
		}
		seen[r[i].Location] = true
		o = append(o, r[i])
	}
	return o
}

func Country(r []RawData, country string) []RawData {
	o := make([]RawData, 0, len(r))
	for _, p := range r {
		if p.Location == country {
			o = append(o, p)
		}
	}
	return o
}
