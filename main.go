package main

import (
	"fmt"
	rss "github.com/jteeuwen/go-pkg-rss"
	"net/http"
	"os"
	"regexp"
	"strings"
	"sync"
)

func main() {

	var (
		err   error
		feeds map[string]*rss.Feed
		river []*rss.Feed
	)

	feeds = map[string]*rss.Feed{
		"Fefe":     rss.New(2, true, chanHandler, itemHandler),
		"HN":       rss.New(2, true, chanHandler, itemHandler),
		"Slashdot": rss.New(2, true, chanHandler, itemHandler),
	}

	feeds["HN"].Url = "https://news.ycombinator.com/rss"
	feeds["Fefe"].Url = "https://blog.fefe.de/rss.xml?html"
	feeds["Slashdot"].Url = "http://rss.slashdot.org/Slashdot/slashdot"

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		var (
			s          string
			f          *rss.Feed
			c          *rss.Channel
			stripFuncs []func(string) string
		)

		s = HTMLHeader()
		river = FeedRiver(feeds)

		stripFuncs = []func(string) string{stripTags}

		for i := 0; i < len(river); i++ {
			f = river[i]

			if strings.Contains(f.Url, "slashdot") {
				stripFuncs = append(stripFuncs, stripSlashdot)
			} else if strings.Contains(f.Url, "ycombinator") {
				stripFuncs = append(stripFuncs, stripHN)
			}

			c = f.Channels[0]

			for k := 0; k < len(c.Items); k++ {
				s += "<h3>" + c.Items[k].Title + "</h3>\n"
				s += c.Items[k].Description + "\n"
				for f := 0; f < len(stripFuncs); f++ {
					s = stripFuncs[f](s)
				}
			}
		}
		s += HTMLFooter()

		ret, err := w.Write([]byte(s))
		if err != nil {
			fmt.Println("Return code %i, error %s", ret, err)
		}
	})

	err = http.ListenAndServe(":4242", nil)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func FeedRiver(feeds map[string]*rss.Feed) []*rss.Feed {

	var (
		w     sync.WaitGroup
		river []*rss.Feed
	)

	w.Add(len(feeds))

	for _, feed := range feeds {
		go func(f *rss.Feed) {
			f.Fetch(f.Url, nil)
			river = append(river, f)
			w.Done()
		}(feed)
	}

	w.Wait()
	return river
}

func chanHandler(feed *rss.Feed, newchannels []*rss.Channel) {
}

func itemHandler(feed *rss.Feed, ch *rss.Channel, newitems []*rss.Item) {
}

func HTMLHeader() string {
	return `<!DOCTYPE html>
<html lang="en">
<head>
<link href="http://netdna.bootstrapcdn.com/bootstrap/3.1.1/css/bootstrap.min.css" rel="stylesheet">
<meta name="viewport" content="width=device-width, initial-scale=1">
<meta http-equiv="Content-Type" content="text/html; charset=utf-8">
<link href='http://fonts.googleapis.com/css?family=Scada' rel='stylesheet' type='text/css'>
<style type="text/css">
body {
	font-size: 18px;
}
h1, h2, h3, h4, h5, h6 {
    font-family: 'Scada';
}
p, div {
    font-family: 'Scada';
}
</style>
</head>
<body>
<div class="container">`
}

func HTMLFooter() string {
	return `
</div>
</body>
</html>`
}

func stripTags(content string) string {
	var (
		re      *regexp.Regexp
		c       string
		pattern string
		tags    []string
	)

	tags = []string{"img", "br", "a", "iframe"}

	c = content
	for i := 0; i < len(tags); i++ {
		pattern = "(?i)</?" + tags[i] + "[^>]*>"
		re = regexp.MustCompile(pattern)
		c = re.ReplaceAllString(c, "")
	}
	return c
}

func stripHN(content string) string {
	return strings.Replace(content, "Comments", "", -1)
}

func stripSlashdot(content string) string {
	return strings.Replace(content, "Read more of this story at Slashdot.", "", -1)
}
