package who

import (
	"encoding/csv"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

const sourceURL = "https://covid.ourworldindata.org/data/full_data.csv"
const nCols = 6

type rawData struct {
	date        time.Time
	location    string
	newCases    int
	newDeaths   int
	totalCases  int
	totalDeaths int
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
	return strings.TrimSpace(rec[i]), nil
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

func ReadFullData(in io.Reader) ([]rawData, error) {
	var out []rawData
	r := csv.NewReader(in)
	headers, err := r.Read()
	if err != nil {
		return out, err
	}
	if len(headers) != nCols {
		return out, fmt.Errorf("Want %d fields in csv stream, got %d", nCols, len(headers))
	}

	for {
		rec, err := r.Read()
		if len(rec) == 0 {
			break
		}
		if len(rec) != nCols {
			return out, fmt.Errorf("Want %d fields in csv stream, got %d", nCols, len(rec))
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
			return out, fmt.Errorf("Could not parse date in CSV stream, %w", err)
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

		out = append(out, rawData{
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

func DownloadCSV(url string) (o []rawData, err error) {
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
