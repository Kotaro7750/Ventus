package wind

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/PuerkitoBio/goquery"
)

const (
	MidNight  = "午前3時"
	Morning   = "午前9時"
	Afternoon = "午後3時"
	Night     = "午後9時"
)

// ForecastData is a data of forecast of specific date
type ForecastData struct {
	Date               string
	WindSpeedMidNight  int
	WindSpeedMorning   int
	WindSpeedAfternoon int
	WindSpeedNight     int
}

type ForecastDatas []ForecastData

//MakeForecastData returns array of ForecastData
func MakeForecastData(url string, filePath string) ForecastDatas {
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
		var windSpeed = []string{}
		for i := 2; i <= 8; i += 2 {
			windSpeed = append(windSpeed, strings.Split(forecastDataArray[i], "m")[0])
		}

		windSpeedMidNight, _ := strconv.Atoi(windSpeed[0])
		windSpeedMorning, _ := strconv.Atoi(windSpeed[1])
		windSpeedAfternoon, _ := strconv.Atoi(windSpeed[2])
		windSpeedNight, _ := strconv.Atoi(windSpeed[3])

		forecastDatas = append(forecastDatas, ForecastData{
			Date:               date,
			WindSpeedMidNight:  windSpeedMidNight,
			WindSpeedMorning:   windSpeedMorning,
			WindSpeedAfternoon: windSpeedAfternoon,
			WindSpeedNight:     windSpeedNight,
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

// IsExceededLimit judges if windspeed exceed limit
func (forecastData *ForecastData) isExceededLimit(limit int) (bool, string) {
	isExceed := false
	res := ""
	if forecastData.WindSpeedMidNight >= limit {
		isExceed = true
		res += MidNight + "、"
	}
	if forecastData.WindSpeedMorning >= limit {
		isExceed = true
		res += Morning + "、"
	}
	if forecastData.WindSpeedAfternoon >= limit {
		isExceed = true
		res += Afternoon + "、"
	}
	if forecastData.WindSpeedNight >= limit {
		isExceed = true
		res += Night + "、"
	}

	if isExceed {
		return isExceed, forecastData.Date + res
	}
	return isExceed, ""
}

//MaxSpeed returns maxspeed and it's time
func (forecastData *ForecastData) maxSpeed() (int, string) {
	max := -1
	res := ""
	if forecastData.WindSpeedMidNight > max {
		max = forecastData.WindSpeedMidNight
		res = MidNight
	}
	if forecastData.WindSpeedMorning > max {
		max = forecastData.WindSpeedMorning
		res = Morning
	}
	if forecastData.WindSpeedAfternoon > max {
		max = forecastData.WindSpeedAfternoon
		res = Afternoon
	}
	if forecastData.WindSpeedNight > max {
		max = forecastData.WindSpeedNight
		res = Night
	}

	return max, res
}

// MakeWindReport returns windreport
func (forecastDatas ForecastDatas) MakeWindReport(limit int) string {
	forecastDataNum := len(forecastDatas)

	text := "この" + strconv.Itoa(forecastDataNum) + "日間の最大風速は"
	exceedLimit := ""
	max := -1
	maxDay := ""
	maxTime := ""

	for i := 0; i < forecastDataNum; i++ {
		forecastData := forecastDatas[i]
		if dayMax, res := forecastData.maxSpeed(); dayMax > max {
			max = dayMax
			maxDay = forecastData.Date
			maxTime = res
		}
		if isExceed, res := forecastData.isExceededLimit(limit); isExceed {
			exceedLimit += res
		}
	}

	text += maxDay + maxTime + "の" + strconv.Itoa(max) + "m/sだよ！\n" + strconv.Itoa(limit) + "m/sを超える日は"
	if exceedLimit != "" {
		text += exceedLimit + "だよ〜！"
	} else {
		text += "ありません！"
	}
	return text
}
