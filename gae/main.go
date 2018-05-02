package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	"google.golang.org/appengine"
	aelog "google.golang.org/appengine/log"
	"google.golang.org/appengine/urlfetch"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/line/line-bot-sdk-go/linebot/httphandler"
)

func init() {
	handler, err := httphandler.New(
		os.Getenv("CHANNEL_SECRET"),
		os.Getenv("CHANNEL_TOKEN"),
	)
	if err != nil {
		log.Fatal(err)
	}

	handler.HandleEvents(func(events []*linebot.Event, r *http.Request) {
		ctx := appengine.NewContext(r)
		client, err := handler.NewClient(linebot.WithHTTPClient(urlfetch.Client(ctx)))
		if err != nil {
			aelog.Errorf(ctx, "%v", err)
			return
		}

		for _, event := range events {
			switch event.Type {
			case linebot.EventTypeMessage:
				switch event.Message.(type) {
				case *linebot.TextMessage:
					handleTextMessage(ctx, event, client)
				}
			}
		}
	})

	http.Handle("/callback", handler)
}

func handleTextMessage(ctx context.Context, event *linebot.Event, client *linebot.Client) {
	message, ok := event.Message.(*linebot.TextMessage)
	if !ok {
		return
	}

	handlers := []func(string) (linebot.Message, bool){searchNomiyaHandler}
	for _, handler := range handlers {
		if reply, ok := handler(message.Text); ok {
			client.ReplyMessage(event.ReplyToken, reply).WithContext(ctx).Do()
			return
		}
	}
}

func searchNomiyaHandler(message string) (reply linebot.Message, match bool) {
	match = strings.HasSuffix(message, "の飲み屋")
	if !match {
		return
	}

	query := strings.TrimSuffix(message, "の飲み屋")
	reply = linebot.NewTextMessage(fmt.Sprintf("%sの飲み屋を探します", query))
	return
}

func main() {
	appengine.Main()
}
