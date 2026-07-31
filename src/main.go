package main

import (
	"container/list"
	"fmt"
	"net/http"
	"sync"
	"time"

	"golang.org/x/net/html"
)

func findLinks(links []string, n *html.Node, parentUrl string) []string {
	if n.Type == html.ElementNode && n.Data == "a" {
		for _, attr := range n.Attr {
			if attr.Key != "href" { continue }
			
			if attr.Val[0] == '/' {
				links = append(links, parentUrl + attr.Val[1:])
			} else {
				links = append(links, attr.Val)
			}
		}
	}

	// Recursively traverse child nodes and update the links slice
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		links = findLinks(links, c, parentUrl)
	}

	return links
}

func main() {
	seedUrls := []string{"https://www.gutenberg.org/"}
	queue := list.New()
	var visited sync.Map 
	
	for _, url := range seedUrls {
		queue.PushBack(url)
		visited.Store("url", "")
	}

	for queue.Len() > 0 {
		var wg sync.WaitGroup

		for queue.Len() > 0 {
			urlObject := queue.Front()
			queue.Remove(urlObject)
			if urlObject == nil { continue }

			url, ok := urlObject.Value.(string)
			if !ok { continue }

			fmt.Println(url)

			wg.Go(func() {
				client := &http.Client{}
				req, err := http.NewRequest("GET", url, nil)
				if err != nil {
					fmt.Println(err)
					return 
				}
				req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64)")

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

				var links []string
				nextUrls := findLinks(links, doc, url)

				for _, nextUrl := range nextUrls {
					_, seen := visited.Load(nextUrl)
					if seen { continue }

					queue.PushBack(nextUrl)
					visited.Store(nextUrl, "")
				}
			})
		}

		wg.Wait()
		time.Sleep(500)
	}
}