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
	var pageurl string
	fmt.Println("Please type the URL of the page you would like to inspect and press Enter")
	fmt.Scanln(&pageurl)

	visitReqs := make(chan visitRequest)
	go visitCoordinator(visitReqs)
	
	start := time.Now()
	inspectPage(pageurl, 2, visitReqs)
	fmt.Printf("Time taken: %v\n", time.Since(start))

	close(visitReqs)
}