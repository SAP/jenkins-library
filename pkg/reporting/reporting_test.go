package reporting

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestToHTML(t *testing.T) {
	t.Run("empty table", func(t *testing.T) {
		report := ScanReport{
			Title:      "Report Test Title",
			Subheaders: []Subheader{{Description: "sub 1", Details: "1"}, {Description: "sub 2", Details: "2"}},
			Overview: []OverviewRow{
				{"overview 1", "1", Green},
				{"overview 2", "2", Green},
			},
			FurtherInfo: "this is further information",
			ReportTime:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			DetailTable: ScanDetailTable{
				Headers:       []string{"column 1", "column 2"},
				Rows:          []ScanRow{},
				WithCounter:   true,
				CounterHeader: "Entry #",
				NoRowsMessage: "no rows available",
			},
		}
		expectedSub := `<span>
		sub 1: 1<br />
		sub 2: 2<br />
		</span>
	</h2>`
		expectedOverview := `<h3>
		overview 1: 1<br />
		overview 2: 2<br />
		</h3>`

		res, err := report.ToHTML()
		result := string(res)
		assert.NoError(t, err)
		assert.Contains(t, result, "<h1>Report Test Title</h1>")
		assert.Contains(t, result, expectedSub)
		assert.Contains(t, result, expectedOverview)
		assert.Contains(t, result, `<span>this is further information</span>`)
		assert.Contains(t, result, `<th>Entry #</th>`)
		assert.Contains(t, result, `<th>column 1</th>`)
		assert.Contains(t, result, `<th>column 2</th>`)
		assert.Contains(t, result, "Snapshot taken: Jan 01, 2021 - 00:00:00 UTC")
		assert.Contains(t, result, `<td colspan="3">no rows available</td>`)
	})

	t.Run("table with content", func(t *testing.T) {
		report := ScanReport{
			Title:      "Report Test Title",
			ReportTime: time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
			DetailTable: ScanDetailTable{
				Headers: []string{"column 1", "column 2"},
				Rows: []ScanRow{
					{Columns: []ScanCell{{Content: "c1 r1"}, {Content: "c2 r1"}}},
					{Columns: []ScanCell{{Content: "c1 r2"}, {Content: "c2 r2"}}},
					{Columns: []ScanCell{{Content: "c1 r3", Style: Green}, {Content: "c2 r3", Style: Black}}},
				},
				CounterHeader: "Entry #",
				WithCounter:   true,
			},
		}
		res, err := report.ToHTML()
		result := string(res)
		assert.NoError(t, err)
		assert.Contains(t, result, `<th>Entry #</th>`)
		assert.Contains(t, result, `<td>1</td>`)
		assert.Contains(t, result, `<td>c1 r1</td>`)
		assert.Contains(t, result, `<td>c2 r1</td>`)
		assert.Contains(t, result, `<td>2</td>`)
		assert.Contains(t, result, `<td>c1 r2</td>`)
		assert.Contains(t, result, `<td>c2 r2</td>`)
		assert.Contains(t, result, `<td>3</td>`)
		assert.Contains(t, result, `<td class="green-cell">c1 r3</td>`)
		assert.Contains(t, result, `<td class="black-cell">c2 r3</td>`)
	})
}

func TestToMarkdown(t *testing.T) {
	report := ScanReport{
		Title:      "Report Test Title",
		Subheaders: []Subheader{{Description: "sub 1", Details: "1"}, {Description: "sub 2", Details: "2"}},
		Overview: []OverviewRow{
			{"overview 1", "1", Green},
			{"overview 2", "2", Green},
		},
		FurtherInfo: "this is further information",
		ReportTime:  time.Date(2021, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	res, err := report.ToMarkdown()
	result := string(res)
	assert.NoError(t, err)
	assert.Contains(t, result, `## Report Test Title`)
	assert.Contains(t, result, `<td><b>sub 1:</b></td><td>1</td>`)
	assert.Contains(t, result, `<td><b>sub 2:</b></td><td>2</td>`)
	assert.Contains(t, result, `<tr><td>overview 1:</td><td>1</td></tr>`)
	assert.Contains(t, result, `<tr><td>overview 2:</td><td>2</td></tr>`)
	assert.Contains(t, result, `this is further information`)
	assert.Contains(t, result, "Snapshot taken: _Jan 01, 2021 - 00:00:00 UTC_")
}

func TestTableColumnCount(t *testing.T) {
	t.Run("table without counter", func(t *testing.T) {
		details := ScanDetailTable{
			Headers:     []string{"column 1", "column 1"},
			WithCounter: false,
		}
		assert.Equal(t, 2, tableColumnCount(details))
	})
	t.Run("table with counter", func(t *testing.T) {
		details := ScanDetailTable{
			Headers:     []string{"column 1", "column 1"},
			WithCounter: true,
		}
		assert.Equal(t, 3, tableColumnCount(details))
	})
}
