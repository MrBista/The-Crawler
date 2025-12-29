package helper

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/MrBista/The-Crawler/internal/dto"
	"github.com/PuerkitoBio/goquery"
)

func ProcessCrawl(job dto.CrawlJob) {
	log.Printf("[Worker] starting to crawl for: %s", job.URL)

	if !strings.HasPrefix(job.URL, "http") {
		log.Printf("[Worker] Invalid url schema: %s", job.URL)
		return
	}

	client := &http.Client{
		Timeout: 15 * time.Second,
	}

	req, err := http.NewRequest(http.MethodGet, job.URL, nil)
	if err != nil {
		log.Printf("[Worker] Failed to create request %v", job.URL)
		return
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/120.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9,id;q=0.8")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		log.Printf("[Worker] failed to fetch url %s", job.URL)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		log.Printf("[Worker] Non-200 status code: %d", resp.StatusCode)
		return
	}

	doc, err := goquery.NewDocumentFromReader(resp.Body)
	if err != nil {
		log.Printf("[Worker] Failed to parse HTML: %v", err)
		return
	}

	html, err := doc.Find("body").Html()
	if err != nil {
		log.Printf("[Worker] Failed to render HTML: %v", err)
		return
	}

	saveRawHTml(job.ID, []byte(html))

	title := doc.Find("title").Text()

	fmt.Println("---------------------------------------------------")
	fmt.Printf("✅ SUCCESS CRAWL\n")
	fmt.Printf("ID    : %s\n", job.ID)
	fmt.Printf("URL   : %s\n", job.URL)
	fmt.Printf("TITLE : %s\n", strings.TrimSpace(title))
	fmt.Println("---------------------------------------------------")

	// doc.Find("script, style, noscript, iframe, embed, object").Remove()

	// doc.Find("*").Each(func(i int, s *goquery.Selection) {
	// 	s.RemoveAttr("onclick")
	// 	s.RemoveAttr("onload")
	// 	s.RemoveAttr("onerror")
	// })

	// html, err := doc.Find("body").Html()
	// if err != nil {
	// 	log.Printf("[Worker] Failed to render HTML: %v", err)
	// 	return
	// }

	// generatePDF(html, job.ID)

	/*
		if job.Depth > 1 {
			doc.Find("a").Each(func(i int, s *goquery.Selection) {
				href, exists := s.Attr("href")
				if exists {
					childJob := dto.CrawlJob{
						ID:    uuid.New().String(),
						URL:   fixUrl(href, job.URL),
						Depth: job.Depth - 1,
					}

					// !TODO use method handler

				}
			})
		}
	*/
}

func fixUrl(hrefUrl, parentUrl string) string {
	return ""
}
