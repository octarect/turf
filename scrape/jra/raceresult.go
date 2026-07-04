package jra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/octarect/turf/model"
	"github.com/octarect/xtract"
	"golang.org/x/text/width"
)

type raceResultPage struct {
	Weather    weather    `xpath:"normalize-space(//div[@class='cell baba']/ul/li[@class='weather']//span[@class='txt']/text())"`
	Going      going      `xpath:"normalize-space(//div[@class='cell baba']/ul/li[@class='turf' or @class='durt']//span[@class='txt']/text())"`
	FemaleOnly femaleOnly `xpath:"//div[@class='cell rule']"`
	WeightRule weightRule `xpath:"//div[@class='cell weight']"`

	PostTime struct {
		Hour   int `xpath:"replace(//text(), '(.+)時(.+)分', '$1')"`
		Minute int `xpath:"replace(//text(), '(.+)時(.+)分', '$2')"`
	} `xpath:"//div[@class='cell time']/strong"`

	LapTimes lapTimes `xpath:"normalize-space(//div[contains(@class, 'result_time_data')]/table/tbody/tr[1]/td/text())"`

	CornerFormations []struct {
		Corner    int    `xpath:"replace(//th, '(\\d)コーナー.*', '$1')"`
		Formation string `xpath:"normalize-space(//td)"`
	} `xpath:"//div[contains(@class, 'result_corner_place')]/table/tbody/tr"`

	Entries []struct {
		FinishPosition finishPosition `xpath:"//td[@class='place']/text()"`
		Bracket        int            `xpath:"replace(//td[@class='waku']/img/@src, '.*([0-9]+)\\.png', '$1')"`
		HorseNo        int            `xpath:"//td[@class='num']/text()"`
		HorseName      struct {
			Text  string `xpath:"//text()"`
			CNAME string `xpath:"replace(//@href, '.*(pw[0-9]{2}[a-z]{3}[0-9]{2}[0-9]{10}/[0-9A-Z]{2}).*', '$1')"`
			ID    string `xpath:"replace(//@href, '.*pw[0-9]{2}[a-z]{3}[0-9]{2}([0-9]{10})/[0-9A-Z]{2}.*', '$1')"`
		} `xpath:"//td[@class='horse']//a"`
		HorseSexAge struct {
			Sex horseSex `xpath:"replace(//text(), '([^0-9]+).*', '$1')"`
			Age int      `xpath:"replace(//text(), '[^0-9]+([0-9]+)', '$1')"`
		} `xpath:"//td[@class='age']"`
		Weight     float64 `xpath:"//td[@class='weight']/text()"`
		JockeyName struct {
			Text      jockeyNameText      `xpath:"//self::*"`
			Allowance jockeyNameAllowance `xpath:"/span[@class='mark jockey']/@title"`
			CNAME     string              `xpath:"replace(//a/@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
			ID        string              `xpath:"replace(//a/@onclick, '.*pw[0-9]{2}[a-z]{3}[0-9]+([0-9]{5})/[0-9A-Za-z]+.*', '$1')"`
		} `xpath:"//td[@class='jockey']"`
		FinishTime      finishTime `xpath:"//td[@class='time']"`
		Margin          margin     `xpath:"normalize-space(//td[@class='margin']/text())"`
		CornerPositions []struct {
			Corner   int            `xpath:"replace(//@title, '([0-9])コーナー通過順位.*', '$1')"`
			Position cornerPosition `xpath:"normalize-space(//text())"`
		} `xpath:"//td[@class='corner']/div/ul/li"`
		Last3F      optionalFloat64 `xpath:"//td[@class='f_time']/text()"`
		HorseWeight horseWeight     `xpath:"//td[@class='h_weight']"`
		TrainerName struct {
			Text  string `xpath:"//text()"`
			CNAME string `xpath:"replace(//@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
			ID    string `xpath:"replace(//@onclick, '.*pw[0-9]{2}[a-z]{3}[0-9]+([0-9]{5})/[0-9A-Za-z]+.*', '$1')"`
		} `xpath:"//td[@class='trainer']"`
		WinFavorite optionalInt `xpath:"//td[@class='pop']/text()"`
	} `xpath:"//div[@id='race_result']/div/table/tbody/tr"`
}

type weather model.Weather

func (w *weather) UnmarshalXPath(text []byte) error {
	switch string(text) {
	case "晴":
		*w = weather(model.WeatherFine)
	case "曇":
		*w = weather(model.WeatherCloudy)
	case "小雨":
		*w = weather(model.WeatherDrizzle)
	case "雨":
		*w = weather(model.WeatherRainy)
	case "小雪":
		*w = weather(model.WeatherLightSnow)
	case "雪":
		*w = weather(model.WeatherSnow)
	default:
		return fmt.Errorf("unknown weather type found. text=%s", string(text))
	}
	return nil
}

type going model.Going

func (g *going) UnmarshalXPath(text []byte) error {
	switch string(text) {
	case "良":
		*g = going(model.GoingTurfGoodToFirm)
	case "稍重":
		*g = going(model.GoingTurfGood)
	case "重":
		*g = going(model.GoingTurfYielding)
	case "不良":
		*g = going(model.GoingTurfSoft)
	default:
		return fmt.Errorf("unknown going type found. text=%s", string(text))
	}
	return nil
}

// OfSurface adjusts the going value based on the race surface.
// The Going enum interleaves turf and dirt values, so dirt Going = turf Going + 1
// for the same condition level (e.g. GoingTurfGoodToFirm+1 = GoingDirtStandard).
func (g *going) OfSurface(surface model.Surface) model.Going {
	if surface == model.SurfaceDirt {
		return model.Going(*g + 1)
	}
	return model.Going(*g)
}

type femaleOnly model.FemaleOnly

func (fo *femaleOnly) UnmarshalXPath(text []byte) error {
	*fo = femaleOnly(strings.Contains(string(text), "牝"))
	return nil
}

type weightRule model.WeightRule

func (wg *weightRule) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))
	switch s {
	case "馬齢":
		*wg = weightRule(model.WeightRuleAge)
	case "別定", "定量":
		*wg = weightRule(model.WeightRuleSpecial)
	case "ハンデ":
		*wg = weightRule(model.WeightRuleHandicap)
	default:
		return fmt.Errorf("unknown weight rule found. text=%s", s)
	}
	return nil
}

type lapTimes []float64

func (ts *lapTimes) UnmarshalXPath(lapTimesStr []byte) error {
	ss := strings.SplitSeq(string(lapTimesStr), "-")
	for s := range ss {
		s0 := strings.TrimSpace(s)
		lapTime, err := strconv.ParseFloat(s0, 64)
		if err != nil {
			// Jump races use a mileage-based format (e.g. "1マイル 1分49秒3 4F 55.5 - 3F 41.8")
			// that cannot be parsed as split lap times. Treat as no lap data.
			*ts = lapTimes{}
			return nil
		}
		*ts = append(*ts, lapTime)
	}
	return nil
}

type horseWeight struct {
	Weight int
	Diff   int
	Other  string
}

type horseSex model.HorseSex

func (hs *horseSex) UnmarshalXPath(text []byte) error {
	var hs0 model.HorseSex

	switch string(text) {
	case "牡":
		hs0 = model.HorseSexMale
	case "牝":
		hs0 = model.HorseSexFemale
	case "セ", "せん":
		hs0 = model.HorseSexGelding
	default:
		return fmt.Errorf("unknown horse sex found. sexStr=%s", string(text))
	}

	*hs = horseSex(hs0)

	return nil
}

type jockeyNameText string

var regexpJockeyName = regexp.MustCompile(`^[▲△☆★◇]?(.*)$`)

// UnmarshalXPath strips leading apprentice marks from a jockey name.
// JRA prefixes apprentice jockey names with symbols indicating their weight allowance:
//
//	▲ = 3kg, △ = 2kg, ☆ = 1kg (first year), ★ = 1kg, ◇ = special designation
func (jnt *jockeyNameText) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))

	m := regexpJockeyName.FindStringSubmatch(s)
	if m == nil || len(m) < 2 {
		return fmt.Errorf("invalid jockey name text. text=%s", s)
	}

	*jnt = jockeyNameText(m[1])

	return nil
}

type jockeyNameAllowance int

var regexpJockeyAllowance = regexp.MustCompile(`^([0-9])kg減量$`)

// UnmarshalXPath parses the jockey weight allowance from a title attribute string like "3kg減量".
// An empty string means no allowance (senior jockey).
func (jna *jockeyNameAllowance) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))

	if s == "" {
		return nil
	}

	m := regexpJockeyAllowance.FindStringSubmatch(s)
	if m == nil || len(m) < 2 {
		return fmt.Errorf("invalid jockey allowance text. text=%s", s)
	}

	n, err := strconv.Atoi(m[1])
	if err != nil {
		return fmt.Errorf("invalid jockey allowance text. text=%s", s)
	}

	*jna = jockeyNameAllowance(n)

	return nil
}

var regexpFinishTime = regexp.MustCompile(`([0-9]+):([0-5][0-9])\.([0-9])`)

type finishPosition struct {
	Position int
	Status   model.FinishStatus
}

// UnmarshalXPath parses a finish position from JRA's result page.
// Normally this is a numeric placement (1, 2, 3, ...).
// Non-finishing entries are represented by Japanese status strings:
//
//	除外 = Withdrawn (excluded before race), 取消 = NonStarter (scratched),
//	中止 = PulledUp (did not finish), 失格 = Disqualified
func (fp *finishPosition) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))
	switch s {
	case "除外":
		fp.Status = model.FinishStatusWithdrawn
	case "取消":
		fp.Status = model.FinishStatusNonStarter
	case "中止":
		fp.Status = model.FinishStatusPulledUp
	case "失格":
		fp.Status = model.FinishStatusDisqualified
	default:
		n, err := strconv.Atoi(s)
		if err != nil {
			return fmt.Errorf("invalid finish position. text=%s", s)
		}
		fp.Position = n
	}
	return nil
}

type cornerPosition int

func (cp *cornerPosition) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid corner position. text=%s", s)
	}
	*cp = cornerPosition(n)
	return nil
}

type optionalFloat64 float64

func (f *optionalFloat64) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return nil
	}
	v, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return fmt.Errorf("invalid format of float. error=%w", err)
	}
	*f = optionalFloat64(v)
	return nil
}

type optionalInt int

func (i *optionalInt) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))
	if s == "" {
		return nil
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid format of int. error=%w", err)
	}
	*i = optionalInt(n)
	return nil
}

type finishTime float64

// UnmarshalXPath parses a finish time string like "1:23.4" into total seconds as a float64.
// The format is M:SS.T where M=minutes, SS=seconds, T=tenths of a second.
// An empty string (e.g. for non-finishers) is treated as zero.
func (ft *finishTime) UnmarshalXPath(timeStr []byte) error {
	s := string(timeStr)
	if s == "" {
		return nil
	}
	m := regexpFinishTime.FindStringSubmatch(s)
	if len(m) < 4 {
		return fmt.Errorf("invalid finish time text. timeStr=%s", s)
	}
	minute, err := strconv.Atoi(m[1])
	if err != nil {
		return fmt.Errorf("invalid finish time minute found. timeStr=%s", s)
	}
	second, err := strconv.Atoi(m[2])
	if err != nil {
		return fmt.Errorf("invalid finish time second found. timeStr=%s", s)
	}
	tenth, err := strconv.Atoi(m[3])
	if err != nil {
		return fmt.Errorf("invalid finish time tenth found. timeStr=%s", s)
	}
	*ft = finishTime(float64(minute)*60 + float64(second) + float64(tenth)*0.1)
	return nil
}

type margin model.Margin

// parseFraction parses a margin length string into a float64.
// JRA uses full-width characters and mixed fractions (e.g. "１ １/２" for 1.5 lengths).
// The string is first normalized to half-width, then split into whole and fractional parts.
func parseFraction(s string) (float64, error) {
	whole, frac, _ := strings.Cut(width.Fold.String(s), " ")
	if frac == "" {
		frac = whole
		whole = "0"
	}

	w, err := strconv.ParseFloat(whole, 64)
	if err != nil {
		return 0, err
	}

	num, den, ok := strings.Cut(frac, "/")
	if !ok {
		return strconv.ParseFloat(frac, 64)
	}

	n, err := strconv.ParseFloat(num, 64)
	if err != nil {
		return 0, err
	}
	d, err := strconv.ParseFloat(den, 64)
	if err != nil {
		return 0, err
	}

	return w + n/d, nil
}

var regexpDemotion = regexp.MustCompile(`^\([0-9]+位降着\)$`)

// UnmarshalXPath parses a Japanese margin string into a Margin value.
// Named margins map to MarginKind constants:
//
//	同着=DeadHeat, ハナ=Nose, アタマ=Head, クビ=Neck, 大差=Distance
//
// Numeric margins (including mixed fractions like "1 1/2") are parsed via parseFraction.
// An empty string indicates the race winner (no margin).
// Demotion annotations like "(2位降着)" are ignored (margin is left as zero value).
func (m *margin) UnmarshalXPath(marginStr []byte) error {
	s := string(marginStr)

	kind := model.MarginKindLength
	length := 0.0

	switch s {
	case "":
		// 1st
		return nil
	case "同着":
		kind = model.MarginKindDeadHeat
	case "ハナ":
		kind = model.MarginKindNose
	case "アタマ":
		kind = model.MarginKindHead
	case "クビ":
		kind = model.MarginKindNeck
	case "大差":
		kind = model.MarginKindDistance
	default:
		if regexpDemotion.MatchString(s) {
			return nil
		}
		var err error
		length, err = parseFraction(s)
		if err != nil {
			return err
		}
	}

	*m = margin(model.Margin{
		Kind:   kind,
		Length: length,
	})

	return nil
}

var regexpHorseWeight = regexp.MustCompile(`(?P<weight>[0-9]+)(?:\((?:(?P<diff>[-+]?[0-9]+)|(?P<other>[^)]+))\))?`)

// UnmarshalXPath parses a horse weight string like "480(+2)", "480(計不)", or "計不".
// The format is: weight(diff) where diff is a signed integer or a special annotation.
// "計不" (計量不能) means the weight could not be measured.
func (hw *horseWeight) UnmarshalXPath(weightStr []byte) error {
	s := strings.TrimSpace(string(weightStr))
	if s == "" {
		// Weight == 0 means the horse weight is unavailable (not measured or not recorded).
		return nil
	}
	m := regexpHorseWeight.FindStringSubmatch(s)
	if len(m) < 2 {
		if strings.HasPrefix(s, "計不") {
			// "計不" (計量不能) means the weight could not be measured.
			// Zero value (Weight=0, Diff=0) indicates unavailable.
			return nil
		}
		return fmt.Errorf("invalid horse weight found. weightStr=%s", s)
	}

	var err error
	hw.Weight, err = strconv.Atoi(m[1])
	if err != nil {
		return fmt.Errorf("invalid horse weight found. weightStr=%s", s)
	}

	if m[2] != "" {
		hw.Diff, err = strconv.Atoi(m[2])
		if err != nil {
			return fmt.Errorf("invalid horse weight diff found. weightStr=%s", s)
		}
	}

	hw.Other = m[3]

	return nil
}

func (c *JRAClient) GetRaceResult(ctx context.Context, raceCard *model.RaceCard) (*model.RaceResult, error) {
	reqURL := fmt.Sprintf("%s?CNAME=%s", jraAccessSPath, url.QueryEscape(raceCard.CNAME))

	req, err := c.client.NewRequest(http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(ctx, req)
	if err != nil {
		return nil, err
	}

	body, err := io.ReadAll(resp)
	if err != nil {
		return nil, err
	}

	var page raceResultPage
	if err := xtract.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	// Post Time
	postTime := time.Date(raceCard.Fixture.Date.Year(), raceCard.Fixture.Date.Month(), raceCard.Fixture.Date.Day(), page.PostTime.Hour, page.PostTime.Minute, 0, 0, timeJST)

	// Corner Formations
	cornerFormations := make([]model.CornerFormation, 0, len(page.CornerFormations))
	for _, cf := range page.CornerFormations {
		cf0 := model.CornerFormation{
			Corner:    cf.Corner,
			Formation: cf.Formation,
		}
		cornerFormations = append(cornerFormations, cf0)
	}

	// Entries
	entries := make([]model.Entry, 0, len(page.Entries))
	for _, e := range page.Entries {
		// Corner Positions
		cornerPositions := make([]model.CornerPosition, 0, len(e.CornerPositions))
		for _, cp := range e.CornerPositions {
			cornerPositions = append(cornerPositions, model.CornerPosition{Corner: cp.Corner, Position: int(cp.Position)})
		}

		e0 := model.Entry{
			Finish: model.Finish{
				Position: e.FinishPosition.Position,
				Status:   e.FinishPosition.Status,
			},
			Bracket:         e.Bracket,
			Num:             e.HorseNo,
			Weight:          e.Weight,
			FinishTime:      float64(e.FinishTime),
			Margin:          model.Margin(e.Margin),
			CornerPositions: cornerPositions,
			Last3F:          float64(e.Last3F),
			WinFavorite:     int(e.WinFavorite),
			Horse: model.EntryHorse{
				ID:         model.HorseID(e.HorseName.ID),
				Name:       e.HorseName.Text,
				Sex:        model.HorseSex(e.HorseSexAge.Sex),
				Age:        e.HorseSexAge.Age,
				Weight:     e.HorseWeight.Weight,
				WeightDiff: e.HorseWeight.Diff,
				CNAME:      e.HorseName.CNAME,
			},
			Jockey: model.EntryJockey{
				ID:        model.JockeyID(e.JockeyName.ID),
				Name:      string(e.JockeyName.Text),
				Allowance: int(e.JockeyName.Allowance),
				CNAME:     e.JockeyName.CNAME,
			},
			Trainer: model.EntryTrainer{
				ID:    model.TrainerID(e.TrainerName.ID),
				Name:  e.TrainerName.Text,
				CNAME: e.TrainerName.CNAME,
			},
		}
		entries = append(entries, e0)
	}

	return &model.RaceResult{
		RaceCard:         raceCard,
		Going:            page.Going.OfSurface(raceCard.Surface),
		FemaleOnly:       model.FemaleOnly(page.FemaleOnly),
		WeightRule:       model.WeightRule(page.WeightRule),
		Weather:          model.Weather(page.Weather),
		PostTime:         postTime,
		Entries:          entries,
		LapTimes:         page.LapTimes,
		CornerFormations: cornerFormations,
	}, nil
}
