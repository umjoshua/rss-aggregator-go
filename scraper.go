package main

import (
	"context"
	"database/sql"
	"encoding/xml"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/umjoshua/rss-aggregator-go/internal/database"
)

func startScraping(db *database.Queries, concurrency int, intervel time.Duration) {
	ticker := time.NewTicker(intervel)

	for ; ; <-ticker.C {
		feeds, err := db.GetNextFeedsToFetch(context.Background(), int32(concurrency))
		if err != nil {
			log.Println("Couldn't get next feeds", err)
			continue
		}

		wg := &sync.WaitGroup{}

		for _, feed := range feeds {
			wg.Add(1)
			go scrapeFeed(db, wg, feed)
		}
		wg.Wait()
	}
}

func scrapeFeed(db *database.Queries, wg *sync.WaitGroup, feed database.Feed) {
	defer wg.Done()

	_, err := db.MarkFeedFetched(context.Background(), feed.ID)

	if err != nil {
		log.Println("Couldn't mark feed", feed.Name, "fetched", err)
		return
	}

	feedData, err := fetchFeed(feed.Url)
	if err != nil {
		log.Println("Couldn't collect feed", feed.Name, "fetched", err)
		return
	}

	for _, item := range feedData.Channel.Item {
		description := sql.NullString{}

		if item.Description != "" {
			description.String = item.Description
		}

		pubAt, err := time.Parse(time.RFC1123Z, item.PubDate)

		if err != nil {
			continue
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			ID:          uuid.New(),
			CreatedAt:   time.Now().UTC(),
			UpdatedAt:   time.Now().UTC(),
			Title:       item.Title,
			Description: description,
			PublishedAt: pubAt,
			Url:         item.Link,
			FeedID:      feed.ID,
		})

		if err != nil {
			if strings.Contains(err.Error(), "duplicate key") {
				continue
			}
			log.Printf("failed to insert post %v", err)
			continue
		}
	}

	log.Printf("Feed %v collected", feed.Name)
}

type RSSFeed struct {
	Channel struct {
		Title       string    `xml:"title"`
		Link        string    `xml:"link"`
		Description string    `xml:"description"`
		Language    string    `xml:"language"`
		Item        []RSSItem `xml:"item"`
	} `xml:"channel"`
}

type RSSItem struct {
	Title       string `xml:"title"`
	Link        string `xml:"link"`
	Description string `xml:"description"`
	PubDate     string `xml:"pubDate"`
}

func fetchFeed(feedUrl string) (*RSSFeed, error) {
	httpClient := http.Client{
		Timeout: time.Second * 10,
	}
	data, err := httpClient.Get(feedUrl)

	if err != nil {
		return nil, err
	}
	defer data.Body.Close()

	dat, err := io.ReadAll(data.Body)

	if err != nil {
		return nil, err
	}

	var rssFeed RSSFeed

	err = xml.Unmarshal(dat, &rssFeed)

	if err != nil {
		return nil, err
	}

	return &rssFeed, nil

}
