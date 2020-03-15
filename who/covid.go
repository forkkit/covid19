package who

import (
	"encoding/csv"
	"fmt"
	"io"
	"strconv"
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

func atoi(s string) (int, error) {
	if s == "" {
		return 0, nil
	}
	return strconv.Atoi(s)

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
		date, err := time.Parse("2006-01-02", rec[0])
		if err != nil {
			return out, err
		}
		newCases, err := atoi(rec[2])
		if err != nil {
			return out, err
		}
		newDeaths, err := atoi(rec[3])
		if err != nil {
			return out, err
		}
		totalCases, err := atoi(rec[4])
		if err != nil {
			return out, err
		}
		totalDeaths, err := atoi(rec[5])
		if err != nil {
			return out, err
		}

		out = append(out, rawData{
			date,
			rec[1],
			newCases,
			newDeaths,
			totalCases,
			totalDeaths,
		})
	}
	return out, err
}
