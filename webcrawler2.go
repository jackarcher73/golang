package main

import (
	"fmt"
)

type Fetcher interface {
	// Fetch returns the body of URL and
	// a slice of URLs found on that page.
	Fetch(url string) (body string, urls []string, err error)
}

// Crawl uses fetcher to recursively crawl
// pages starting with url, to a maximum of depth.
var enter chan int = make(chan int)
var quit chan int = make(chan int)

type ReadStatus struct {
	url  string
	resp chan int
}

type FetchRequest struct {
	url  string
	depth int
}

var monitor = make(chan *ReadStatus)

func Crawl(in chan *FetchRequest, fetcher Fetcher) {
	// TODO: Fetch URLs in parallel.
	// TODO: Don't fetch the same URL twice.
	// This implementation doesn't do either:
	for request := range in {
		url := request.url
		depth := request.depth
		readStatus := ReadStatus{url, make(chan int)}
		monitor <- &readStatus
		readFlag := <-readStatus.resp
		if readFlag == 1 {
			quit <- 1
			continue
		}

		if depth <= 0 {
			quit <- 1
			continue
		}
		fmt.Println("Craw: ", url)
		body, urls, err := fetcher.Fetch(url)
		if err != nil {
			fmt.Println(err)
			quit <- 1
			continue
		}
		fmt.Printf("found: %s %q\n", url, body)
		for _, u := range urls {
			enter <- 1
			in <- &FetchRequest{u, depth-1}
		}
		quit <- 1
	}
}

func main() {
	go func() {
		crawed := make(map[string]int)
		for readStatus :=  range monitor {
			oldValue := crawed[readStatus.url]
			if oldValue != 1 {
				crawed[readStatus.url] = 1
			}
			readStatus.resp <- oldValue
		}
	}()

	pending := make(chan *FetchRequest)
	for i:=0;i<4;i++ {
		go Crawl(pending, fetcher)
	}

	pending <- &FetchRequest{"http://golang.org/", 4}

	flag := 1
	for {
		select {
		case <-enter:
			flag++
		case <-quit:
			flag--
			if flag == 0 {
				return
			}
		}
	}
}

// fakeFetcher is Fetcher that returns canned results.
type fakeFetcher map[string]*fakeResult

type fakeResult struct {
	body string
	urls []string
}

func (f fakeFetcher) Fetch(url string) (string, []string, error) {
	if res, ok := f[url]; ok {
		return res.body, res.urls, nil
	}
	return "", nil, fmt.Errorf("not found: %s", url)
}

// fetcher is a populated fakeFetcher.
var fetcher = fakeFetcher{
	"http://golang.org/": &fakeResult{
		"The Go Programming Language",
		[]string{
			"http://golang.org/pkg/",
			"http://golang.org/cmd/",
		},
	},
	"http://golang.org/pkg/": &fakeResult{
		"Packages",
		[]string{
			"http://golang.org/",
			"http://golang.org/cmd/",
			"http://golang.org/pkg/fmt/",
			"http://golang.org/pkg/os/",
		},
	},
	"http://golang.org/pkg/fmt/": &fakeResult{
		"Package fmt",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
	"http://golang.org/pkg/os/": &fakeResult{
		"Package os",
		[]string{
			"http://golang.org/",
			"http://golang.org/pkg/",
		},
	},
}
