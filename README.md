# Ghost (Page) Buster

**Ghost (Page) Busters** is a fast, concurrent web crawler built in Go that recursively checks for dead or invalid links on a webpage and its subpages. Inspired by the frustration of broken documentation or outdated resources, this tool aims to help developers and site maintainers quickly audit their websites.

## Features

- Recursively scans webpages for all internal and external links
- Concurrent HTTP requests using goroutines for fast performance
- Normalizes URLs to avoid redundant checks (e.g., handling `http`, `https`, `www`)
- Identifies and categorizes dead links with appropriate HTTP status codes
### Planned
- CLI interface with adjustable crawl depth
- Optional integration with [Google Safe Browsing API](https://developers.google.com/safe-browsing) for malicious URL detection

## Usage

```
$ go run ./src
Please type the URL of the page you would like to inspect and press Enter
...
Inspecting ...
```

## Performance
When tested on a page with 330 distinct links, the tool completed crawling in ~1.2 seconds using concurrent HTTP requests (depth=1, timeout=1s). 

In contrast, sequential execution would have taken over 5 minutes.

## Structure
`main.go` CLI entry point and argument parsing \
`inspect.go` Core logic for crawling, parsing, and validating links

## Requirements
Go 1.18+\
Internet connection for HTTP requests

## Installation
```
git clone https://github.com/your-username/ghost-page-busters.git
cd ghost-page-busters
go run ./src
```

## TODO
Add support for robots.txt and crawl delay\
Integrate Safe Browsing API