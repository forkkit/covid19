package who

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
2020-02-25,Afghanistan,1,2,3,4
`
	o, err := ReadFullData(strings.NewReader(testData))
	require.NoError(t, err)
	require.Equal(t, 1, len(o))
	assert.Equal(t, dt(2020, 2, 25), o[0].date)

	assert.Equal(t, 1, o[0].newCases)
	assert.Equal(t, 2, o[0].newDeaths)
	assert.Equal(t, 3, o[0].totalCases)
	assert.Equal(t, 4, o[0].totalDeaths)
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
	assert.Equal(t, dt(2020, 3, 9), o[0].date)
	assert.Equal(t, 67, o[0].newCases)
	assert.Equal(t, 0, o[0].newDeaths)
	assert.Equal(t, 277, o[0].totalCases)
	assert.Equal(t, 2, o[0].totalDeaths)
	assert.Equal(t, "United Kingdom", o[0].location)

	assert.Equal(t, dt(2020, 3, 13), o[4].date)
	assert.Equal(t, 134, o[4].newCases)
	assert.Equal(t, 2, o[4].newDeaths)
	assert.Equal(t, 594, o[4].totalCases)
	assert.Equal(t, 8, o[4].totalDeaths)
	assert.Equal(t, "United Kingdom", o[4].location)
}

func Test_readFullError(t *testing.T) {
	testData := `date,xxx,new_cases,new_deaths,total_cases,total_deaths
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
