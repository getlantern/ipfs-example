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
	"sort"
	"sync"
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

	feed = &Feed{
		Sources: []*Source{
			&Source{Name: "BBC", Url: "http://www.bbc.com/zhongwen/simp/indepth/cluster_panama_papers/index.xml"},
			&Source{Name: "China Digital Times", Url: "https://chinadigitaltimes.net/feed/"},
		},
		Items: make(map[string]FeedItems),
	}

	feedMap = map[string]string{
		"http://www.bbc.com/zhongwen/simp/indepth/cluster_panama_papers/index.xml": "BBC",
		"https://theinitium.com/newsfeed/":                                         "The Initium",
		"https://chinadigitaltimes.net/feed/":                                      "China Digital Times",
	}

	timeLayout = "Mon Jan 2 15:04:05 MST"
	timeout    = 5
)

type Feed struct {
	Sources []*Source
	Items   map[string]FeedItems
	Full    FeedItems
}

type FeedItem struct {
	Image       string
	Description string
	Title       string
	Url         string
	Source      string
	Date        time.Time
}

type FeedItems []FeedItem

func (f FeedItems) Len() int {
	return len(f)
}

func (f FeedItems) Less(i, j int) bool {
	return f[i].Date.Before(f[j].Date)
}

func (f FeedItems) Swap(i, j int) {
	f[i], f[j] = f[j], f[i]
}

type Source struct {
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
	var wg sync.WaitGroup
	wg.Add(len(feed.Sources))

	for _, source := range feed.Sources {
		go func(url string) {
			defer wg.Done()
			pollFeed(url, timeout, charsetReader)
		}(source.Url)
	}

	wg.Wait()
	publishFeed()
}

func pollFeed(uri string, timeout int, cr xmlx.CharsetFunc) error {

	r := rss.New(timeout, true, nil, itemHandler)

	if err := r.Fetch(uri, cr); err != nil {
		log.Errorf("[e] %s: %s\n", uri, err)
		return err
	}

	return nil
}

func encodeFeed(items []FeedItem) (string, error) {
	var buffer bytes.Buffer        // Stand-in for a network connection
	enc := gob.NewEncoder(&buffer) // Will write to network.

	err := enc.Encode(items)
	if err != nil {
		log.Fatalf("Encode error: %v", err)
	}

	return createTempFile(buffer.Bytes())
}

func encodeFeedJson() (string, error) {

	b, err := json.Marshal(feed)
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

func itemHandler(r *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
	log.Debugf("%d new item(s) in %s\n", len(newitems), r.Url)

	for i := 0; i < len(newitems) && i < 10; i++ {
		item := newitems[i]
		link := item.Links[0]
		if link != nil && link.Href != "" {
			item, err := parseFeedItem(feedMap[r.Url], link.Href)
			if err != nil {
				log.Errorf("Could not parse feed item: %v", err)
			} else {
				printFeedItem(item)
				if item.Title != "" {
					feed.Items[item.Source] = append(feed.Items[item.Source], *item)
					feed.Full = append(feed.Full, *item)
				}
			}
		}
	}
}

func publishFeed() {
	sort.Sort(feed.Full)

	fn, err := encodeFeedJson()
	if err != nil {
		log.Errorf("Error encoding feed: %v", err)
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

func parseFeedItem(source, url string) (*FeedItem, error) {
	opt := &diffbot.Options{Fields: "*"}
	article, err := diffbot.ParseArticle(token, url, opt)
	if err != nil {
		if apiErr, ok := err.(*diffbot.Error); ok {
			// ApiError, e.g. {"error":"Not authorized API token.","errorCode":401}
			log.Error(apiErr)
			return nil, apiErr
		}
		log.Error(err)
		return nil, err
	}

	log.Debugf("DIFFBOT ARTICLE: %v", article)

	item := &FeedItem{
		Title:  article.Title,
		Url:    article.Url,
		Source: source,
	}

	t, err := time.Parse(article.Date, timeLayout)
	if err != nil {
		item.Date = t
	}

	if desc := article.Meta["description"]; desc != nil {
		item.Description = desc.(string)
	}

	for _, img := range article.Images {
		if img.Primary == "true" {
			item.Image = img.Url
			break
		}
	}
	return item, nil
}

func printFeedItem(item *FeedItem) {
	log.Debugf("URL: %s TITLE: %s TEXT: %s IMAGE: %s",
		item.Url, item.Title,
		item.Description, item.Image)
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
