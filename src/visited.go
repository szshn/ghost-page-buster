package main

type visitRequest struct {
    url   string
    reply chan bool // true if not visited before
}

func visitCoordinator(reqs <-chan visitRequest) {
    visited := make(map[string]struct{})
    for req := range reqs {
        _, seen := visited[req.url]
        if !seen {
            visited[req.url] = struct{}{}
        }
        req.reply <- !seen
    }
}