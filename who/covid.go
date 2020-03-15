package who

import (
	"encoding/csv"
	"fmt"
	"io"
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

func getField(rec []string, headers []string, item string) (string, error) {
	i := 0
	h := ""
	for i, h = range headers {
		if h == item {
			break
		}
	}
	if h != item {
		return "", fmt.Errorf("No such column %q in CSV steam", item)
	}
	if len(rec) <= i {
		return "", fmt.Errorf("Too few fields in record")
	}
	return rec[i], nil
}

func getIntField(rec []string, headers []string, item string) (int, error) {
	s, err := getField(rec, headers, item)
	if err != nil {
		return 0, err
	}
	if s == "" {
		return 0, nil
	}
	i, err := strconv.Atoi(s)
	if err != nil {
		return 0, fmt.Errorf("Error parsing CSV stream: bad format for %s, could not parse %q as an integer", item, s)
	}
	return i, nil
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
		tf, err := getField(rec, headers, "date")
		if err != nil {
			return out, err
		}
		date, err := time.Parse("2006-01-02", tf)
		if err != nil {
			return out, fmt.Errorf("Could not parse Date in CSV stream, %w", err)
		}
		location, err := getField(rec, headers, "location")
		if err != nil {
			return out, err
		}

		newCases, err := getIntField(rec, headers, "new_cases")
		if err != nil {
			return out, err
		}
		newDeaths, err := getIntField(rec, headers, "new_deaths")
		if err != nil {
			return out, err
		}
		totalCases, err := getIntField(rec, headers, "total_cases")
		if err != nil {
			return out, err
		}
		totalDeaths, err := getIntField(rec, headers, "total_deaths")
		if err != nil {
			return out, err
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

func DownloadCSV(url string) (o []RawData, err error) {
	r, err := http.Get(url)
	if err != nil {
		return o, err
	}
	if r.StatusCode != http.StatusOK {
		return o, fmt.Errorf("Status %q when downloading CSV file from %q", r.Status, url)
	}
	defer r.Body.Close()
	return ReadFullData(r.Body)
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
