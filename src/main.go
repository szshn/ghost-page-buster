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
	var depth int
	fmt.Println("Please type the URL of the page you would like to inspect and desired query depth. Press Enter to contninue.")
	fmt.Println("(e.g. www.example.com 3)")
	fmt.Scanf("%s %d\n", &pageurl, &depth)

	visitReqs := make(chan visitRequest)
	go visitCoordinator(visitReqs)
	
	start := time.Now()
	inspectPage(pageurl, depth, visitReqs)
	fmt.Printf("Time taken: %v\n", time.Since(start))

	close(visitReqs)
}