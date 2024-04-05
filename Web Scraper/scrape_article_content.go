package main

import (
	"context"
	"encoding/csv"
	"fmt"
	"os"
	"strings"

	"github.com/chromedp/chromedp"
	"golang.org/x/net/html"
)

func extractTextFromPTag(tokenizer *html.Tokenizer) string {
		 var extractedText strings.Builder

    for {
        tokenType := tokenizer.Next()

        switch tokenType {
        case html.ErrorToken:
            return extractedText.String() 

        case html.TextToken:
            extractedText.WriteString(tokenizer.Token().Data)

        case html.EndTagToken:
            tagName, _ := tokenizer.TagName()
            if string(tagName) == "p" {
                return extractedText.String() // End of <p> tag
            }
        }
    }
}

func scrapePTags(url string) string {
	var articleContent strings.Builder

	context, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var htmlContent string
	loadingErr := chromedp.Run(context, chromedp.Navigate(url), chromedp.OuterHTML(`html`, &htmlContent))

	if loadingErr != nil{
		fmt.Println("There was an error fetching the url: ", loadingErr)
	}

	tokenizer := html.NewTokenizer(strings.NewReader(htmlContent))

	for {
			tokenType := tokenizer.Next()

			switch tokenType {
			case html.ErrorToken:
					err := tokenizer.Err()
					if err != nil {
							fmt.Println("Tokenization error:", err)
							return articleContent.String()
					}

			case html.StartTagToken:
					tagName, _ := tokenizer.TagName()
					if string(tagName) == "p" {
							articleContent.WriteString(extractTextFromPTag(tokenizer))
					}

			}
	}
}

func ScrapeArticleContent() {
	csvFile, err := os.OpenFile("results.csv", os.O_RDWR, 0644)
	if err != nil{
		fmt.Println("error opening csv: ", err)
		return
	}
	defer csvFile.Close()

	csvReader := csv.NewReader(csvFile)
	records, err := csvReader.ReadAll()
	if err != nil{
		fmt.Println("Error reading data: ", err)
		return
	}

	csvWriter := csv.NewWriter(csvFile)
	defer csvWriter.Flush()

	headers := records[0]
	headers = append(headers, "Article Content")

	_, err = csvFile.Seek(0,0)
	if err != nil {
		fmt.Println("Error seeking file: ", err)
	}

	err = csvWriter.Write(headers)
	if err != nil{
		fmt.Println("Error writing new header: ", err)
	}

	for i, record := range records{
		if i != 0{
			url := record[3]
			fmt.Println("starting to scrape url no. :", i)
			articleContent := scrapePTags(url)
			fmt.Println("finished scraping url no. :", i, "\n")

			record = append(record, articleContent)
			csvWriter.Write(record)
		}
	}

	fmt.Println("CSV updated successfully.")
}