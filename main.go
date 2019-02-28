package main

import (
	"SAGA_Crawler/evaluator"
	"SAGA_Crawler/resourcer"
	"bufio"
	"encoding/csv"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

var searchUrl = "https://www.saga.hamburg/immobiliensuche?type=wohnungen"
var basicUrl = "https://www.saga.hamburg"

var f = fmt.Println
var p = log.Println

func main() {
	if resourcer.DEBUG {
		log.SetFlags(log.LstdFlags | log.Lshortfile)
		p("DEBUGGING! Change at resourcer.DEBUG")
	} else {
		log.SetOutput(ioutil.Discard)
	}
	resourcer.InitConfigData()

	runCrawler()
	if !resourcer.DEBUG {
		for range time.Tick(time.Minute * time.Duration(resourcer.PersonalInfo.Frequency)) {
			runCrawler()
		}
	}
}

func runCrawler() {
	f("fetching...")

	res, err := http.Get(searchUrl)
	if err != nil {
		f(err)
		return
	}
	if res.StatusCode != 200 {
		f(err)
		return
	}
	// Load the HTML document
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		f(err)
		return
	}
	err = res.Body.Close()
	if err != nil {
		f(err)
		return
	}
	findUrls(doc)
}

func findUrls(doc *goquery.Document) {
	var newUrls, allUrls []string
	doc.Find(".media").Each(eval(&newUrls, &allUrls))
	if len(newUrls) > 0 {
		evaluator.Process(newUrls)
	}
	cleanFile(allUrls)
	f("Done!")
	f("Next check in " + strconv.Itoa(resourcer.PersonalInfo.Frequency) + " minutes\n")
}

func eval(newUrls *[]string, allUrls *[]string) func(i int, s *goquery.Selection) {
	return func(i int, s *goquery.Selection) {
		href, exists := s.Attr("href")
		if exists {
			url := basicUrl + href
				*allUrls = append(*allUrls, url)
			if urlNew(url) || resourcer.DEBUG {
				f(url)
				*newUrls = append(*newUrls, url)
				saveUrl(url)
			}
		}
	}
}

func urlNew(url string) bool {
	csvFile, err := os.Open("SAGA_Crawler_settings/urls.csv")
	if err != nil {
		csvFile, err = os.Create("SAGA_Crawler_settings/urls.csv")
		if err != nil {
			f(err)
		}
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))
	for {
		line, err := reader.Read()
		if err == io.EOF || err != nil {
			break
		} else {
			if line[0] == url {
				return false
			}

		}
	}
	return true
}

func saveUrl(url string) {
	csvFile, err := os.OpenFile("SAGA_Crawler_settings/urls.csv", os.O_APPEND|os.O_WRONLY, 0777)
	if err != nil {
		csvFile, err = os.Create("SAGA_Crawler_settings/urls.csv")
		if err != nil {
			f(err)
		}
	}
	err = csv.NewWriter(csvFile).WriteAll([][]string{{url}})
	if err != nil {
		f(err)
	}
	err = csvFile.Close()
	if err != nil {
		f(err)
	}
}

func cleanFile(urls []string) {
	csvFile, err := os.Open("SAGA_Crawler_settings/urls.csv")
	if err != nil {
		csvFile, err = os.Create("SAGA_Crawler_settings/urls.csv")
		if err != nil {
			f(err)
		}
	}
	reader := csv.NewReader(bufio.NewReader(csvFile))

	var relevantUrls []string
	for {
		line, err := reader.Read()
		if err == io.EOF || err != nil {
			break
		} else {
			if contains(urls, line[0]) && !contains(relevantUrls, line[0]) {
				relevantUrls = append(relevantUrls, line[0])
			}
		}
	}
	if err != nil {
		f(err)
	}
	if len(relevantUrls) > 0 {

		csvFile, err = os.OpenFile("SAGA_Crawler_settings/urls.csv", os.O_TRUNC|os.O_WRONLY, 0777)
		if err != nil {
			f(err)
		}
		err = csvFile.Close()
		if err != nil {
			f(err)
		}
		for _, url := range relevantUrls {
			saveUrl(url)
		}
	}

}
func contains(s []string, e string) bool {
	for _, a := range s {
		if a == e {
			return true
		}
	}
	return false
}
