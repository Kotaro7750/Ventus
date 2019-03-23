package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/Kotaro7750/Ventus/wind"
	"github.com/nlopes/slack"
)

const (
	// action is used for slack attament action.
	actionWind   = "wind"
	actionOrder  = "order"
	actionCancel = "cancel"
	color        = "#1e90ff"
)

type SlackListener struct {
	client    *slack.Client
	botID     string
	channelID string
}

// LstenAndResponse listens slack events and response
// particular messages. It replies by slack message button.
func (s *SlackListener) ListenAndResponse() {
	rtm := s.client.NewRTM()

	// Start listening slack events
	go rtm.ManageConnection()

	// Handle slack events
	for msg := range rtm.IncomingEvents {
		switch ev := msg.Data.(type) {
		case *slack.MessageEvent:
			if err := s.handleMessageEvent(ev); err != nil {
				log.Printf("[ERROR] Failed to handle message: %s", err)
			}
		}
	}
}

// handleMesageEvent handles message events.
func (s *SlackListener) handleMessageEvent(ev *slack.MessageEvent) error {
	// Only response in specific channel. Ignore else.
	if ev.Channel != s.channelID {
		log.Printf("%s %s", ev.Channel, ev.Msg.Text)
		return nil
	}

	// Only response mention to bot. Ignore else.
	if !strings.HasPrefix(ev.Msg.Text, fmt.Sprintf("<@%s> ", s.botID)) {
		return nil
	}

	// value is passed to message handler when request is approved.
	attachment := slack.Attachment{
		Text:       "なあに？",
		Color:      color,
		CallbackID: "command",
		Actions: []slack.AttachmentAction{
			{
				Name:  actionWind,
				Text:  "風速",
				Type:  "button",
				Style: "primary",
			},

			{
				Name:  actionOrder,
				Text:  "広報物発注板",
				Type:  "button",
				Style: "default",
			},

			{
				Name:  actionCancel,
				Text:  "なんでもない！",
				Type:  "button",
				Style: "danger",
			},
		},
	}

	msgOptText := slack.MsgOptionText("", true)

	msgOptAttachment := slack.MsgOptionAttachments(attachment)

	if _, _, err := s.client.PostMessage(ev.Channel, msgOptText, msgOptAttachment); err != nil {
		return fmt.Errorf("failed to post message: %s", err)
	}

	return nil
}

// PostWindReport returns windreport
func (s *SlackListener) PostWindReport() {
	forecastDatas := wind.MakeForecastData(ForecastURL, ForecastFilePath)

	text := forecastDatas.MakeWindReport(LimitSpeed)

	msgOptText := slack.MsgOptionText("", true)
	attachment := slack.Attachment{
		Text:  text,
		Color: color,
	}

	msgOptAttachment := slack.MsgOptionAttachments(attachment)

	if _, _, err := s.client.PostMessage(s.channelID, msgOptText, msgOptAttachment); err != nil {
		log.Printf("failed to post message: %s", err)
	}
}
