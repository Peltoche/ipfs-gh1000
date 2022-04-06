package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"go.uber.org/multierr"
)

type result struct {
	RepositoryURL     string
	Rank              int
	NbStars           int
	LastMetadataFetch time.Time
}

func fetchMetadatasFromLink(ctx context.Context, link string) (*result, error) {
	res, err := http.Get("https://gitstar-ranking.com" + link)
	if err != nil {
		return nil, fmt.Errorf("failed to get the gitstar-ranking.com page: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status during the metadata fetch: %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the metadata HTML: %w", err)
	}

	rank := 0
	stars := 0

	doc.Find(".repository_info").Each(func(i int, s *goquery.Selection) {
		s.Find(".col-xs-6").Each(func(i int, s2 *goquery.Selection) {
			attribute := strings.TrimSpace(s2.Find(".repository_attribute").Text())
			rawValue := strings.TrimSpace(s2.Find(".repository_value").Text())

			// log.Printf("raw value of %v: %v\n", attribute, rawValue)

			value, err := strconv.Atoi(rawValue)
			if err != nil {
				err = multierr.Append(err, fmt.Errorf("failed to parse the value of attribute %q: %w", attribute, err))
				return
			}

			switch attribute {
			case "Star":
				stars = value
				break
			case "Rank":
				rank = value
				break
			}
		})
	})
	if err != nil {
		return nil, err
	}

	rawLastUpdate := doc.Find(".queued_at").Text()
	lastUpdateParts := strings.Split(rawLastUpdate, "on")
	if len(lastUpdateParts) != 2 {
		return nil, fmt.Errorf("invalid format for the last update field (%q): expected something like \"Fetch on 2022/04/05 11:40\" ", lastUpdateParts)
	}

	rawDate := strings.TrimSpace(lastUpdateParts[1])

	lastUpdate, err := time.Parse("2006/01/02 15:04", rawDate)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the lastUpdate time %q: %w", rawDate, err)
	}

	return &result{
		RepositoryURL:     "https://github.com" + link,
		Rank:              rank,
		NbStars:           stars,
		LastMetadataFetch: lastUpdate,
	}, nil
}

func main() {
	res, err := http.Get("https://gitstar-ranking.com/repositories")
	if err != nil {
		log.Fatalf("failed to fetch the listing page: %s", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		log.Fatalf("invalid status during the fetch of the listing page errors: %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatalf("failed to parse the site listing page HTML: %s", err)
	}

	links := []string{}

	// Find the review items
	doc.Find(".list-group-item").Each(func(i int, s *goquery.Selection) {
		// For each item found, get the title
		link, _ := s.Attr("href")
		links = append(links, link)
	})

	ctx := context.Background()

	results := make([]result, len(links))

	for i, link := range links {
		res, err := fetchMetadatasFromLink(ctx, link)
		if err != nil {
			log.Fatalf("failed to fetch metadata for %s: %s", link, err)
		}

		results[i] = *res
		fmt.Printf("%d -> %+v\n", i, res)
	}

}
