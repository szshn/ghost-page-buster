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
	client := &http.Client{Timeout: time.Second * 10}

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
				handleHref(pageurl, a.Val, result.subpages)
				break
			}
		}
	}

	return result
}

func handleHref(pageurl, val string, history map[string]struct{}) {
	if  strings.HasPrefix(val, "mailto:") ||	// Ignore mail, phone, and javascript protocols
		strings.HasPrefix(val, "sms:") ||
		strings.HasPrefix(val, "tel:") || 
		strings.HasPrefix(val, "javascript:") {
		return
		
	} else if strings.HasPrefix(val, "http") {	// Absolute URL
		history[normalizeURL(val)] = struct{}{}

	} else if strings.HasPrefix(val, "/") {		// Root directory relative
		parsedURL, err := url.Parse(pageurl)
		if err != nil {
			fmt.Printf("Error parsing %s: %v\n", pageurl, err)
			return
		}
		fullURL := "https://" + parsedURL.Hostname() + val
		history[normalizeURL(fullURL)] = struct{}{}
		
	} else if strings.HasPrefix(val, "../") {	// Parent relative
		fullURL := strings.TrimSuffix(pageurl, "/")

		path := val
		for strings.HasPrefix(path, "../") {
			// scope out from current directory
			if idx := strings.LastIndex(fullURL, "/"); idx != -1 {
				fullURL = fullURL[:idx-1]
			}
			path = strings.TrimPrefix(path, "../")	// remove ../ in path
		}
		fullURL += path

		history[normalizeURL(fullURL)] = struct{}{}

	} else if strings.HasPrefix(val, "#") { // && a.Val != "#" {
		// check that the corresponding element exists in the html body?
		return
	} else { // Current directory relative
		fullURL := strings.TrimSuffix(pageurl, "/")				
		if idx := strings.LastIndex(fullURL, "/"); idx != -1 {
			fullURL = fullURL[:idx-1]
		}
		path := strings.TrimPrefix(val, "/")
		fullURL += "/" + path

		history[normalizeURL(fullURL)] = struct{}{}
	}
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
	// TODO--some website needs www for requests else returns 403
	u.Host = strings.TrimPrefix(u.Host, "www.")
	u.Path = strings.TrimRight(u.Path, "/")
	u.Fragment = ""

	return u.String()
}