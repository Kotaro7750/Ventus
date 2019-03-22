package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"

	//"github.com/Kotaro7750/Ventus/wind"
	"./wind"
	"github.com/kelseyhightower/envconfig"
	"github.com/nlopes/slack"
)

const (
	forecastURL      = "https://tenki.jp/forecast/3/16/4410/13110/10days.html"
	forecastFilePath = "./tmp.txt"
	limitSpeed       = 10
)

// interactionHandler handles interactive message response.
type interactionHandler struct {
	slackClient       *slack.Client
	verificationToken string
}

func (h interactionHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		log.Printf("[ERROR] Invalid method: %s", r.Method)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	buf, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Printf("[ERROR] Failed to read request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	jsonStr, err := url.QueryUnescape(string(buf)[8:])
	if err != nil {
		log.Printf("[ERROR] Failed to unespace request body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	var message slack.AttachmentActionCallback
	if err := json.Unmarshal([]byte(jsonStr), &message); err != nil {
		log.Printf("[ERROR] Failed to decode json message from slack: %s", jsonStr)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Only accept message from slack with valid token
	if message.Token != h.verificationToken {
		log.Printf("[ERROR] Invalid token: %s", message.Token)
		w.WriteHeader(http.StatusUnauthorized)
		return
	}

	var env EnvConfig
	if err := envconfig.Process("", &env); err != nil {
		log.Printf("[ERROR] Failed to process env var: %s", err)
		return
	}

	action := message.Actions[0]
	switch action.Name {
	case actionWind:

		forecastDatas := wind.MakeForecastData(forecastURL, forecastFilePath)
		forecastDataNum := len(forecastDatas)

		text := "この" + strconv.Itoa(forecastDataNum) + "日間の最大風速は"
		exceedLimit := ""
		max := -1
		maxDay := ""
		maxTime := ""

		for i := 0; i < forecastDataNum; i++ {
			forecastData := forecastDatas[i]
			if dayMax, res := forecastData.MaxSpeed(); dayMax > max {
				max = dayMax
				maxDay = forecastData.Date
				maxTime = res
			}
			if isExceed, res := forecastData.IsExceededLimit(limitSpeed); isExceed {
				exceedLimit += res
			}
		}

		text += maxDay + maxTime + "の" + strconv.Itoa(max) + "m/sだよ！\n" + strconv.Itoa(limitSpeed) + "m/sを超える日は"
		if exceedLimit != "" {
			text += exceedLimit + "だよ〜！"
		} else {
			text += "ありません！"
		}

		responseMessage(w, message.OriginalMessage, "", text)
		return
	case actionOrder:
		text := "発注板だよ〜！\n" + env.OrderURL
		responseMessage(w, message.OriginalMessage, "", text)
		return
	case actionCancel:
		text := message.User.Name + "さん、じゃあね〜！"
		responseMessage(w, message.OriginalMessage, "", text)
		return
	default:
		log.Printf("[ERROR] ]Invalid action was submitted: %s", action.Name)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
}

// responseMessage response to the original slackbutton enabled message.
// It removes button and replace it with message which indicate how bot will work
func responseMessage(w http.ResponseWriter, original slack.Message, title, value string) {
	original.Attachments[0].Text = ""
	original.Attachments[0].Actions = []slack.AttachmentAction{} // empty buttons
	original.Attachments[0].Fields = []slack.AttachmentField{
		{
			Title: title,
			Value: value,
			Short: false,
		},
	}

	original.ReplaceOriginal = true

	w.Header().Add("Content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&original)
}
