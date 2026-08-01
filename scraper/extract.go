package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"golang.org/x/net/html"
)

type ScrapeResults struct {
	nextURLs []string
}

func extractLinks(n *html.Node, base *url.URL) []string {
	var links []string

	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key != "href" || attr.Val == "" {
				continue
			}

			if ref, err := url.Parse(attr.Val); err == nil {
				links = append(links, base.ResolveReference(ref).String())
			}
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = append(links, extractLinks(c, base)...)
	}

	return links
}

func extractWords(n *html.Node, pageURL string, words *sync.Map) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "noscript") {
		return
	}

	if n.Type == html.TextNode {
		for word := range strings.FieldsSeq(n.Data) {
			actual, _ := words.LoadOrStore(word, []string{})
			words.Store(word, append(actual.([]string), pageURL))
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractWords(c, pageURL, words)
	}
}

func extractFromPage(botAgent string, pageURL string, wordToUrls *sync.Map, results chan<- ScrapeResults) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", pageURL, nil)

	if err != nil {
		fmt.Println(err)
		return
	}

	req.Header.Set("User-Agent", botAgent)
	response, err := client.Do(req)

	if err != nil {
		fmt.Println(err)
		return
	}

	defer response.Body.Close()

	doc, err := html.Parse(response.Body)

	if err != nil {
		fmt.Println(err)
		return
	}

	base, _ := url.Parse(pageURL)
	links := extractLinks(doc, base)
	extractWords(doc, pageURL, wordToUrls)

	results <- ScrapeResults{nextURLs: links}
}