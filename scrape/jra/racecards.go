package jra

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"regexp"
	"strconv"
	"strings"

	"github.com/octarect/turf/model"
	"github.com/octarect/xtract"
)

type raceCardListPage struct {
	RaceCards []raceCard `xpath:"//table[@id='race_list']/tbody/tr"`
}

type raceCard struct {
	Num          int      `xpath:"replace(//th[@class='race_num']/a/img/@alt, '([0-9]+)レース', '$1')"`
	Name         raceName `xpath:"//td[@class='race_name']/div/div[1]"`
	SubName      raceName `xpath:"//td[@class='race_name']/div/div[2]"`
	GradeIconSrc string   `xpath:"//td[@class='race_name']//span[@class='grade_icon']/img/@src"`
	Distance     distance `xpath:"//td[@class='dist']/text()"`
	Surface      surface  `xpath:"//td[@class='course']/text()"`
	Runners      int      `xpath:"normalize-space(//td[@class='num']/text())"`
	CNAME        string   `xpath:"replace(//th[@class='race_num']/a/@href, '.*(pw[0-9A-Za-z/]+).*', '$1')"`
}

var regexpAgeGroup = regexp.MustCompile(`([234])歳(以上)?`)

// parseAgeGroup parses a Japanese age restriction string (e.g. "3歳以上") into an AgeGroup.
// It matches patterns like "2歳", "3歳", "3歳以上", "4歳以上".
func parseAgeGroup(s string) (model.AgeGroup, bool) {
	m := regexpAgeGroup.FindStringSubmatch(s)
	if m == nil {
		return 0, false
	}
	age, over := m[1], m[2] != ""
	switch {
	case age == "2":
		return model.AgeGroup2, true
	case age == "3" && over:
		return model.AgeGroup3Plus, true
	case age == "3":
		return model.AgeGroup3, true
	case age == "4" && over:
		return model.AgeGroup4Plus, true
	}
	return 0, false
}

// toModel converts a scraped raceCard into a model.RaceCard.
// Grade is determined by inspecting the subName and grade icon image src:
// open races are identified first, then graded races (G1-G3, JG1-JG3) via icon filename,
// followed by win-class races (3勝/2勝/1勝クラス) and newcomer races.
// A race is considered "special" (i.e. has a proper name) when it has a subName and is not a newcomer race.
func (rc *raceCard) toModel(fixture *model.Fixture) (*model.RaceCard, error) {
	grade := model.Grade0W

	name := string(rc.Name)
	subName := string(rc.SubName)
	if strings.Contains(subName, "オープン") {
		grade = model.GradeOP
		if strings.Contains(rc.GradeIconSrc, "jg1") {
			grade = model.GradeJG1
		} else if strings.Contains(rc.GradeIconSrc, "jg2") {
			grade = model.GradeJG2
		} else if strings.Contains(rc.GradeIconSrc, "jg3") {
			grade = model.GradeJG3
		} else if strings.Contains(rc.GradeIconSrc, "g1") {
			grade = model.GradeG1
		} else if strings.Contains(rc.GradeIconSrc, "g2") {
			grade = model.GradeG2
		} else if strings.Contains(rc.GradeIconSrc, "g3") {
			grade = model.GradeG3
		}
	} else if strings.Contains(subName+name, "3勝") {
		grade = model.Grade3W
	} else if strings.Contains(subName+name, "2勝") {
		grade = model.Grade2W
	} else if strings.Contains(subName+name, "1勝") {
		grade = model.Grade1W
	} else if strings.Contains(subName+name, "新馬") {
		grade = model.GradeNewComer
	}

	isSpecial := subName != "" && grade != model.GradeNewComer

	ageGroup, ok := parseAgeGroup(name + subName)
	if !ok {
		return nil, fmt.Errorf("age group not found in race: name=%q subName=%q", name, subName)
	}

	specialName := ""
	if isSpecial {
		specialName = name
	}

	// Some jump races display "芝" in the course cell instead of "芝→ダート",
	// so we override the surface based on the race name/subname.
	surface := model.Surface(rc.Surface)
	if strings.Contains(name+subName, "障害") {
		surface = model.SurfaceJump
	}

	return &model.RaceCard{
		SpecialName: specialName,
		Num:         rc.Num,
		Grade:       grade,
		AgeGroup:    ageGroup,
		Surface:     surface,
		Distance:    int(rc.Distance),
		Runners:     rc.Runners,
		CNAME:       rc.CNAME,
		Fixture:     fixture,
	}, nil
}

type raceName string

var (
	regexpRaceName       = regexp.MustCompile(`\s*牝?(?:[（\(［\[].*)?$`)
	stakesNamesWithParen = []string{"天皇賞(春)", "天皇賞(秋)"}
)

// UnmarshalXPath strips trailing suffixes from a race name scraped from JRA.
// JRA appends gender markers (牝) and parenthetical annotations (e.g. "[指定]", "（混）") to race names.
// Special cases like "天皇賞(春)" and "天皇賞(秋)" are preserved as-is because their parentheses
// are part of the official name, not a suffix.
func (rn *raceName) UnmarshalXPath(text []byte) error {
	s := strings.TrimSpace(string(text))

	if s == "" {
		return nil
	}

	for _, sn := range stakesNamesWithParen {
		if s == sn {
			*rn = raceName(sn)
			return nil
		}
	}

	*rn = raceName(regexpRaceName.ReplaceAllString(s, ""))

	return nil
}

type distance int

func (d *distance) UnmarshalXPath(text []byte) error {
	s := strings.ReplaceAll(string(text), ",", "")
	n, err := strconv.Atoi(s)
	if err != nil {
		return fmt.Errorf("invalid distance found. text=%q", string(text))
	}
	*d = distance(n)
	return nil
}

type surface model.Surface

func (rt *surface) UnmarshalXPath(text []byte) error {
	switch string(text) {
	case "芝":
		*rt = surface(model.SurfaceTurf)
	case "ダート":
		*rt = surface(model.SurfaceDirt)
	case "芝→ダート":
		*rt = surface(model.SurfaceJump)
	default:
		return fmt.Errorf("invalid race type found. text=%q", string(text))
	}
	return nil
}

func (c *JRAClient) ListRaceCards(ctx context.Context, fixture *model.Fixture) ([]*model.RaceCard, error) {
	form := &url.Values{}
	form.Add("cname", fixture.CNAME)

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

	var page raceCardListPage
	if err := xtract.Unmarshal(body, &page); err != nil {
		return nil, err
	}

	raceCards := make([]*model.RaceCard, 0, len(page.RaceCards))
	for _, pr := range page.RaceCards {
		rc, err := pr.toModel(fixture)
		if err != nil {
			return nil, err
		}
		raceCards = append(raceCards, rc)
	}

	return raceCards, nil
}
