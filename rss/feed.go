package rss

import (
	"fmt"
	"time"

	"github.com/mmcdole/gofeed"
)

type FeedItem struct {
	Title           string
	Description     string
	Summary         string
	Link            string
	PublicationDate time.Time
}

type FeedProcessor struct {
	parser *gofeed.Parser
}

func NewFeedProcessor() *FeedProcessor {
	return &FeedProcessor{
		parser: gofeed.NewParser(),
	}
}

func (fp *FeedProcessor) ParseFeedFromURL(url string) ([]*FeedItem, error) {
	feed, err := fp.parser.ParseURL(url)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed from URL %s: %w", url, err)
	}

	return fp.extractItems(feed), nil
}

func (fp *FeedProcessor) ParseFeedFromString(feedContent string) ([]*FeedItem, error) {
	feed, err := fp.parser.ParseString(feedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse feed from string: %w", err)
	}

	return fp.extractItems(feed), nil
}

func (fp *FeedProcessor) extractItems(feed *gofeed.Feed) []*FeedItem {
	items := make([]*FeedItem, 0, len(feed.Items))

	for _, item := range feed.Items {
		feedItem := &FeedItem{
			Title:       item.Title,
			Description: item.Description,
			Summary:     item.Content,
			Link:        item.Link,
		}

		if item.PublishedParsed != nil {
			feedItem.PublicationDate = *item.PublishedParsed
		} else if item.UpdatedParsed != nil {
			feedItem.PublicationDate = *item.UpdatedParsed
		} else {
			feedItem.PublicationDate = time.Now()
		}

		if feedItem.Summary == "" {
			feedItem.Summary = feedItem.Description
		}

		items = append(items, feedItem)
	}

	return items
}
