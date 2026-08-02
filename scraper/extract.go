package scraper

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
	"unicode"

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

func normalizeWord(word string) string {
	chars := make([]rune, 0, len(word))

	for _, r := range word {
		if !unicode.IsLetter(r) {
			continue
		}

		r = unicode.ToLower(r)
		chars = append(chars, r)
	}

	return string(chars)
}

func extractWords(n *html.Node, pageURL string, wordsToUrls *WordsToUrls) {
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style" || n.Data == "noscript") {
		return
	}

	if n.Type == html.TextNode {
		for word := range strings.FieldsSeq(n.Data) {
			word := normalizeWord(word)

			if len(word) == 0 {
				continue
			}

			wordsToUrls.Add(word, pageURL)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		extractWords(c, pageURL, wordsToUrls)
	}
}

func extractFromPage(botAgent string, pageURL string, wordsToUrls *WordsToUrls, results chan<- ScrapeResults) {
	client := &http.Client{Timeout: 2 * time.Second}
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
	extractWords(doc, pageURL, wordsToUrls)

	results <- ScrapeResults{nextURLs: links}
}
