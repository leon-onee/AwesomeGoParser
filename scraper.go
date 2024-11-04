package main

import (
	"encoding/csv"
	"fmt"
	"log"
	"net/http"
	"os"
	"reflect"

	"github.com/PuerkitoBio/goquery"
)

type Library struct {
	url, title, desc, rating string
}

func getHTMLDocument(url string) (*http.Response, error) {
	log.Printf("Fetching URL: %s\n", url)
	res, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to the target page: %w", err)
	}
	if res.StatusCode != 200 {
		defer res.Body.Close()
		return nil, fmt.Errorf("HTTP Error %d: %s", res.StatusCode, res.Status)
	}

	return res, nil
}

func fetchAndParseDocument(url string) (*goquery.Document, error) {
	res, err := getHTMLDocument(url)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse the HTML document: %w", err)
	}

	return doc, nil
}

func parseHTMLDocument(doc *goquery.Document) ([]Library, error) {
	var libraries []Library
	count := 0

	doc.Find(`.markdown-body ul li a[href^="https://github.com/"]`).Each(func(i int, s *goquery.Selection) {
		url, _ := s.Attr("href")
		rating, err := getLibraryRating(url)
		if err != nil {
			log.Printf("Failed to get rating for %s: %v", url, err)
			rating = "N/A"
		}
		library := Library{
			url:    url,
			title:  s.Text(),
			desc:   s.Parent().Text(),
			rating: rating,
		}
		libraries = append(libraries, library)
		count++
	})

	return libraries, nil
}

func getLibraryRating(url string) (string, error) {
	doc, err := fetchAndParseDocument(url)
	if err != nil {
		return "", err
	}

	rating := doc.Find("#repo-stars-counter-star").Text()
	return rating, nil
}

func writeCSV(filename string, libraries []Library) error {
	file, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("failed to create the output CSV file: %w", err)
	}
	defer file.Close()

	writer := csv.NewWriter(file)
	defer writer.Flush()

	val := reflect.ValueOf(libraries[0])
	headers := make([]string, val.NumField())
	for i := 0; i < val.NumField(); i++ {
		headers[i] = val.Type().Field(i).Name
	}

	if err := writer.Write(headers); err != nil {
		return fmt.Errorf("failed to write CSV headers: %w", err)
	}

	for _, library := range libraries {
		record := []string{
			library.url,
			library.title,
			library.desc,
			library.rating,
		}
		if err := writer.Write(record); err != nil {
			return fmt.Errorf("failed to write CSV record: %w", err)
		}
	}

	return nil
}

func main() {

	// download the target HTML document
	doc, err := fetchAndParseDocument("https://github.com/avelino/awesome-go")
	if err != nil {
		log.Fatal(err)
	}

	libraries, err := parseHTMLDocument(doc)
	if err != nil {
		log.Fatal(err)
	}

	if err := writeCSV("libraries.csv", libraries); err != nil {
		log.Fatal(err)
	}

	log.Println("CSV file has been created successfully.")
}

// 2024/06/27 19:29:16 CSV file has been created successfully.
// парсил минут 25, спарсилось 2513 строк
