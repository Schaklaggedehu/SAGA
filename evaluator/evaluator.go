package evaluator

import (
	"SAGA_Crawler/emailer"
	"SAGA_Crawler/automater"
	"SAGA_Crawler/logger"
	"SAGA_Crawler/resourcer"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"googlemaps.github.io/maps"
	"log"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var f = fmt.Println
var p = log.Println

func Process(urls []string) {
	resourcer.InitConditionData()
	var rentals []resourcer.RentalProperty
	for _, url := range urls {

		res, err := http.Get(url)
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
		rentals = append(rentals, findParameters(doc, url))
	}

	processHouseData(rentals)
}

func findParameters(doc *goquery.Document, url string) resourcer.RentalProperty {
	rental := resourcer.RentalProperty{}
	rental.Conditions = resourcer.Conditions
	rental.Url = url
	texts := map[string]string{}
	contentEle := doc.Find(".main-content-left")
	contentEle.Find("*").Each(createTextMap(&texts))
	rental.InfoTexts = texts
	titleEle := contentEle.Find("H1")
	rental.Title = titleEle.Text()
	descriptionEle := doc.Find(".col-pad-big-b")
	addressEle := descriptionEle.Find("p")
	data := map[string]string{}
	dataEle := descriptionEle.Find("dl>*")
	dataEle.Each(createDataMap(&data))
	rental.Data = data
	exp, _ := regexp.Compile(`  +`)
	trimmedAddress := exp.ReplaceAllString(addressEle.Text(), " ")

	longAddress := strings.Replace(trimmedAddress, "\n", "", -1)
	addressExp, _ := regexp.Compile(`\S+\.* *[\S]+ \d{5} \w*`)
	address := addressExp.FindString(longAddress)
	if address == "" {
		logger.Log("Regex not working for: " + longAddress)
	}
	rental.Address = address
	return rental
}

func processHouseData(rentals []resourcer.RentalProperty) {

	var requestExposeUrls []string
	f("Checking rental quality...")
	for i := range rentals {
		rentals[i].Red = false
		rentals[i].Address1 = resourcer.PersonalInfo.Address1
		rentals[i].Address2 = resourcer.PersonalInfo.Address2
		route1 := getDistance(rentals[i].Address, rentals[i].Address1)
		route2 := getDistance(rentals[i].Address, rentals[i].Address2)
		rentals[i].Distance1 = int(route1.TotalDuration / 1000000000 / 60)
		rentals[i].Distance2 = int(route2.TotalDuration / 1000000000 / 60)

		routeSteps1 := strings.Split(route1.toSteps(), "RB")
		if len(routeSteps1) > 1 {
			routeSteps1 = append(routeSteps1, "RB")
		} else {
			routeSteps1 = append(routeSteps1, "", "")
		}
		rentals[i].Journey1 = routeSteps1
		routeSteps2 := strings.Split(route2.toSteps(), "RB")
		if len(routeSteps2) > 1 {
			routeSteps2 = append(routeSteps2, "RB")
		} else {
			routeSteps2 = append(routeSteps2, "", "")
		}
		rentals[i].Journey2 = routeSteps2
		for key, info := range rentals[i].Data {
			switch key {
			case "Zimmer":
				rentals[i].Rooms, _ = strconv.Atoi(info[:1])
			case "Gesamtmiete":
				rentals[i].Rent = info
			case "Wohnfläche ca.":
				rentals[i].Size = info
			case "Etage":
				rentals[i].Floor = info
			case "Verfügbar ab":
				rentals[i].AvailableFrom = info
			case "Besichtigung":
				rentals[i].Viewing = info
			default:
				logger.Log("Unhandeled parameter: " + info)
			}
		}
		price, err := strconv.ParseFloat(euroToNumber(rentals[i].Rent), 64)
		if err != nil {
			f(err)
		}
		size, err := strconv.ParseFloat(sqmToNumber(rentals[i].Size), 64)
		if err != nil {
			f(err)
		}
		pricePerSQMString := fmt.Sprintf("%.2f €/m²", price/size)
		pricePerSQMString = strings.Replace(pricePerSQMString, ".", ",", 1)
		rentals[i].PricePerSQM = pricePerSQMString

		rentals[i].InfoTexts["Titel"] = rentals[i].Title

		goodwords := resourcer.PersonalInfo.KeywordsGood
		getKeywords(rentals, i, goodwords, true)
		badwords := resourcer.PersonalInfo.KeywordsBad
		getKeywords(rentals, i, badwords, false)

		if rentals[i].Distance1 <= resourcer.Conditions.MaxCommute {
			rentals[i].Distance1Status[0] = "green"
			rentals[i].Distance1Status[1] = "✔"
		} else if float64(rentals[i].Distance1) <= float64(resourcer.Conditions.MaxCommute)*1.1 {
			rentals[i].Distance1Status[0] = "yellow"
			rentals[i].Distance1Status[1] = "✔"
		} else {
			rentals[i].Distance1Status[0] = "red"
			rentals[i].Distance1Status[1] = "✘"
			rentals[i].Red = true
		}
		if rentals[i].Distance2 <= resourcer.Conditions.MaxCommute {
			rentals[i].Distance2Status[0] = "green"
			rentals[i].Distance2Status[1] = "✔"
		} else if float64(rentals[i].Distance2) <= float64(resourcer.Conditions.MaxCommute)*1.1 {
			rentals[i].Distance2Status[0] = "yellow"
			rentals[i].Distance2Status[1] = "✔"
		} else {
			rentals[i].Distance2Status[0] = "red"
			rentals[i].Distance2Status[1] = "✘"
			rentals[i].Red = true
		}
		if rentals[i].Rooms >= resourcer.Conditions.MinRooms {
			rentals[i].RoomsStatus[0] = "green"
			rentals[i].RoomsStatus[1] = "✔"
		} else {
			rentals[i].RoomsStatus[0] = "red"
			rentals[i].RoomsStatus[1] = "✘"
			rentals[i].Red = true
		}
		if size >= float64(resourcer.Conditions.MinSize) {
			rentals[i].SizeStatus[0] = "green"
			rentals[i].SizeStatus[1] = "✔"
		} else if size >= float64(resourcer.Conditions.MinSize)*0.9 {
			rentals[i].SizeStatus[0] = "yellow"
			rentals[i].SizeStatus[1] = "✔"
		} else {
			rentals[i].SizeStatus[0] = "red"
			rentals[i].SizeStatus[1] = "✘"
			rentals[i].Red = true
		}
		if price <= float64(resourcer.Conditions.MaxRent) {
			rentals[i].RentStatus[0] = "green"
			rentals[i].RentStatus[1] = "✔"
		} else if price <= float64(resourcer.Conditions.MaxRent)*1.1 {
			rentals[i].RentStatus[0] = "yellow"
			rentals[i].RentStatus[1] = "✔"
		} else {
			rentals[i].RentStatus[0] = "red"
			rentals[i].RentStatus[1] = "✘"
			rentals[i].Red = true
		}
		if !rentals[i].Red || resourcer.DEBUG {
			requestExposeUrls = append(requestExposeUrls, rentals[i].Url)
		}
	}
	emailer.SendResultMail(rentals)
	if len(requestExposeUrls) > 0 && resourcer.PersonalInfo.SendResume {
		automater.RequestExposes(requestExposeUrls)
	}
}

func getKeywords(rentals []resourcer.RentalProperty, i int, words []string, good bool) {
	badWords := strings.Join(words, "|")
	keywords, _ := regexp.Compile(`(?i)` + badWords)
	for _, content := range rentals[i].InfoTexts {
		words := keywords.FindAllString(content, -1)
		for _, keyword := range words {
			if !good {
				rentals[i].Red = true
			}
			area, _ := regexp.Compile(`(?i).{0,20}` + keyword + `.{0,20}`)
			word, _ := regexp.Compile(`(?i)` + keyword)
			areaVal := "..." + area.FindString(content) + "..."
			wordVal := word.FindString(content)
			value := strings.Split(areaVal, wordVal)
			wordVal = strings.Replace(wordVal, ".", "", -1) //quickfix because otherwise wilhelm.tel becomes a href
			value = append(value, wordVal)
			if good {
				rentals[i].InfoGood = append(rentals[i].InfoGood, value)
			} else {
				rentals[i].InfoBad = append(rentals[i].InfoBad, value)
			}
		}

	}
}

func euroToNumber(old string) string {
	i := strings.Index(old, ",")
	s := strings.Replace(old, " €", "", 1)
	s = s[:i]
	return strings.Replace(s, ".", "", 1)
}
func sqmToNumber(old string) string {
	return strings.Replace(old, " m²", "", 1)
}
func getDistance(origin string, destination string) route {
	c, err := maps.NewClient(maps.WithAPIKey("AIzaSyDUJC_TJLx749V7WHVjpXEKCG12_HG3rtQ"))
	if err != nil {
		log.Fatalf("fatal error: %s", err)
	}
	now := time.Now()
	date := time.Date(now.Year(), now.Month(), now.Day(), 8, 0, 0, 0, time.Local)
	r := &maps.DirectionsRequest{
		Origin:        origin,
		Destination:   destination,
		Mode:          maps.TravelModeTransit,
		DepartureTime: strconv.FormatInt(date.Unix(), 10),
		Units:         maps.UnitsMetric,
		Language:      "de",
	}
	gmRoute, _, err := c.Directions(context.Background(), r)
	if err != nil {
		f(err)
	}

	route := route{}
	route.TotalDuration = gmRoute[0].Legs[0].Duration.Nanoseconds()

	for s := range gmRoute[0].Legs[0].Steps {
		gmStep := gmRoute[0].Legs[0].Steps[s]
		step := step{}

		step.TravelMode = gmStep.TravelMode
		step.Duration = gmStep.Duration.String()
		if gmStep.TravelMode != "WALKING" {
			step.ShortName = gmStep.TransitDetails.Line.ShortName
		}
		route.Steps = append(route.Steps, step)
	}
	return route
}

type step struct {
	TravelMode string
	Duration   string
	ShortName  string
}

type route struct {
	TotalDuration int64
	Steps         []step
}

func (route *route) toSteps() string {
	routeString := ""
	for _, step := range route.Steps {
		if routeString != "" {
			routeString += " -> "
		}
		if step.TravelMode != "WALKING" {
			routeString += step.ShortName + " " + step.Duration
		} else {
			routeString += " Zu Fuß " + step.Duration
		}
	}
	return routeString
}

func createDataMap(d *map[string]string) func(int, *goquery.Selection) {
	data := *d
	var key string
	var keyUsed = true
	return func(i int, s *goquery.Selection) {
		if s.Text() != "Netto-Kalt-Miete" &&
			s.Text() != "Objektnummer" &&
			s.Text() != "Betriebskosten" &&
			s.Text() != "Sonstige Kosten" &&
			s.Text() != "Terrasse" &&
			s.Text() != "Balkon" &&
			s.Text() != "Heizkosten" {
			if i%2 == 0 {
				key = s.Text()
				keyUsed = false
			} else if !keyUsed {
				data[key] = s.Text()
				keyUsed = true
			}
		}
	}
}

func createTextMap(d *map[string]string) func(int, *goquery.Selection) {
	texts := *d
	var key string
	var keyUsed = true
	return func(i int, s *goquery.Selection) {
		if s.HasClass("h4") && s.Text() != "Downloads" {
			key = s.Text()
			keyUsed = false
		} else if !keyUsed {
			texts[key] = s.Text()
			keyUsed = true
		}
	}
}
