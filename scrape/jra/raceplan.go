package jra

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/octarect/turf/model"
	"github.com/octarect/xtract"
)

type racePlanPage struct {
	FemaleOnly femaleOnly `xpath:"//div[@class='cell rule']"`
	WeightRule weightRule `xpath:"//div[@class='cell weight']"`

	PostTime struct {
		Hour   int `xpath:"replace(//text(), '(.+)時(.+)分', '$1')"`
		Minute int `xpath:"replace(//text(), '(.+)時(.+)分', '$2')"`
	} `xpath:"//div[@class='cell time']/strong"`

	Entries []struct {
		Bracket   int `xpath:"replace(.//td[@class='waku']/img/@src, '.*([0-9]+)\\.png', '$1')"`
		HorseNo   int `xpath:".//td[@class='num']/text()"`
		HorseName struct {
			Text  string `xpath:"./text()"`
			CNAME string `xpath:"replace(./@href, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
			ID    string `xpath:"replace(./@href, '.*pw[0-9]{2}[a-z]{3}[0-9]{2}([0-9]{10})/[0-9A-Z]{2}.*', '$1')"`
		} `xpath:".//td[@class='horse']//div[@class='name']//a"`
		HorseSexAge struct {
			Sex horseSex `xpath:"replace(normalize-space(text()), '([^0-9]+).*', '$1')"`
			Age int      `xpath:"replace(normalize-space(text()), '[^0-9]+([0-9]+)/.+', '$1')"`
		} `xpath:".//td[@class='jockey']/p[@class='age']"`
		HorseWeight horseWeight `xpath:".//td[@class='horse']//div[contains(@class,'weight')]"`
		Weight      float64     `xpath:"normalize-space(.//td[@class='jockey']/p[@class='weight']/text())"`
		JockeyName  struct {
			Text      jockeyNameText      `xpath:"//self::*"`
			Allowance jockeyNameAllowance `xpath:"./span[@class='mark jockey']/@title"`
			CNAME     string              `xpath:"replace(.//a/@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
			ID        string              `xpath:"replace(.//a/@onclick, '.*pw[0-9]{2}[a-z]{3}[0-9]+([0-9]{5})/[0-9A-Za-z]+.*', '$1')"`
		} `xpath:".//td[@class='jockey']/p[@class='jockey']"`
		TrainerName struct {
			Text  string `xpath:"./text()"`
			CNAME string `xpath:"replace(./@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
			ID    string `xpath:"replace(./@onclick, '.*pw[0-9]{2}[a-z]{3}[0-9]+([0-9]{5})/[0-9A-Za-z]+.*', '$1')"`
		} `xpath:".//td[@class='horse']//p[@class='trainer']//a"`
	} `xpath:"//div[@id='syutsuba']/table/tbody/tr"`
}

func (c *JRAClient) GetRacePlan(ctx context.Context, raceCard *model.RaceCard) (*model.RacePlan, error) {
	reqURL := fmt.Sprintf("%s?CNAME=%s", jraAccessDPath, url.QueryEscape(raceCard.CNAME))

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

	var page racePlanPage
	if err := xtract.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	// Entries
	entries := make([]model.RacePlanEntry, 0, len(page.Entries))
	for _, e := range page.Entries {
		e0 := model.RacePlanEntry{
			Bracket: e.Bracket,
			Num:     e.HorseNo,
			Weight:  e.Weight,
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

	// Post Time
	postTime := time.Date(raceCard.Fixture.Date.Year(), raceCard.Fixture.Date.Month(), raceCard.Fixture.Date.Day(), page.PostTime.Hour, page.PostTime.Minute, 0, 0, timeJST)

	return &model.RacePlan{
		RaceCard:   raceCard,
		FemaleOnly: model.FemaleOnly(page.FemaleOnly),
		WeightRule: model.WeightRule(page.WeightRule),
		PostTime:   postTime,
		Entries:    entries,
	}, nil
}
