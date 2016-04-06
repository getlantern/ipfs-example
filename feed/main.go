package main

import (
	"bytes"
	"encoding/gob"
	"errors"
	"flag"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/diffbot/diffbot-go-client"
	"github.com/getlantern/golog"
	"github.com/ipfs/go-ipfs/core"
	fsrepo "github.com/ipfs/go-ipfs/repo/fsrepo"
	rss "github.com/jteeuwen/go-pkg-rss"
	"github.com/jteeuwen/go-pkg-xmlx"
	"golang.org/x/net/context"
)

var (
	log = golog.LoggerFor("lantern.everfeed.extractor")
	url = flag.String("url", "https://chinadigitaltimes.net/feed/", "")

	token    string
	articles []Article
)

type Article struct {
	Image string
	Text  string
	Title string
	Url   string
}

func init() {
	token = os.Getenv("DIFFBOT_TOKEN")
	if token == "" {
		log.Fatal("No diffbot token!")
	}
}

func pollFeed(uri string, timeout int, cr xmlx.CharsetFunc) {
	feed := rss.New(timeout, true, chanHandler, itemHandler)

	if err := feed.Fetch(uri, cr); err != nil {
		log.Errorf("[e] %s: %s\n", uri, err)
		return
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
		log.Fatalf("Could not write tmp file: %v", err)
		return
	}

	tmpfn := filepath.Join(dir, "feed")
	if err := ioutil.WriteFile(tmpfn, buffer.Bytes(), 0666); err != nil {
		log.Fatal(err)
	}

	addToIpfs(tmpfn)
}

func addToIpfs(file string) {
	log.Debugf("File path is %s", file)
	cmd := exec.Command("ipfs", "add", file)

	err := cmd.Start()
	if err != nil {
		log.Fatal(err)
	}
	log.Debugf("Waiting for command to finish...")
	err = cmd.Wait()
	log.Debugf("Command finished with error: %v", err)

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
	dArticle, err := diffbot.ParseArticle(token, url, opt)
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

func setupIpfs() (*core.IpfsNode, error) {
	// Assume the user has run 'ipfs init'
	r, err := fsrepo.Open("~/.ipfs")
	if err != nil {
		return nil, err
	}

	cfg := &core.BuildCfg{
		Repo:   r,
		Online: true,
	}

	return core.NewNode(context.Background(), cfg)
}

func main() {
	flag.Parse()

	nd, err := setupIpfs()
	if err != nil {
		log.Error(err)
		return
	}

	log.Debugf("I am peer %s", nd.Identity)

	log.Debugf("Fetching articles from: %s", *url)

	pollFeed(*url, 5, charsetReader)
}
