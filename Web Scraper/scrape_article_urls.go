package main

import (
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
)

type searchResult struct {
	date string
	publisher string
	title string
	url string
}

func findNextBtn(node *html.Node, isValidURL *bool, url *string) {
	if node == nil{
		return
	}
	
	attributeDictionary := make(map[string]string)
	for _, attribute := range node.Attr{
		attributeDictionary[attribute.Key] = attribute.Val
	}

	if node.Type == html.ElementNode && node.Data == "a"{
		if label, ok := attributeDictionary["aria-label"]; ok && label == "Next page"{
			*url = attributeDictionary["href"]
			*isValidURL = true
			return
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling{
		findNextBtn(child, isValidURL, url)
	}
}

func findText(node *html.Node) ([]string, []string, []string, []string) {
	var publishers []string
	var titles []string
	var articleUrls []string
	var dates []string

	attributeDictionary := make(map[string]string)
	for _, attribute := range node.Attr{
		attributeDictionary[attribute.Key] = attribute.Val
	}

	if node.Type == html.ElementNode && node.Data == "a"{
		if _, ok := attributeDictionary["data-ved"]; ok{
			articleUrls = append(articleUrls, attributeDictionary["href"])
		}
	} else if node.Type == html.ElementNode && node.Data == "div"{
		if classNames, ok := attributeDictionary["class"]; ok && classNames == "BNeawe UPmit AP7Wnd lRVwie"{
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.TextNode {
					publishers = append(publishers, child.Data)
				}
			}
			} else if classNames, ok := attributeDictionary["class"]; ok && classNames == "BNeawe vvjwJb AP7Wnd"{
				for child := node.FirstChild; child != nil; child = child.NextSibling {
					if child.Type == html.TextNode {
					titles = append(titles, child.Data)
				}
			}
		}
	} else if node.Type == html.ElementNode && node.Data == "span"{
		if classNames, ok := attributeDictionary["class"]; ok && classNames == "r0bn4c rQMQod"{
			for child := node.FirstChild; child != nil; child = child.NextSibling {
				if child.Type == html.TextNode {
					dates = append(dates, child.Data)
				}
			}
		}
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling{
		publisherValues, titleValues, articleUrlValues, dateValues := findText(child)
		
		publishers = append(publishers, publisherValues...)
		titles = append(titles, titleValues...)
		articleUrls = append(articleUrls, articleUrlValues...)
		dates = append(dates, dateValues...)
	}

	return publishers, titles, articleUrls, dates
}

func scrapeAndFollow(url string) []searchResult {
	var allExtractedValues []searchResult

	response, fetchErr := http.Get(url)
	if fetchErr != nil{
		log.Fatal(fetchErr)
	}
	defer response.Body.Close()

	body, readErr := io.ReadAll(response.Body)
	if readErr != nil{
		log.Fatal(readErr)
	}

	document, parsingErr := html.Parse(strings.NewReader(string(body)))
	if parsingErr != nil{
		log.Fatal(parsingErr)
	}

	publishers, titles, articleURls, dates := findText(document)

	for i:=len(publishers)-1; i >= 0; i-- {
		articleUrl := strings.TrimPrefix(articleURls[i], "/url?q=")
		articleUrl = strings.Split(articleUrl, "&sa=U&")[0] + " "
		
		if publishers[i] != "YouTube" {
			result := searchResult{
				date: dates[i],
				publisher: publishers[i],
				title: titles[i],
				url: articleUrl,
			}

			allExtractedValues = append(allExtractedValues, result)
		}
	}

	var isValidURL bool = false
	var nextUrl string = ""

	findNextBtn(document, &isValidURL, &nextUrl)

	if isValidURL {
		extractedValues := scrapeAndFollow("https://www.google.com" + nextUrl)
		allExtractedValues = append(allExtractedValues, extractedValues...)
	}

	return allExtractedValues
}

func ScrapeArticleURls() {
	url := "https://www.google.com/search?q=simplygo+singapore&sca_esv=72c8a1a5d0cf3c73&tbs=sbd:1&tbm=nws&prmd=insmvbtz&sxsrf=ACQVn088Ww94FpKrKItP5Rsp1nzqSgZ6mg:1712028591321&ei=r3sLZq6cE8mM4-EPiKuq6AU&start=0&sa=N&ved=2ahUKEwjur5_Ay6KFAxVJxjgGHYiVCl04RhDy0wN6BAgBEAQ&biw=1920&bih=957&dpr=1"

	scrapeResults := scrapeAndFollow(url)
	
	file, err := os.Create("results.csv")
    if err != nil {
        fmt.Println("Error creating CSV:", err)
        return
    }
    defer file.Close()

    writer := csv.NewWriter(file)
    defer writer.Flush()

    headers := []string{"Date", "Publisher", "Title", "URL"}
    err = writer.Write(headers)
    if err != nil {
        fmt.Println("Error writing header:", err)
        return
    }

    for _, result := range scrapeResults {
        row := []string{result.date, result.publisher, result.title, result.url}
        err = writer.Write(row)
        if err != nil {
            fmt.Println("Error writing row:", err)
            return
        }
    }

    fmt.Println("Search results saved to results.csv")
}