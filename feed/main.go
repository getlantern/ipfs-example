package main

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"time"

	"github.com/diffbot/diffbot-go-client"
	"github.com/getlantern/golog"
	"github.com/getlantern/ipfs-example/ipfs"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/jteeuwen/go-pkg-xmlx"
)

var (
	log   = golog.LoggerFor("lantern.everfeed.extractor")
	node  *ipfs.IpfsNode
	token string

	feeds = []*Feed{
		&Feed{Name: "BBC", Url: "http://www.bbc.com/zhongwen/simp/indepth/cluster_panama_papers/index.xml"},
		/*&Feed{Name: "China Digital Times", Url: "https://chinadigitaltimes.net/feed/"},
		&Feed{Name: "Solidot", Url: "http://solidot.org.feedsportal.com/c/33236/f/556826/index.rss"},
		&Feed{Name: "Boxun", Url: "http://www.boxun.com/news/rss/focus.xml"},
		&Feed{Name: "NYTimes", Url: "http://cn.nytimes.com/rss.html"},*/
	}

	timeout = 5
)

type Article struct {
	Image       string
	Description string
	Title       string
	Url         string
}

type Feed struct {
	Name string
	Url  string
}

func init() {
	var err error

	token = os.Getenv("DIFFBOT_TOKEN")
	if token == "" {
		log.Fatal("No diffbot token!")
	}

	node, err = ipfs.Start("~/.ipfs")
	if err != nil {
		log.Fatal(err)
	}
}

func syncFeeds() {

	for _, feed := range feeds {
		go pollFeed(feed.Url, timeout, charsetReader)
	}
}

func pollFeed(uri string, timeout int, cr xmlx.CharsetFunc) error {
	feed := rss.New(timeout, true, nil, itemHandler)

	if err := feed.Fetch(uri, cr); err != nil {
		log.Errorf("[e] %s: %s\n", uri, err)
		return err
	}

	return nil
}

func encodeFeed(articles []Article) (string, error) {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.

	err := enc.Encode(articles)
	if err != nil {
		log.Fatalf("Encode error: %v", err)
	}

	return createTempFile(buffer.Bytes())
}

func encodeFeedJson(articles []Article) (string, error) {
	b, err := json.Marshal(articles)
	if err != nil {
		log.Fatal(err)
	}
	return createTempFile(b)
}

func createTempFile(bytes []byte) (string, error) {
	dir, err := ioutil.TempDir("", "lantern")
	if err != nil {
		log.Errorf("Could not write tmp file: %v", err)
		return "", err
	}

	tmpfn := filepath.Join(dir, "feed")
	if err := ioutil.WriteFile(tmpfn, bytes, 0666); err != nil {
		return "", err
	}
	return tmpfn, nil
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	log.Debugf("%d new item(s) in %s\n", len(newitems), feed.Url)

	var articles []Article

	for i := 0; i < len(newitems) && i < 10; i++ {
		item := newitems[i]
		link := item.Links[0]
		if link != nil && link.Href != "" {
			article, err := parseArticle(link.Href)
			if err != nil {
				log.Errorf("Could not parse article: %v", err)
			} else {
				printArticle(article)
				if article.Title != "" {
					articles = append(articles, *article)
				}
			}
		}
	}

	log.Debugf("Number of articles: %d", len(articles))

	fn, err := encodeFeedJson(articles)
	if err != nil {
		log.Errorf("Error encoding feed: %v", err)
		return
	}

	path, _, err := node.AddFile(fn, "CoolSite")
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Added feed %s at /ipfs/%s", feed.Url, path)

	ns, err := node.Publish(path)
	if err != nil {
		log.Error(err)
		return
	}
	log.Debugf("Published to /ipns/%s", ns)
}

func parseArticle(url string) (*Article, error) {
	opt := &diffbot.Options{Fields: "*"}
	dArticle, err := diffbot.ParseArticle(token, url, opt)
	if err != nil {
		if apiErr, ok := err.(*diffbot.Error); ok {
			// ApiError, e.g. {"error":"Not authorized API token.","errorCode":401}
			log.Error(apiErr)
			return nil, apiErr
		}
		log.Error(err)
		return nil, err
	}

	log.Debugf("DIFFBOT ARTICLE: %v", dArticle.Meta)

	article := &Article{
		Title:       dArticle.Title,
		Url:         dArticle.Url,
		Description: dArticle.Meta["description"].(string),
	}
	for _, img := range dArticle.Images {
		if img.Primary == "true" {
			article.Image = img.Url
			break
		}
	}
	return article, nil
}

func printArticle(article *Article) {
	log.Debugf("URL: %s TITLE: %s TEXT: %s IMAGE: %s",
		article.Url, article.Title,
		article.Description, article.Image)
}

func charsetReader(charset string, r io.Reader) (io.Reader, error) {
	if charset == "ISO-8859-1" || charset == "iso-8859-1" {
		return r, nil
	}
	return nil, errors.New("Unsupported character set encoding: " + charset)
}

func main() {

	interval := 10 * time.Minute
	t := time.NewTimer(0)
	for {
		<-t.C
		t.Reset(interval)
		syncFeeds()
	}
}
