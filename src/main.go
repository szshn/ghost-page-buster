package main

import (
	"fmt"
	"time"
)

type PageResult struct {
	pageurl      string
	subpages map[string]struct{}
}

func main() {
	ghostbust()
}

func ghostbust() {
	var pageurl string
	fmt.Println("Please type the URL of the page you would like to inspect and press Enter")
	fmt.Scanln(&pageurl)
	// pageurl := "https://pkg.go.dev/rsc.io/quote"  

	start := time.Now()
	inspectPage(pageurl)

	fmt.Printf("Time taken: %v\n", time.Since(start))
}