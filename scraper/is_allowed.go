package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"sync"

	"github.com/temoto/robotstxt"
)

type RobotsGuard struct {
	userAgent     string
	robotsAllowed sync.Map
}

func (g *RobotsGuard) IsAllowed(targetURL string) (bool, error) {
	value, exists := g.robotsAllowed.Load(targetURL)

	if exists {
		return value.(bool), nil
	}

	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return false, err
	}

	robotsURL := fmt.Sprintf("%s://%s/robots.txt", parsedURL.Scheme, parsedURL.Host)
	resp, err := http.Get(robotsURL)

	if err != nil {
		return false, nil
	}

	defer resp.Body.Close()

	robotsData, err := robotstxt.FromResponse(resp)

	if err != nil {
		return false, nil 
	}

	group := robotsData.FindGroup(g.userAgent)

	allowed := group.Test(parsedURL.Path)
	g.robotsAllowed.Store(targetURL, allowed)

	return allowed, nil
}