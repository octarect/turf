package jra

import (
	"context"
	"io"
	"net/url"
	"time"

	"github.com/octarect/turf/model"
	"github.com/octarect/xtract"
)

type latestFixturesPage struct {
	Dates []struct {
		MonthDay struct {
			Month time.Month `xpath:"replace(//text(), '\\s*(1[0-2]|[1-9])月.*', '$1')"`
			Day   int        `xpath:"replace(//text(), '.*月([1-9]|[12][0-9]|3[0-1])日.*', '$1')"`
		} `xpath:"//h3[@class='sub_header']"`

		Items []struct {
			Fixture struct {
				Course string `xpath:"replace(//text(), '^[1-9]回([^0-9]+).*$', '$1')"`
				Season int    `xpath:"replace(//text(), '^([1-9])回.*', '$1')"`
				Day    int    `xpath:"replace(//text(), '^.*[^0-9](1[0-9]|[1-9])日$', '$1')"`
			} `xpath:"//self::*"`
			CNAME string `xpath:"replace(//self::*/@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
		} `xpath:"//div[@class='waku']/a"`
	} `xpath:"//div[@id='main']/div[contains(@class, 'panel')]"`
}

const rootLatestCNAME = "pw01dli00/F3"

const jraAccessDPath = "/JRADB/accessD.html"

func (c *JRAClient) ListLatestFixtures(ctx context.Context) ([]*model.Fixture, error) {
	form := &url.Values{}
	form.Add("cname", rootLatestCNAME)

	req, err := c.client.NewFormRequest(jraAccessDPath, form)
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

	var page latestFixturesPage
	if err := xtract.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	var fs []*model.Fixture
	for _, d := range page.Dates {
		date0 := time.Date(time.Now().Year(), d.MonthDay.Month, d.MonthDay.Day, 0, 0, 0, 0, timeJST)
		for _, item := range d.Items {
			f, err := model.NewFixture(date0, date0.Year(), item.Fixture.Season, item.Fixture.Day, item.Fixture.Course, string(item.CNAME))
			if err != nil {
				return nil, err
			}
			fs = append(fs, f)
		}
	}

	return fs, nil
}
