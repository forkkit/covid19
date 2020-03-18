package who

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"net/http"
	"net/http/httptest"
	"os"
	"sort"
	"strings"
	"testing"
	"time"
)

func Test_readFullDataEmpty(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths`
	_, err := ReadFullData(strings.NewReader(testData))
	assert.NoError(t, err)
}

func dt(y, m, d int) time.Time {
	return time.Date(y, time.Month(m), d, 0, 0, 0, 0, time.UTC)
}

func Test_readFullDataOne(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		2020-02-25,Afghanistan,1,2,3,4`
	o, err := ReadFullData(strings.NewReader(testData))
	require.NoError(t, err)
	require.Equal(t, 1, len(o))
	assert.Equal(t, dt(2020, 2, 25), o[0].Date)
	assert.Equal(t, 1, o[0].NewCases)
	assert.Equal(t, 2, o[0].NewDeaths)
	assert.Equal(t, 3, o[0].TotalCases)
	assert.Equal(t, 4, o[0].TotalDeaths)
}

func Test_readFullData(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		2020-03-09,United Kingdom,67,0,277,2
		2020-03-10,United Kingdom,46,1,323,3
		2020-03-11,United Kingdom,50,3,373,6
		2020-03-12,United Kingdom,87,0,460,6
		2020-03-13,United Kingdom,134,2,594,8`
	o, err := ReadFullData(strings.NewReader(testData))
	require.NoError(t, err)
	require.Equal(t, 5, len(o))
	assert.Equal(t, dt(2020, 3, 9), o[0].Date)
	assert.Equal(t, 67, o[0].NewCases)
	assert.Equal(t, 0, o[0].NewDeaths)
	assert.Equal(t, 277, o[0].TotalCases)
	assert.Equal(t, 2, o[0].TotalDeaths)
	assert.Equal(t, "United Kingdom", o[0].Location)

	assert.Equal(t, dt(2020, 3, 13), o[4].Date)
	assert.Equal(t, 134, o[4].NewCases)
	assert.Equal(t, 2, o[4].NewDeaths)
	assert.Equal(t, 594, o[4].TotalCases)
	assert.Equal(t, 8, o[4].TotalDeaths)
	assert.Equal(t, "United Kingdom", o[4].Location)
}

func Test_readFullError(t *testing.T) {
	testData := `Date,xxx,new_cases,new_deaths,total_cases,total_deaths
		2020-02-25,Afghanistan,1,2,3,4`
	_, err := ReadFullData(strings.NewReader(testData))
	assert.Error(t, err)
}

func Test_readFullErrorBadDate(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		abc,Afghanistan,1,2,3,4`
	_, err := ReadFullData(strings.NewReader(testData))
	assert.Error(t, err)
}

func Test_readFullNotCSV(t *testing.T) {
	testData := `"`
	_, err := ReadFullData(strings.NewReader(testData))
	assert.Error(t, err)
}

func TestDownloadCSV(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `date,location,new_cases,new_deaths,total_cases,total_deaths
			2020-02-25,Afghanistan,1,2,3,4`)
	}))
	defer ts.Close()

	o, _, err := DownloadCSV(ts.URL)
	require.NoError(t, err)
	require.Equal(t, 1, len(o))
	assert.Equal(t, dt(2020, 2, 25), o[0].Date)
	assert.Equal(t, 1, o[0].NewCases)
	assert.Equal(t, 2, o[0].NewDeaths)
	assert.Equal(t, 3, o[0].TotalCases)
	assert.Equal(t, 4, o[0].TotalDeaths)
}

func Test_lastDate(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		2020-03-09,United Kingdom,67,0,277,2
		2020-03-10,United Kingdom,46,1,323,3
		2020-03-11,United Kingdom,50,3,373,6
		2020-03-12,United Kingdom,87,0,460,6
		2020-03-13,United Kingdom,134,2,594,8`
	o, err := ReadFullData(strings.NewReader(testData))
	assert.NoError(t, err)
	assert.Equal(t, dt(2020, 3, 13), LastDate(o))
}

func TestLatest(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		2020-03-09,United Kingdom,67,0,277,2
		2020-03-10,United Kingdom,46,1,323,3
		2020-03-11,United Kingdom,50,3,373,6
		2020-03-12,United Kingdom,87,0,460,6
		2020-03-13,United Kingdom,134,2,594,8
		2020-03-11,United States,224,6,696,25
		2020-03-12,United States,291,4,987,29
		2020-03-13,United States,277,7,1264,36
		2020-03-14,United States,414,5,1678,41`
	o, err := ReadFullData(strings.NewReader(testData))
	assert.NoError(t, err)
	ol := Latest(o)
	sort.Slice(ol, func(i, j int) bool {
		return ol[i].Location < ol[j].Location
	})
	assert.Equal(t, 2, len(ol))
	assert.Equal(t, dt(2020, 3, 13), ol[0].Date)
	assert.Equal(t, "United Kingdom", ol[0].Location)
	assert.Equal(t, dt(2020, 3, 14), ol[1].Date)
	assert.Equal(t, "United States", ol[1].Location)
}

func TestCountry(t *testing.T) {
	testData := `date,location,new_cases,new_deaths,total_cases,total_deaths
		2020-03-09,United Kingdom,67,0,277,2
		2020-03-10,United Kingdom,46,1,323,3
		2020-03-11,United Kingdom,50,3,373,6
		2020-03-12,United Kingdom,87,0,460,6
		2020-03-13,United Kingdom,134,2,594,8
		2020-03-11,United States,224,6,696,25
		2020-03-12,United States,291,4,987,29
		2020-03-13,United States,277,7,1264,36
		2020-03-14,United States,414,5,1678,41`
	o, err := ReadFullData(strings.NewReader(testData))
	assert.NoError(t, err)
	us := Country(o, "United States")
	require.Equal(t, 4, len(us))
}

func TestCountryCount(t *testing.T) {
	f, err := os.Open("testdata/full_data.csv")
	require.NoError(t, err)
	defer f.Close()
	o, err := ReadFullData(f)
	require.NoError(t, err)
	require.Equal(t, 137, len(Latest(o)))
}
