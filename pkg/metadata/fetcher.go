package metadata

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ipfs/go-cid"
	"go.uber.org/multierr"
)

type RepoMetadata struct {
	RepositoryURL     string
	Rank              int
	NbStars           int
	LastMetadataFetch time.Time
	Repo              *cid.Cid
}

type Fetcher struct {
	url *url.URL
}

func NewFetcher(urlStr string) (*Fetcher, error) {
	u, err := url.Parse(urlStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the url: %w", err)
	}

	return &Fetcher{
		url: u,
	}, nil
}

func (f *Fetcher) FetchLinkPage(ctx context.Context) ([]string, error) {
	res, err := http.Get(f.url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get the page: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status: %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the HTML: %w", err)
	}

	links := []string{}

	doc.Find(".list-group-item").Each(func(i int, s *goquery.Selection) {
		link, _ := s.Attr("href")

		if link == "" {
			return
		}

		link = strings.TrimPrefix(link, "/")

		links = append(links, link)
	})

	return links, nil
}

func (f *Fetcher) FetchMetadataForLink(ctx context.Context, link string) (*RepoMetadata, error) {
	url, err := f.url.Parse(link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the link %q: %w", link, err)
	}

	res, err := http.Get(url.String())
	if err != nil {
		return nil, fmt.Errorf("failed to get the page: %w", err)
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != 200 {
		return nil, fmt.Errorf("invalid status code: %s", res.Status)
	}

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the HTML: %w", err)
	}

	rank, stars, err := f.parseRankAndStars(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the rank and stars: %w", err)
	}

	lastUpdateDate, err := f.parseLastUpdateDate(doc)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the last update date: %w", err)
	}

	repoURL, err := url.Parse("https://github.com/" + link)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the repo url: %w", err)
	}

	return &RepoMetadata{
		RepositoryURL:     repoURL.String(),
		Rank:              rank,
		NbStars:           stars,
		LastMetadataFetch: lastUpdateDate,
		Repo:              nil,
	}, nil
}

func (f *Fetcher) parseLastUpdateDate(doc *goquery.Document) (time.Time, error) {
	rawLastUpdate := doc.Find(".queued_at").Text()

	// Expect the following format: "Fetch on 2022/04/05 11:40"
	lastUpdateParts := strings.Split(rawLastUpdate, "on")
	if len(lastUpdateParts) != 2 {
		return time.Time{}, fmt.Errorf("invalid format, can't separate the date: %q", lastUpdateParts)
	}

	rawDate := strings.TrimSpace(lastUpdateParts[1])

	lastUpdate, err := time.Parse("2006/01/02 15:04", rawDate)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to parse the lastUpdate time %q: %w", rawDate, err)
	}

	return lastUpdate, nil
}

func (f *Fetcher) parseRankAndStars(doc *goquery.Document) (int, int, error) {
	var (
		rank  int
		stars int
		err   error
	)

	doc.Find(".repository_info").Each(func(i int, s *goquery.Selection) {
		s.Find(".col-xs-6").Each(func(i int, s2 *goquery.Selection) {
			attribute := strings.TrimSpace(s2.Find(".repository_attribute").Text())
			rawValue := strings.TrimSpace(s2.Find(".repository_value").Text())

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

	return rank, stars, err
}
