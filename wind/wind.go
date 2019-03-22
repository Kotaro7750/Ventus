package wind

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

// ForecastData is a data of forecast of specific date
type ForecastData struct {
	Date               string
	WindSpeedMidNight  string
	WindSpeedMorning   string
	WindSpeedAfternoon string
	WindSpeedEvening   string
}

//MakeForecastData returns array of ForecastData
func MakeForecastData(url string, filePath string) []ForecastData {
	saveForecastPage(url, filePath)
	formateForecastPage(filePath)

	var forecastDatas = []ForecastData{}

	fp, err := os.Open(filePath)
	defer fp.Close()

	if err != nil {
		fmt.Printf("failes to open %s", filePath)
	}

	scanner := bufio.NewScanner(fp)
	for scanner.Scan() {
		text := scanner.Text()
		forecastDataArray := strings.Fields(text)

		date := forecastDataArray[0]
		windSpeedMidNight := forecastDataArray[2]
		windSpeedMorning := forecastDataArray[4]
		windSpeedAfternoon := forecastDataArray[6]
		windSpeedNight := forecastDataArray[8]

		forecastDatas = append(forecastDatas, ForecastData{
			Date:               date,
			WindSpeedMidNight:  windSpeedMidNight,
			WindSpeedMorning:   windSpeedMorning,
			WindSpeedAfternoon: windSpeedAfternoon,
			WindSpeedEvening:   windSpeedNight,
		})
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("failes to scan %s", filePath)
	}

	return forecastDatas
}

func saveForecastPage(url string, filePath string) {
	doc, err := goquery.NewDocument(url)
	if err != nil {
		fmt.Printf("faled to get %s", url)
	}

	body := ""
	doc.Find("body > #main-column > .section-wrap > .forecast-point-10days > tbody").Each(func(i int, s *goquery.Selection) {
		body += s.Text()
	})
	if err != nil {
		fmt.Print("failes to get DOM")
	}

	ioutil.WriteFile(filePath, []byte(body), os.ModePerm)
}

func formateForecastPage(filePath string) {
	fp, err := os.Open(filePath)
	defer fp.Close()

	if err != nil {
		fmt.Printf("failes to open %s", filePath)
	}

	scanner := bufio.NewScanner(fp)
	body := ""
	for scanner.Scan() {
		text := scanner.Text()
		text = strings.TrimSpace(text)
		isRequired, linefeed := isRequired(text)
		if isRequired {
			if linefeed {
				text = "\n" + text
			}
			body += text + " "
		}
	}
	body += "\n"
	body = body[1:]

	if err := scanner.Err(); err != nil {
		fmt.Printf("failes to scan %s", filePath)
	}
	ioutil.WriteFile(filePath, []byte(body), os.ModePerm)
}

func isRequired(text string) (bool, linefeed bool) {
	if utf8.RuneCountInString(text) == 0 {
		return false, false
	}
	targetRegExp := regexp.MustCompile(`日付|天気|気温|降水確率|降水量|降水量|湿度|風`)
	if targetRegExp.MatchString(text) {
		return false, false
	}

	targetRegExp = regexp.MustCompile(`(^[0-9].月[0-9].日)`)
	if targetRegExp.MatchString(text) == true {
		return true, true
	}

	targetRegExp = regexp.MustCompile(`(m/s)|([0-9].:[0-9])`)
	if targetRegExp.MatchString(text) == false {
		return false, false
	}

	return true, false
}
