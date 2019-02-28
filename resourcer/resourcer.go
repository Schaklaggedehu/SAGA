package resourcer

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
)

//const DEBUG = true
const DEBUG = false

type RentalProperty struct {
	Title           string
	Url             string
	Address         string
	Address1        string
	Address2        string
	Distance1       int
	Distance2       int
	Journey1        []string
	Journey2        []string
	Data            map[string]string
	InfoTexts       map[string]string
	Rooms           int
	Rent            string
	Size            string
	Floor           string
	AvailableFrom   string
	Viewing         string
	PricePerSQM     string
	InfoGood        [][]string
	InfoBad         [][]string
	StatusSymbol    map[string]string
	Distance1Status [2]string
	Distance2Status [2]string
	RoomsStatus     [2]string
	SizeStatus      [2]string
	RentStatus      [2]string
	Red             bool
	Conditions      conditionData
}
type configInfo struct {
	From         string   `json:"from"`
	To           string   `json:"to"`
	Password     string   `json:"password"`
	Server       string   `json:"server"`
	Port         int      `json:"port"`
	Frequency    int      `json:"frequency"`
	Address1     string   `json:"address1"`
	Address2     string   `json:"address2"`
	KeywordsBad  []string `json:"keywordsBad"`
	KeywordsGood []string `json:"keywordsGood"`
	SendResume   bool     `json:"autoRequestResume"`
}

var PersonalInfo configInfo

type conditionData struct {
	MaxRent    int `json:"maxRent"`
	MinRooms   int `json:"minRooms"`
	MinSize    int `json:"minSize"`
	MinFloor   int `json:"minFloor"`
	MaxCommute int `json:"maxCommute"`
}

var Conditions conditionData

var f = fmt.Println
var p = log.Println

func InitConfigData() {
	_ = os.Mkdir("SAGA_Crawler_settings", os.ModePerm)
	config, err := os.Open("SAGA_Crawler_settings/config.json")
	if err != nil {
		config, err = os.Create("SAGA_Crawler_settings/config.json")
		if err != nil {
			f(err)
		}
		path, err := filepath.Abs(filepath.Dir(config.Name()))
		if err != nil {
			f(err)
		}
		init := configInfo{}
		init.From = ""
		init.To = ""
		init.Password = ""
		init.Server = "smtp.gmail.com"
		init.Port = 587
		init.Frequency = 10
		init.Address1 = ""
		init.KeywordsBad = []string{}
		init.KeywordsGood = []string{}
		init.SendResume = false
		jsonString, err := json.MarshalIndent(init, "", "  ")
		err = ioutil.WriteFile("SAGA_Crawler_settings/config.json", jsonString, 0644)
		if err != nil {
			f(err)
		}
		err = config.Close()
		if err != nil {
			f(err)
		}
		f("Data missing")
		f("Please edit configInfo file")
		f(path + "/" + config.Name())
		os.Exit(1)
	}
	content, err := ioutil.ReadFile("SAGA_Crawler_settings/config.json")
	if err != nil {
		f(err)
	}
	PersonalInfo = configInfo{}
	err = json.Unmarshal(content, &PersonalInfo)
	if err != nil {
		f(err)
	}
	if PersonalInfo.From == "" ||
		PersonalInfo.To == "" ||
		PersonalInfo.Password == "" ||
		PersonalInfo.Server == "" ||
		PersonalInfo.Address1 == "" ||
		PersonalInfo.Port == 0 {
		f("Data missing")
		f("Please edit configInfo file")
		path, err := filepath.Abs(filepath.Dir(config.Name()))
		if err != nil {
			f(err)
		}
		f(path + "/" + config.Name())
		os.Exit(1)
	}
	if PersonalInfo.Frequency < 3 {
		PersonalInfo.Frequency = 3
	}
}

func InitConditionData() {
	config, err := os.Open("SAGA_Crawler_settings/conditions.json")
	if err != nil {
		config, err = os.Create("SAGA_Crawler_settings/conditions.json")
		if err != nil {
			f(err)
		}
		init := conditionData{}
		init.MaxCommute = 60
		init.MaxRent = 1000
		init.MinFloor = 0
		init.MinRooms = 1
		init.MinSize = 30
		jsonString, err := json.MarshalIndent(init, "", "  ")
		err = ioutil.WriteFile("SAGA_Crawler_settings/conditions.json", jsonString, 0644)
		if err != nil {
			f(err)
		}
		err = config.Close()
		if err != nil {
			f(err)
		}
	}
	content, err := ioutil.ReadFile("SAGA_Crawler_settings/conditions.json")
	if err != nil {
		f(err)
	}
	Conditions = conditionData{}
	err = json.Unmarshal(content, &Conditions)
	if err != nil {
		f(err)
	}
}
