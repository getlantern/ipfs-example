package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"path/filepath"
	"time"

	"github.com/diffbot/diffbot-go-client"
	"github.com/getlantern/golog"
	"github.com/getlantern/ipfs-example/ipfs"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/jteeuwen/go-pkg-xmlx"
)

var (
	log      = golog.LoggerFor("lantern.everfeed.extractor")
	url      = flag.String("url", "https://chinadigitaltimes.net/feed/", "")
	token    = flag.String("token", "6c6ab23583a10bc48c65ca2a1ff78b43", "diffbot token")
	articles []Article
)

type Article struct {
	Image string
	Text  string
	Title string
	Url   string
}

func pollFeed(uri string, timeout int, cr xmlx.CharsetFunc) (string, error) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	if err := feed.Fetch(uri, cr); err != nil {
		log.Errorf("[e] %s: %s\n", uri, err)
		return "", err
	}

	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.

	err := enc.Encode(articles)
	if err != nil {
		log.Fatalf("Encode error: %v", err)
	}

	log.Debugf("Bytes is %v", buffer.Bytes())
	dir, err := ioutil.TempDir("", "lantern")
	if err != nil {
		log.Errorf("Could not write tmp file: %v", err)
		return "", err
	}

	tmpfn := filepath.Join(dir, "feed")
	if err := ioutil.WriteFile(tmpfn, buffer.Bytes(), 0666); err != nil {
		return "", err
	}
	return tmpfn, nil
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	log.Debugf("%d new item(s) in %s\n", len(newitems), feed.Url)

	for i := 0; i < len(newitems) && i <= 3; i++ {
		item := newitems[i]
		link := item.Links[0]
		if link != nil && link.Href != "" {
			parseArticle(link.Href)
		}
	}
}

func parseArticle(url string) {
	opt := &diffbot.Options{Fields: "*"}
	dArticle, err := diffbot.ParseArticle(*token, url, opt)
	if err != nil {
		if apiErr, ok := err.(*diffbot.Error); ok {
			// ApiError, e.g. {"error":"Not authorized API token.","errorCode":401}
			log.Error(apiErr)
		}
		log.Error(err)
	}

	article := Article{
		Title: dArticle.Title,
		Url:   dArticle.Url,
		Text:  dArticle.Text,
	}
	for _, img := range dArticle.Images {
		if img.Primary == "true" {
			article.Image = img.Url
			break
		}
	}
	articles = append(articles, article)
	printArticle(article)
}

func printArticle(article Article) {
	log.Debugf("URL: %s TITLE: %s TEXT: %s IMAGE: %s",
		article.Url, article.Title,
		article.Text, article.Image)
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}

func main() {
	flag.Parse()

	node, err := ipfs.Start("~/.ipfs")
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("Fetching articles from: %s", *url)
	interval := 1 * time.Minute
	t := time.NewTimer(0)
	for {
		<-t.C
		t.Reset(interval)
		fn, err := pollFeed(*url, 5, charsetReader)
		if err != nil {
			log.Error(err)
			return
		}
		path, _, err := node.AddFile(fn, "CoolSite")
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("Added at /ipfs/%s", path)

		ns, err := node.Publish(path)
		if err != nil {
			log.Error(err)
			return
		}
		log.Debugf("Published to /ipns/%s", ns)
	}
}
