package main

import (
	"fmt"
	"log"

	"github.com/PuerkitoBio/goquery"
	"github.com/nkanaev/yarr/src/httpclient"
)

func main() {

	url := "https://codemls.com/articles"
	client := httpclient.NewClient()
	res, _ := client.Get(url)
	defer res.Body.Close()
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".gap-6 a.rounded-lg").Each(func(i int, s *goquery.Selection) {
		// fmt.Println(s.Html())
		title := s.Find("h2").Text()
		link := s.AttrOr("href", "")
		date := s.Find("span").Text()
		fmt.Printf("---------- %s: %s: %s\n", title, link, date)
	})
}
