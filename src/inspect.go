package main

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"golang.org/x/net/html"
)

func inspectPage(pageURL string, depth int, visitReqs chan visitRequest) {
	if depth <= 0 { return }

	reply := make(chan bool)
	visitReqs <- visitRequest{pageURL, reply}
	if !<-reply { return } // already visited

	fmt.Printf("Inspecting %v\n", pageURL)
	r := findSubpages(pageURL)

	var wg sync.WaitGroup
	inactive := make(chan bool)
	unknown := make(chan bool)
	var inactive_count, unknown_count int

	go func() {
		for {
			select {
			case <-inactive:
				inactive_count++
			case <-unknown:
				unknown_count++
			}
		}
	}()

	for subpage := range r.subpages {
		wg.Add(1)
		go func(url string) {
			defer wg.Done()
			inspectSubpage(url, inactive, unknown)
			inspectPage(url, depth-1, visitReqs)
		}(subpage)
	}

	wg.Wait()
	close(inactive)
	close(unknown)

	fmt.Printf("\nResult from %s: %d total pages, %d active pages, %d ghost pages, and %d unknown pages\nDone inspection\n",
		pageURL,
		len(r.subpages),
		len(r.subpages)-inactive_count-unknown_count,
		inactive_count,
		unknown_count)
}

func inspectSubpage(url string, inactive, unknown chan bool) {
	client := &http.Client{Timeout: time.Second * 10}

	resp, err := client.Get(url)
	if err != nil {
		unknown <- true
		fmt.Printf("%v -- ❓ Error occurred %v\n", url, err)
		return
	}
	defer resp.Body.Close()

	var statusMsg string
	switch resp.StatusCode {
	case 200:
		statusMsg = "✅"
	case 404, 410:
		statusMsg = "❌"
		inactive <- true
	default:
		if strings.Contains(url, "twitter.com") {
			statusMsg = "❓ (Unable to verify Twitter at the moment)"
		} else {
			statusMsg = "❓ (Unable to verify at the moment)"
		}
		unknown <- true
	}
	fmt.Printf("%v -- Status code %v %s\n", url, resp.StatusCode, statusMsg)
}

func findSubpages(orig string) PageResult {
	pageurl := normalizeURL(orig)
	client := &http.Client{Timeout: time.Second}

	resp, err := client.Get(pageurl)
	if err != nil {
		fmt.Printf("%v -- Error occurred %v\n", pageurl, err)
		return PageResult{}
	}
	defer resp.Body.Close()

	result := PageResult{pageurl, make(map[string]struct{})}
	z := html.NewTokenizer(resp.Body)

	for {
		tokenType := z.Next()
		if tokenType == html.ErrorToken {
			break
		}
		if tokenType != html.StartTagToken {
			continue
		}

		t := z.Token()
		if t.Data != "a" {
			continue
		}

		for _, a := range t.Attr {
			if a.Key == "href" {
				// Add to a list of subpages
				if strings.HasPrefix(a.Val, "http") {
					result.subpages[normalizeURL(a.Val)] = struct{}{}
				} else if strings.HasPrefix(a.Val, "?") {
					result.subpages[normalizeURL(pageurl+a.Val)] = struct{}{}
				} else if strings.HasPrefix(a.Val, "#") && a.Val != "#" {
					// check that the corresponding element exists in the html body
				} else if strings.HasPrefix(a.Val, "/") {
					parsedURL, err := url.Parse(pageurl)
					if err != nil {
						fmt.Println("Error parsing url:", err)
						return PageResult{}
					}
					fullURL := "https://" + parsedURL.Hostname() + a.Val

					result.subpages[normalizeURL(fullURL)] = struct{}{}
				}
				break
			}
		}
	}

	return result
}

func normalizeURL(raw string) string {
	u, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	if u.Scheme == "http" || u.Scheme == "" {
		u.Scheme = "https"
	}

	// Remove 'www.' prefix if present
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.Path = strings.TrimRight(u.Path, "/")
	u.Fragment = ""

	return u.String()
}