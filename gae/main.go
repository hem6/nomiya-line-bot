package app

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strings"
	"time"

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

	handlers := []func(context.Context, string) ([]linebot.Message, bool){searchNomiyaHandler}
	for _, handler := range handlers {
		if replies, ok := handler(ctx, message.Text); ok {
			_, err := client.ReplyMessage(event.ReplyToken, replies...).WithContext(ctx).Do()
			if err != nil {
				aelog.Warningf(ctx, "reply error: %v", err)
			}
			return
		}
	}
}

func searchNomiyaHandler(ctx context.Context, message string) (replies []linebot.Message, match bool) {
	match = strings.HasSuffix(message, "の飲み屋")
	if !match {
		return
	}

	query := strings.TrimSuffix(message, "の飲み屋")
	gnaviClient := NewGnaviClient(os.Getenv("GNAVI_KEY"), urlfetch.Client(ctx))
	restaurants, err := gnaviClient.SearchResturant(fmt.Sprintf("居酒屋,%s", query))
	if err != nil {
		return replies, false
	}

	if len(restaurants) == 0 {
		replies = []linebot.Message{
			linebot.NewTextMessage("すまんな、ええ飲み屋が見つからんかったわ……"),
		}
		return
	}

	rand.Seed(time.Now().UnixNano())
	choices := rand.Perm(len(restaurants))

	var columns []*linebot.CarouselColumn
	for i, randomIndex := range choices {
		if i >= 3 {
			break
		}

		restaurant := restaurants[randomIndex]
		if restaurant.Name == "" {
			restaurant.Name = "店名なし"
		}
		if restaurant.URL == "" {
			restaurant.URL = "https://www.gnavi.co.jp/"
		}
		if restaurant.ImageURL == "" {
			restaurant.ImageURL = "https://" + appengine.DefaultVersionHostname(ctx) + "/images/nomikai_salaryman.png"
		}

		action := linebot.NewURITemplateAction("ぐるなびを開く", restaurant.URL)
		columns = append(columns, linebot.NewCarouselColumn(restaurant.ImageURL, "", restaurant.Name, action))
	}

	moreAction := linebot.NewMessageTemplateAction("もう一回聞く", message)
	columns = append(columns, linebot.NewCarouselColumn("https://"+appengine.DefaultVersionHostname(ctx)+"/images/50th_beer.png", "", "どや？　ええところはあったか？", moreAction))
	template := linebot.NewCarouselTemplate(columns...)

	replies = []linebot.Message{
		linebot.NewTextMessage("ほれ、調べたったで\n(Supported by ぐるなびWebService : http://api.gnavi.co.jp/api/scope/)"),
		linebot.NewTemplateMessage("(飲み屋おじさんの飲み屋リスト)", template),
	}
	return
}
