package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"html"
	"io"
	"net/http"
)

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(ctx context.Context, feedUrl string) (*RSSFeed, error) {
    req, err := http.NewRequestWithContext(ctx, "GET", feedUrl, nil)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error while creating request: %v", err)
    }

    req.Header.Set("User-Agent", "gator")

    client := &http.Client{}

    res, err := client.Do(req)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error while fetching response: %v", err)
    }
    defer res.Body.Close()

    if res.StatusCode < 200 || res.StatusCode >= 300 {
        return &RSSFeed{}, fmt.Errorf("Something went wrong: %v", res.Status)
    }

    body, err := io.ReadAll(res.Body)
    if err != nil {
        return &RSSFeed{}, fmt.Errorf("Error while reading response body: %v", err)
    }

    var result RSSFeed

    if err := xml.Unmarshal(body, &result); err != nil {
        return &RSSFeed{}, fmt.Errorf("Error while unmarshaling xml: %v", err)
    }

    return &result, nil
}

func decodeEscapedHTML(data *RSSFeed) {
    for _, item := range data.Channel.Item {
        item.Title = html.UnescapeString(item.Title)
        item.Description = html.UnescapeString(item.Description)
    }

    data.Channel.Title = html.UnescapeString(data.Channel.Title)
    data.Channel.Description = html.UnescapeString(data.Channel.Description)
}
