package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/alecthomas/kong"
	"github.com/octarect/turf/cmd/turf/output"
	"github.com/octarect/turf/model"
	"github.com/octarect/turf/scrape/jra"
	"github.com/octarect/turf/scrape/jraen"
	"github.com/octarect/turf/turf"
)

type CLI struct {
	Output    string `short:"o" help:"Output format: json, custom-columns=HEADER:.path,..." name:"output"`
	UserAgent string `help:"User-Agent header sent with each HTTP request." name:"user-agent"`

	Fixtures       FixturesCmd       `cmd:"" help:"List fixtures (race meetings)."`
	Races          RacesCmd          `cmd:"" help:"List races for a fixture."`
	Result         ResultCmd         `cmd:"" help:"Get the result of a race."`
	LatestFixtures LatestFixturesCmd `cmd:"latest-fixtures" help:"List this week's fixtures."`
	LatestRaces    LatestRacesCmd    `cmd:"latest-races" help:"List races for a this-week fixture."`
	Plan           PlanCmd           `cmd:"" help:"Get the race plan (entry list) for a this-week race."`
}

type FixturesCmd struct {
	Month  string `help:"Month to query (YYYY-MM)." xor:"period"`
	Date   string `help:"Date to query (YYYY-MM-DD)." xor:"period"`
	Course string `help:"Filter by course (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)."`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

type RacesCmd struct {
	Date   string `help:"Date to query (YYYY-MM-DD)." required:""`
	Course string `help:"Course name (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)." required:""`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

type ResultCmd struct {
	Date   string `help:"Date to query (YYYY-MM-DD)." required:""`
	Course string `help:"Course name (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)." required:""`
	Race   int    `help:"Race number." required:""`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

type LatestFixturesCmd struct {
	Course string `help:"Filter by course (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)."`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

type LatestRacesCmd struct {
	Date   string `help:"Date to query (YYYY-MM-DD)." required:""`
	Course string `help:"Course name (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)." required:""`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

type PlanCmd struct {
	Date   string `help:"Date to query (YYYY-MM-DD)." required:""`
	Course string `help:"Course name (tokyo, kyoto, hanshin, nakayama, chukyo, kokura, sapporo, hakodate, fukushima, niigata)." required:""`
	Race   int    `help:"Race number." required:""`

	Output    string `kong:"-"`
	UserAgent string `kong:"-"`
}

var fixturesDefaultColumns = []output.Column{
	{Header: "DATE", Path: []string{"date"}},
	{Header: "COURSE", Path: []string{"course"}},
	{Header: "SEASON", Path: []string{"season"}},
	{Header: "DAY", Path: []string{"day"}},
}

var racesDefaultColumns = []output.Column{
	{Header: "NUM", Path: []string{"num"}},
	{Header: "NAME", Path: []string{"nameEN"}},
	{Header: "GRADE", Path: []string{"grade"}},
	{Header: "SURFACE", Path: []string{"surface"}},
	{Header: "DIST", Path: []string{"distance"}},
	{Header: "RUNNERS", Path: []string{"runners"}},
}

var resultDefaultColumns = []output.Column{
	{Header: "FP", Path: []string{"entries", "[*]", "finish", "position"}},
	{Header: "BK", Path: []string{"entries", "[*]", "bracket"}},
	{Header: "NUM", Path: []string{"entries", "[*]", "num"}},
	{Header: "HORSE", Path: []string{"entries", "[*]", "horse", "nameEN"}},
	{Header: "SEX", Path: []string{"entries", "[*]", "horse", "sex"}},
	{Header: "AGE", Path: []string{"entries", "[*]", "horse", "age"}},
	{Header: "WEIGHT", Path: []string{"entries", "[*]", "weight"}},
	{Header: "JOCKEY", Path: []string{"entries", "[*]", "jockey", "nameEN"}},
	{Header: "TIME", Path: []string{"entries", "[*]", "finishTime"}, Format: formatFinishTime},
	{Header: "MARGIN", Path: []string{"entries", "[*]", "margin"}, Format: formatMargin},
	{Header: "LAST3F", Path: []string{"entries", "[*]", "last3F"}},
	{Header: "CORNER", Path: []string{"entries", "[*]", "cornerPositions"}, Format: formatCornerPositions},
	{Header: "TRAINER", Path: []string{"entries", "[*]", "trainer", "nameEN"}},
}

var planDefaultColumns = []output.Column{
	{Header: "BK", Path: []string{"entries", "[*]", "bracket"}},
	{Header: "NUM", Path: []string{"entries", "[*]", "num"}},
	{Header: "HORSE", Path: []string{"entries", "[*]", "horse", "nameEN"}},
	{Header: "SEX", Path: []string{"entries", "[*]", "horse", "sex"}},
	{Header: "AGE", Path: []string{"entries", "[*]", "horse", "age"}},
	{Header: "HWT", Path: []string{"entries", "[*]", "horse", "weight"}},
	{Header: "WEIGHT", Path: []string{"entries", "[*]", "weight"}},
	{Header: "JOCKEY", Path: []string{"entries", "[*]", "jockey", "nameEN"}},
	{Header: "TRAINER", Path: []string{"entries", "[*]", "trainer", "nameEN"}},
}

func (cmd *FixturesCmd) Run() error {
	query, err := cmd.query()
	if err != nil {
		return err
	}

	client := newServices(cmd.UserAgent)
	fixtures, err := client.fixtureSvc.ListFixtures(context.Background(), query)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, fixtures, cmd.Output, fixturesDefaultColumns)
}

func (cmd *RacesCmd) Run() error {
	query, err := cmd.query()
	if err != nil {
		return err
	}

	client := newServices(cmd.UserAgent)
	raceCards, err := client.raceCardSvc.ListRaceCards(context.Background(), query)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, raceCards, cmd.Output, racesDefaultColumns)
}

func (cmd *ResultCmd) Run() error {
	query, err := cmd.query()
	if err != nil {
		return err
	}

	client := newServices(cmd.UserAgent)
	result, err := client.raceResultSvc.GetRaceResult(context.Background(), query)
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, result, cmd.Output, resultDefaultColumns)
}

func (cmd *LatestFixturesCmd) Run() error {
	course, err := parseCourse(cmd.Course)
	if err != nil {
		return err
	}

	client := newServices(cmd.UserAgent)
	fixtures, err := client.latestFixtureSvc.ListLatestFixtures(context.Background(), turf.ListLatestFixturesQuery{Course: course})
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, fixtures, cmd.Output, fixturesDefaultColumns)
}

func (cmd *LatestRacesCmd) Run() error {
	date, err := parseDate(cmd.Date)
	if err != nil {
		return err
	}

	course, err := parseCourse(cmd.Course)
	if err != nil {
		return err
	}

	client := newServices(cmd.UserAgent)
	raceCards, err := client.latestRaceCardSvc.ListRaceCards(context.Background(), turf.ListRaceCardsQuery{Date: date, Course: course})
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, raceCards, cmd.Output, racesDefaultColumns)
}

func (cmd *PlanCmd) Run() error {
	date, err := parseDate(cmd.Date)
	if err != nil {
		return err
	}

	course, err := parseCourse(cmd.Course)
	if err != nil {
		return err
	}

	if cmd.Race < 1 {
		return fmt.Errorf("race number must be greater than 0")
	}

	client := newServices(cmd.UserAgent)
	plan, err := client.racePlanSvc.GetRacePlan(context.Background(), turf.GetRacePlanQuery{Date: date, Course: course, RaceNo: cmd.Race})
	if err != nil {
		return err
	}

	return output.Write(os.Stdout, plan, cmd.Output, planDefaultColumns)
}

func (cmd *FixturesCmd) query() (turf.ListFixturesQuery, error) {
	query := turf.ListFixturesQuery{}

	if cmd.Month != "" {
		month, err := parseMonth(cmd.Month)
		if err != nil {
			return query, err
		}
		query.Month = month
	}

	if cmd.Date != "" {
		date, err := parseDate(cmd.Date)
		if err != nil {
			return query, err
		}
		query.Date = date
	}

	course, err := parseCourse(cmd.Course)
	if err != nil {
		return query, err
	}
	query.Course = course

	return query, nil
}

func (cmd *RacesCmd) query() (turf.ListRaceCardsQuery, error) {
	date, err := parseDate(cmd.Date)
	if err != nil {
		return turf.ListRaceCardsQuery{}, err
	}

	course, err := parseCourse(cmd.Course)
	if err != nil {
		return turf.ListRaceCardsQuery{}, err
	}

	return turf.ListRaceCardsQuery{
		Date:   date,
		Course: course,
	}, nil
}

func (cmd *ResultCmd) query() (turf.GetRaceResultQuery, error) {
	date, err := parseDate(cmd.Date)
	if err != nil {
		return turf.GetRaceResultQuery{}, err
	}

	course, err := parseCourse(cmd.Course)
	if err != nil {
		return turf.GetRaceResultQuery{}, err
	}

	if cmd.Race < 1 {
		return turf.GetRaceResultQuery{}, fmt.Errorf("race number must be greater than 0")
	}

	return turf.GetRaceResultQuery{
		Date:   date,
		Course: course,
		RaceNo: cmd.Race,
	}, nil
}

type services struct {
	fixtureSvc        turf.FixtureService
	raceCardSvc       turf.RaceCardService
	raceResultSvc     turf.RaceResultService
	latestFixtureSvc  turf.LatestFixtureService
	latestRaceCardSvc turf.RaceCardService
	racePlanSvc       turf.RacePlanService
}

func newServices(userAgent string) services {
	httpClient := turf.NewClient(nil)
	if userAgent != "" {
		httpClient.SetUserAgent(userAgent)
	}
	jraClient := jra.NewJRAClient(httpClient)

	jraenHTTPClient := turf.NewClient(nil)
	if userAgent != "" {
		jraenHTTPClient.SetUserAgent(userAgent)
	}
	jraenBaseURL, _ := url.Parse("https://jra.jp/JRAEN/AP/")
	jraenHTTPClient.SetBaseURL(jraenBaseURL)
	jraenClient := jraen.NewJRAENClient(jraenHTTPClient)

	fixtureSvc := turf.NewFixtureService(jraClient)
	raceCardSvc := turf.NewRaceCardService(fixtureSvc, jraClient, jraenClient)
	raceResultSvc := turf.NewRaceResultService(raceCardSvc, jraClient, jraenClient)
	latestFixtureSvc := turf.NewLatestFixtureService(jraClient)
	latestRaceCardSvc := turf.NewLatestRaceCardService(latestFixtureSvc, jraClient)
	racePlanSvc := turf.NewRacePlanService(latestFixtureSvc, jraClient, jraClient, jraenClient)

	return services{
		fixtureSvc:        fixtureSvc,
		raceCardSvc:       raceCardSvc,
		raceResultSvc:     raceResultSvc,
		latestFixtureSvc:  latestFixtureSvc,
		latestRaceCardSvc: latestRaceCardSvc,
		racePlanSvc:       racePlanSvc,
	}
}

func parseMonth(value string) (time.Time, error) {
	month, err := time.Parse("2006-01", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid month %q: expected YYYY-MM", value)
	}
	earliestDate := turf.EarliestFixtureDate()
	if month.Before(earliestDate) {
		return time.Time{}, fmt.Errorf("month must be on or after %s", earliestDate.Format("2006-01"))
	}
	return month, nil
}

func parseDate(value string) (time.Time, error) {
	date, err := time.Parse("2006-01-02", value)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid date %q: expected YYYY-MM-DD", value)
	}
	earliestDate := turf.EarliestFixtureDate()
	if date.Before(earliestDate) {
		return time.Time{}, fmt.Errorf("date must be on or after %s", earliestDate.Format("2006-01-02"))
	}
	return date, nil
}

func parseCourse(value string) (model.Course, error) {
	if value == "" {
		return model.CourseUnknown, nil
	}

	course, err := model.CourseByENName(value)
	if err != nil {
		return model.CourseUnknown, err
	}
	return course, nil
}

func main() {
	cli := CLI{}
	ctx := kong.Parse(&cli,
		kong.Name("turf"),
		kong.Description("CLI for fetching JRA horse racing fixtures, races, and results."),
		kong.Vars{"version": turf.Version},
		kong.UsageOnError(),
	)

	cli.Fixtures.Output = cli.Output
	cli.Races.Output = cli.Output
	cli.Result.Output = cli.Output
	cli.LatestFixtures.Output = cli.Output
	cli.LatestRaces.Output = cli.Output
	cli.Plan.Output = cli.Output
	cli.Fixtures.UserAgent = cli.UserAgent
	cli.Races.UserAgent = cli.UserAgent
	cli.Result.UserAgent = cli.UserAgent
	cli.LatestFixtures.UserAgent = cli.UserAgent
	cli.LatestRaces.UserAgent = cli.UserAgent
	cli.Plan.UserAgent = cli.UserAgent

	if err := ctx.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func formatMargin(s string) string {
	var m struct {
		Kind   string  `json:"kind"`
		Length float64 `json:"length"`
	}
	if err := json.Unmarshal([]byte(s), &m); err != nil {
		return s
	}
	switch m.Kind {
	case "length":
		if m.Length == 0 {
			return ""
		}
		return strconv.FormatFloat(m.Length, 'f', -1, 64)
	case "dead_heat":
		return "Dead Heat"
	case "nose":
		return "Nose"
	case "head":
		return "Head"
	case "neck":
		return "Neck"
	case "distance":
		return "Distance"
	default:
		return ""
	}
}

func formatFinishTime(s string) string {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil || f == 0 {
		return ""
	}
	min := int(f) / 60
	sec := f - float64(min*60)
	return fmt.Sprintf("%d:%04.1f", min, sec)
}

func formatCornerPositions(s string) string {
	var positions []struct {
		Position int `json:"position"`
	}
	if err := json.Unmarshal([]byte(s), &positions); err != nil {
		return s
	}
	parts := make([]string, len(positions))
	for i, p := range positions {
		parts[i] = strconv.Itoa(p.Position)
	}
	return strings.Join(parts, "-")
}
