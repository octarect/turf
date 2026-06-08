package jra

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"time"

	"github.com/octarect/turf/model"
	"github.com/octarect/xtract"
)

type fixturesPage struct {
	Dates []struct {
		MonthDay struct {
			Month time.Month `xpath:"replace(//text(), '\\s*(1[0-2]|[1-9])月.*', '$1')"`
			Day   int        `xpath:"replace(//text(), '.*月([1-9]|[12][0-9]|3[0-1])日.*', '$1')"`
		} `xpath:"//div[@class='head']/h3[@class='sub_header']"`
		Items []struct {
			Fixture struct {
				Course string `xpath:"replace(//text(), '^[1-9]回([^0-9]+).*$', '$1')"`
				Season int    `xpath:"replace(//text(), '^([1-9])回.*', '$1')"`
				Day    int    `xpath:"replace(//text(), '^.*[^0-9](1[0-9]|[1-9])日$', '$1')"`
			} `xpath:"//self::*"`
			CNAME string `xpath:"replace(//self::*/@onclick, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
		} `xpath:"//div[@class='cell kaisai']/div/div/a"`
	} `xpath:"//ul[@class='past_result_line mt20']//div[@class='past_result_line_unit']"`
}

func (c *JRAClient) ListFixtures(ctx context.Context, date time.Time) ([]*model.Fixture, error) {
	cname, err := c.generateCNAME(ctx, date)
	if err != nil {
		return nil, err
	}

	form := &url.Values{}
	form.Add("cname", cname)

	req, err := c.client.NewFormRequest(jraAccessRPath, form)
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

	var page fixturesPage
	if err := xtract.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	var fs []*model.Fixture
	for _, d := range page.Dates {
		date0 := time.Date(date.Year(), d.MonthDay.Month, d.MonthDay.Day, 0, 0, 0, 0, timeJST)
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

// generateCNAME builds a CNAME string for the given month.
// The CNAME format is: pw01skl{monthKind}{YYYYMM}/{suffix}
//   - monthKind: "00" for the current month, "10" for past months
//   - suffix: a 2-character hex code obtained from the JRA website via listCNAMESuffixMap
func (c *JRAClient) generateCNAME(ctx context.Context, t time.Time) (string, error) {
	suffixes, err := c.listCNAMESuffixMap(ctx)
	if err != nil {
		return "", err
	}
	suffix, ok := suffixes[t.Format("0601")]
	if !ok {
		return "", fmt.Errorf("no data found in the specified month. month=%s", t.Format("2006/01"))
	}

	monthKind := "10"
	if time.Now().Format("0601") == t.Format("0601") {
		monthKind = "00"
	}

	return fmt.Sprintf("pw01skl%s%s/%s", monthKind, t.Format("200601"), suffix), nil
}

// CNAMESuffixMap maps YYMM keys (e.g. "2506") to their 2-character CNAME suffixes (e.g. "B3").
// The suffix is embedded in JavaScript on the JRA website and is required to construct valid CNAMEs.
// YYMM -> Suffix
type CNAMESuffixMap map[string]string

// rootCNAME is the seed CNAME used to fetch the suffix map from the JRA website.
// Posting this CNAME to accessS.html returns a page containing the JavaScript
// objParam entries from which the suffix map is parsed.
const rootCNAME = "pw01skl00999999/B3"

var regexpObjParam = regexp.MustCompile(`objParam\[\"(\d{4})\"\]=\"([A-Z0-9]{2})\";`)

// listCNAMESuffixMap fetches the JRA root page using rootCNAME, then parses
// the embedded JavaScript to extract the CNAME suffix map.
// The suffix is not algorithmically derivable; it must be read from the page's
// JavaScript variable: objParam["YYMM"] = "XX".
func (c *JRAClient) listCNAMESuffixMap(ctx context.Context) (CNAMESuffixMap, error) {
	form := &url.Values{}
	form.Add("cname", rootCNAME)

	req, err := c.client.NewFormRequest(jraAccessSPath, form)
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

	matches := regexpObjParam.FindAllStringSubmatch(string(body), -1)
	if matches == nil {
		return nil, fmt.Errorf("objParam not found in %s", jraAccessSPath)
	}

	result := make(CNAMESuffixMap, len(matches))
	for _, m := range matches {
		result[m[1]] = m[2]
	}

	return result, nil
}
