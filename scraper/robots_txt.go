package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/temoto/robotstxt"
)

type RobotConfig struct {
	Allowed         bool
	CrawlDelay      time.Duration
	LastRequestTime int64
}

type RobotsRules struct {
	userAgent    string
	robotsConfig sync.Map
}

func (r *RobotsRules) GetRobotConfig(targetURL string) (*RobotConfig, error) {
	parsedURL, err := url.Parse(targetURL)

	if err != nil {
		return nil, err
	}

	host := parsedURL.Host
	value, exists := r.robotsConfig.Load(host)

	if exists {
		return value.(*RobotConfig), nil
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, host)
	resp, err := http.Get(robotsURL)

	if err != nil {
		return nil, nil
	}

	defer resp.Body.Close()

	robotsData, err := robotstxt.FromResponse(resp)

	if err != nil {
		return nil, nil
	}

	group := robotsData.FindGroup(r.userAgent)
	allowed := group.Test(parsedURL.Path)
	config := &RobotConfig{
		Allowed:         allowed,
		CrawlDelay:      group.CrawlDelay,
		LastRequestTime: 0,
	}
	r.robotsConfig.Store(host, config)

	return config, nil
}
