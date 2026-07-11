package jra

import (
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/octarect/turf/turf"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/transform"
)

func setup(t *testing.T) (*JRAClient, *http.ServeMux) {
	t.Helper()

	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}
	tc := turf.NewClient(httpClient)

	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	t.Cleanup(server.Close)

	u, _ := url.Parse(server.URL + "/")
	tc.SetBaseURL(u)

	return NewJRAClient(tc), mux
}

func writeShiftJIS(t *testing.T, w http.ResponseWriter, body string) {
	t.Helper()

	writer := transform.NewWriter(w, japanese.ShiftJIS.NewEncoder())
	defer writer.Close()

	if _, err := io.WriteString(writer, body); err != nil {
		t.Fatalf("failed to write Shift-JIS response: %v", err)
	}
}

func fixturePageHTML(blocks ...string) string {
	return `<ul class="past_result_line mt20">` + strings.Join(blocks, "") + `</ul>`
}

func fixtureDateBlockHTML(monthDay string, items ...string) string {
	return `<div class="past_result_line_unit"><div class="head"><h3 class="sub_header">` + monthDay + `</h3></div>` + strings.Join(items, "") + `</div>`
}

func fixtureItemHTML(cname, fixture string) string {
	return `<div class="cell kaisai"><div><div><a onclick="return doAction('x', '` + cname + `');">` + fixture + `</a></div></div></div>`
}

func racePageHTML(rows ...string) string {
	return `<table id="race_list"><tbody>` + strings.Join(rows, "") + `</tbody></table>`
}

func raceRowHTML(raceNo int, name, subName, gradeIcon, distance, courseType string, runners int, cname string) string {
	gradeSpan := ""
	if gradeIcon != "" {
		gradeSpan = `<span class="grade_icon"><img src="` + gradeIcon + `"/></span>`
	}
	subDiv := ""
	if subName != "" {
		subDiv = `<div>` + subName + `</div>`
	}
	return `<tr>` +
		`<th class="race_num"><a href="/JRADB/accessS.html?CNAME=` + cname + `"><img alt="` + strconv.Itoa(raceNo) + `レース"/></a></th>` +
		`<td class="race_name"><div><div>` + name + gradeSpan + `</div>` + subDiv + `</div></td>` +
		`<td class="dist">` + distance + `</td>` +
		`<td class="course">` + courseType + `</td>` +
		`<td class="num">` + strconv.Itoa(runners) + `</td>` +
		`</tr>`
}

func latestFixturesPageHTML(panels ...string) string {
	return `<div id="main">` + strings.Join(panels, "") + `</div>`
}

func latestFixturesPanelHTML(monthDay string, items ...string) string {
	return `<div class="panel"><h3 class="sub_header">` + monthDay + `</h3>` +
		strings.Join(items, "") + `</div>`
}

func latestFixturesItemHTML(cname, fixture string) string {
	return `<div class="waku"><a onclick="return doAction('x', '` + cname + `');">` + fixture + `</a></div>`
}

type racePlanEntryHTML struct {
	Bracket      int
	HorseNo      int
	HorseName    string
	HorseCNAME   string
	HorseID      string
	SexAge       string
	HorseWeight  string
	Weight       string
	JockeyName   string
	JockeyCNAME  string
	JockeyID     string
	TrainerName  string
	TrainerCNAME string
	TrainerID    string
}

func racePlanPageHTML(postTime string, entries ...racePlanEntryHTML) string {
	doc := []string{
		`<div class="cell rule"></div>`,
		`<div class="cell weight">馬齢</div>`,
		`<div class="cell time"><strong>` + postTime + `</strong></div>`,
		`<div id="syutsuba"><table><tbody>`,
	}
	for _, e := range entries {
		doc = append(doc, racePlanEntryRowHTML(e))
	}
	doc = append(doc, `</tbody></table></div>`)
	return strings.Join(doc, "")
}

func racePlanEntryRowHTML(e racePlanEntryHTML) string {
	return `<tr>` +
		`<td class="waku"><img src="/img/waku` + strconv.Itoa(e.Bracket) + `.png"/></td>` +
		`<td class="num">` + strconv.Itoa(e.HorseNo) + `</td>` +
		`<td class="horse">` +
		`<div class="name"><a href="/JRADB/accessS.html?CNAME=` + e.HorseCNAME + `">` + e.HorseName + `</a></div>` +
		`<div class="weight">` + e.HorseWeight + `</div>` +
		`<p class="trainer"><a onclick="return doAction('x', '` + e.TrainerCNAME + `');">` + e.TrainerName + `</a></p>` +
		`</td>` +
		`<td class="jockey">` +
		`<p class="age">` + e.SexAge + `</p>` +
		`<p class="weight">` + e.Weight + `</p>` +
		`<p class="jockey"><a onclick="return doAction('x', '` + e.JockeyCNAME + `');">` + e.JockeyName + `</a></p>` +
		`</td>` +
		`</tr>`
}

type resultOpts struct {
	Weather          string
	GoingClass       string
	GoingText        string
	FemaleOnly       string
	WeightRule       string
	PostTime         string
	LapTimes         string
	CornerFormations []cornerFormationHTML
	Entries          []entryHTML
}

type cornerFormationHTML struct {
	Corner    int
	Formation string
	SecondLap bool
}

type entryHTML struct {
	FinishPosition  string
	Bracket         int
	HorseNo         int
	HorseName       string
	HorseCNAME      string
	SexAge          string
	Weight          string
	JockeyName      string
	JockeyAllowance int
	JockeyCNAME     string
	FinishTime      string
	Margin          string
	Corners         []cornerPositionHTML
	Last3F          string
	HorseWeight     string
	TrainerName     string
	TrainerCNAME    string
	WinFavorite     int
}

type cornerPositionHTML struct {
	Corner    int
	Position  int
	SecondLap bool
}

func raceResultPageHTML(opts resultOpts) string {
	doc := []string{
		`<div class="cell baba"><ul>`,
		`<li class="weather"><span class="txt">`, opts.Weather, `</span></li>`,
		`<li class="`, opts.GoingClass, `"><span class="txt">`, opts.GoingText, `</span></li>`,
		`</ul></div>`,
		`<div class="cell rule">`, opts.FemaleOnly, `</div>`,
		`<div class="cell weight">`, opts.WeightRule, `</div>`,
		`<div class="cell time"><strong>`, opts.PostTime, `</strong></div>`,
		`<div class="result_time_data"><table><tbody><tr><td>`, opts.LapTimes, `</td></tr></tbody></table></div>`,
		`<div class="result_corner_place"><table><tbody>`,
	}

	for _, cf := range opts.CornerFormations {
		label := strconv.Itoa(cf.Corner) + "コーナー"
		if cf.SecondLap {
			label += "(2周目)"
		}
		doc = append(doc,
			`<tr><th>`,
			label,
			`</th><td>`,
			cf.Formation,
			`</td></tr>`,
		)
	}

	doc = append(doc, `</tbody></table></div>`, `<div id="race_result"><div><table><tbody>`)

	for _, e := range opts.Entries {
		doc = append(doc, entryRowHTML(e))
	}

	doc = append(doc, `</tbody></table></div></div>`)

	return strings.Join(doc, "")
}

func entryRowHTML(e entryHTML) string {
	doc := []string{
		`<tr>`,
		`<td class="place">`, e.FinishPosition, `</td>`,
		`<td class="waku"><img src="/img/waku`, strconv.Itoa(e.Bracket), `.png"/></td>`,
		`<td class="num">`, strconv.Itoa(e.HorseNo), `</td>`,
		`<td class="horse"><a href="/JRADB/accessS.html?CNAME=`, e.HorseCNAME, `">`, e.HorseName, `</a></td>`,
		`<td class="age">`, e.SexAge, `</td>`,
		`<td class="weight">`, e.Weight, `</td>`,
	}

	if e.JockeyCNAME == "" {
		doc = append(doc, `<td class="jockey">`, e.JockeyName, `</td>`)
	} else if e.JockeyAllowance > 0 {
		doc = append(doc,
			`<td class="jockey"><span title="`, strconv.Itoa(e.JockeyAllowance), `kg減量" class="mark jockey">▲</span>`,
			`<a onclick="return doAction('x', '`, e.JockeyCNAME, `');">`, e.JockeyName, `</a></td>`,
		)
	} else {
		doc = append(doc,
			`<td class="jockey"><a onclick="return doAction('x', '`, e.JockeyCNAME, `');">`, e.JockeyName, `</a></td>`,
		)
	}

	doc = append(doc,
		`<td class="time">`, e.FinishTime, `</td>`,
		`<td class="margin">`, e.Margin, `</td>`,
		`<td class="corner"><div><ul>`,
	)

	for _, cp := range e.Corners {
		title := strconv.Itoa(cp.Corner) + "コーナー通過順位"
		if cp.SecondLap {
			title += "(2周目)"
		}
		doc = append(doc,
			`<li title="`,
			title,
			`">`,
			strconv.Itoa(cp.Position),
			`</li>`,
		)
	}

	doc = append(doc, []string{
		`</ul></div></td>`,
		`<td class="f_time">`, e.Last3F, `</td>`,
		`<td class="h_weight">`, e.HorseWeight, `</td>`,
		`<td class="trainer" onclick="return doAction('x', '`, e.TrainerCNAME, `');">`, e.TrainerName, `</td>`,
		`<td class="pop">`, strconv.Itoa(e.WinFavorite), `</td>`,
		`</tr>`,
	}...)

	return strings.Join(doc, "")
}
