package jraen

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strconv"

	"github.com/octarect/turf/model"
	"github.com/octarect/turf/turf"
	"github.com/octarect/xtract"
)

// JRAENClient is a scraper client for the JRA English website (https://jra.jp/JRAEN/).
// It implements RaceResultTranslator, providing English names for races, horses, jockeys, and trainers.
type JRAENClient struct {
	client *turf.Client
}

var (
	_ = turf.RaceResultTranslator(&JRAENClient{})
)

func NewJRAENClient(client *turf.Client) *JRAENClient {
	return &JRAENClient{
		client: client,
	}
}

type raceResultPage struct {
	RaceName string `xpath:"replace(//body/table[1]//tr[2], '\\((G[1-3]|L|J \\. G[1-3])\\)$', '')"`
	Entries  []struct {
		HorseID     string `xpath:"//td[@class='raceHorse'][1]/a/@horseno"`
		HorseName   string `xpath:"replace(//td[@class='raceHorse'][1]/a/@bamei, '\\([A-Z]+\\)', '')"`
		JockeyName  string `xpath:"//td[@class='raceHorse'][4]/text()[1]"`
		TrainerName string `xpath:"//td[@class='raceHorse'][4]/text()[2]"`
	} `xpath:"//body/table[@class='running'][last()]//tr[position() >= 2]"`
}

func (c *JRAENClient) GetRaceResultTranslation(ctx context.Context, raceCard *model.RaceCard) (*model.RaceResultTranslation, error) {
	reqURL := "kaisai/running"
	form := &url.Values{}
	form.Add("raceYmd", raceCard.Fixture.Date.Format("20060102"))
	form.Add("raceJoCd", fmt.Sprintf("%02d", raceCard.Fixture.Course))
	form.Add("raceYear", strconv.Itoa(raceCard.Fixture.Year))
	form.Add("raceKai", fmt.Sprintf("%02d", raceCard.Fixture.Season))
	form.Add("raceHi", fmt.Sprintf("%02d", raceCard.Fixture.Day))
	form.Add("raceNo", fmt.Sprintf("%02d", raceCard.Num))

	req, err := c.client.NewFormRequest(reqURL, form)
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

	entries := make(map[model.HorseID]model.RaceResultTranslationEntry, len(page.Entries))
	for _, e := range page.Entries {
		id := model.HorseID(e.HorseID)
		entries[id] = model.RaceResultTranslationEntry{
			HorseName:   e.HorseName,
			JockeyName:  e.JockeyName,
			TrainerName: e.TrainerName,
		}
	}

	return &model.RaceResultTranslation{
		RaceName: page.RaceName,
		Entries:  entries,
	}, nil
}
