package jra

import (
	"time"

	"github.com/octarect/turf/turf"
)

const (
	jraAccessSPath = "/JRADB/accessS.html"
	jraAccessRPath = "/JRADB/accessR.html"
)

var timeJST = time.FixedZone("Asia/Tokyo", 9*60*60)

// JRAClient is a scraper client for the JRA website (https://www.jra.go.jp/).
// It implements FixtureLister, RaceCardLister, and RaceResultGetter.
type JRAClient struct {
	client *turf.Client
}

var (
	_ = turf.FixtureLister(&JRAClient{})
	_ = turf.RaceCardLister(&JRAClient{})
	_ = turf.RaceResultGetter(&JRAClient{})
)

func NewJRAClient(client *turf.Client) *JRAClient {
	return &JRAClient{
		client: client,
	}
}
