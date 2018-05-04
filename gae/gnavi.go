package app

import (
	"encoding/json"
	"net/http"
	"net/url"
)

const gnaviEndpoint string = "https://api.gnavi.co.jp/RestSearchAPI/20150630/"

type GnaviClient struct {
	key    string
	client *http.Client
}

// ぐるなびAPIではデータがない場合にnullではなく{}が返るため、interface{}で受ける
type apiResponse struct {
	Rest []struct {
		Name     interface{} `json:"name"`
		URL      interface{} `json:"url"`
		ImageURL struct {
			ShopImage1 interface{} `json:"shop_image1"`
		} `json:"image_url"`
	} `json:"rest"`
}

type Restaurant struct {
	Name     string
	URL      string
	ImageURL string
}

func NewGnaviClient(key string, client *http.Client) *GnaviClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &GnaviClient{key: key, client: client}
}

func (gc *GnaviClient) SearchResturant(keyword string) (res []*Restaurant, err error) {
	values := url.Values{}
	values.Set("keyid", gc.key)
	values.Set("format", "json")
	values.Set("hit_per_page", "50")
	values.Set("freeword", keyword)

	url := gnaviEndpoint + "?" + values.Encode()
	resp, err := gc.client.Get(url)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	apiResp := apiResponse{}
	err = json.NewDecoder(resp.Body).Decode(&apiResp)

	res = extractRestaurantsFromAPIResponse(apiResp)
	return
}

func extractRestaurantsFromAPIResponse(apiResponse apiResponse) []*Restaurant {
	var restaurants []*Restaurant
	for _, data := range apiResponse.Rest {
		restaurant := Restaurant{}
		if name, ok := data.Name.(string); ok {
			restaurant.Name = name
		}
		if url, ok := data.URL.(string); ok {
			restaurant.URL = url
		}
		if imageURL, ok := data.ImageURL.ShopImage1.(string); ok {
			restaurant.ImageURL = imageURL
		}
		restaurants = append(restaurants, &restaurant)
	}
	return restaurants
}
