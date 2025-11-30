package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

func main() {

	str := `
	<html>
		<head></head>
		<body>
			<div class="left-content">
				<a href="/review/1">中文test</a>
				<a href="/review/2">中文test2</a>
			</div>
			<div class="left-content">
				<a>中文test2</a>
				<span class="info">&nbsp;&nbsp;
				<i class="mdui-icon material-icons info-icon"></i> 3,426&nbsp;&nbsp;            
				<i class="mdui-icon material-icons info-icon"></i> 2025-11-03&nbsp;&nbsp;            
				<i class="mdui-icon material-icons info-icon"></i>
				</span>
			</div>
		</body>
	</html>
	`

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(str))
	if err != nil {
		log.Fatal(err)
	}

	doc.Find(".left-content").Each(func(i int, s *goquery.Selection) {
		title := s.Find("a").Text()
		link, _ := s.Find("a").Eq(1).Attr("href")
		date := s.Find("span").Text()
		fmt.Printf("%s: %s: %s\n", title, link, date)
	})
}
