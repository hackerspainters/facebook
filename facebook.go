package facebook

import (
	"encoding/json"
	"fmt"
	"github.com/chaosphere2112/go-jsonUtil"
	"net/http"
	"strconv"
	"strings"
)

type AccessToken struct {
	Token  string
	Expiry int64
}

func readHttpBody(response *http.Response) string {

	fmt.Println("Reading body")

	bodyBuffer := make([]byte, 1000)
	var str string

	count, err := response.Body.Read(bodyBuffer)

	for ; count > 0; count, err = response.Body.Read(bodyBuffer) {

		if err != nil {

		}

		str += string(bodyBuffer[:count])
	}

	return str

}

//Converts a code to an Auth_Token
func GetAccessToken(client_id string, code string, secret string, callbackUri string) AccessToken {
	fmt.Println("GetAccessToken")
	//https://graph.facebook.com/oauth/access_token?client_id=YOUR_APP_ID&redirect_uri=YOUR_REDIRECT_URI&client_secret=YOUR_APP_SECRET&code=CODE_GENERATED_BY_FACEBOOK
	response, err := http.Get("https://graph.facebook.com/oauth/access_token?client_id=" +
		client_id + "&redirect_uri=" + callbackUri +
		"&client_secret=" + secret + "&code=" + code)

	if err == nil {

		auth := readHttpBody(response)

		var token AccessToken

		tokenArr := strings.Split(auth, "&")

		token.Token = strings.Split(tokenArr[0], "=")[1]
		expireInt, err := strconv.Atoi(strings.Split(tokenArr[1], "=")[1])

		if err == nil {
			token.Expiry = int64(expireInt)
		}

		return token
	}

	var token AccessToken

	return token
}

func GetMe(token AccessToken) string {
	fmt.Println("Getting me")
	response, err := GetUncachedResponse("https://graph.facebook.com/me?access_token=" + token.Token)

	if err == nil {

		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		if responseBody != "" {
			err = json.Unmarshal([]byte(responseBody), &jsonBlob)

			if err == nil {
				jsonObj := jsonBlob.(map[string]interface{})
				return jsonObj["id"].(string)
			}
		}
		return err.Error()
	}

	return err.Error()
}

func GetUncachedResponse(uri string) (*http.Response, error) {
	fmt.Println("Uncached response GET")
	request, err := http.NewRequest("GET", uri, nil)

	if err == nil {
		request.Header.Add("Cache-Control", "no-cache")

		client := new(http.Client)

		return client.Do(request)
	}

	if err != nil {
	}
	return nil, err

}

type GroupEvent struct {
	Data []struct {
		Name      string
		StartTime string `json:"start_time"`
		Timezone  string
		Location  string
		Id        string
	}
	Paging struct {
		Cursors struct {
			After  string
			Before string
		}
	}
}

func GetGroupEvents(token *AccessToken, groupId string) GroupEvent {

	group_events := GroupEvent{}
	response, err := GetUncachedResponse("https://graph.facebook.com/" + groupId + "/events?access_token=" + token.Token)

	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {
			b := []byte(body)
			json.Unmarshal(b, &group_events)
			return group_events
		}

	}

	return GroupEvent{}
}

func GetGroupEventIds(group_events GroupEvent) []string {

	event_ids := make([]string, len(group_events.Data))
	for i := 0; i < len(group_events.Data); i++ {
		event_ids[i] = group_events.Data[i].Id
	}

	return event_ids
}

type Event struct {
	Id    string
	Owner struct {
		Name    string
		Id      string
	}
	Name            string
	Description     string
	StartTime		string    `bson:"start_time" json:"start_time"`
	EndTime         string    `bson:"end_time" json:"end_time"`
	TimeZone		string
	IsDateOnly		bool
	Location		string
	Venue struct {
		Latitude	float64
		Longitude	float64
		City		string
		Country	    string
		Id			string
		Street		string
		Zip		    string
	}
	UpdatedTime	    string     `bson:"updated_time" json:"updated_time"`
}

func GetEvent(token *AccessToken, eventId string) Event {

	event := Event{}
	response, err := GetUncachedResponse("https://graph.facebook.com/" + eventId + "?access_token=" + token.Token)

	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {
			b := []byte(body)
			json.Unmarshal(b, &event)
			return event
		}

	}

	return Event{}
}

func getPhotoSource(token *AccessToken, photoId string) string {
	fmt.Println("Getting photo source")
	response, err := GetUncachedResponse("https://graph.facebook.com/" + photoId + "?access_token=" + token.Token + "&fields=source")

	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {

			object, err := jsonUtil.JsonFromString(body)

			if err == nil {

				source, err := object.String("source")

				if err == nil {

					return source

				}

			}

		}

	}

	return ""

}

func GetAlbumPhotos(token *AccessToken, albumId string) []string {
	response, err := GetUncachedResponse("https://graph.facebook.com/" + albumId + "/photos?access_token=" + token.Token + "&fields=images&limit=1000")
	fmt.Println("Getting album photos")
	if err == nil && response != nil {

		body := readHttpBody(response)

		if body != "" {

			object, err := jsonUtil.JsonFromString(body)
			if err == nil {
				//Get the "data" array
				var data jsonUtil.JsonArray
				returned := make([]string, 0, 1)

				data, err = object.Array("data")

				if err == nil {
					fmt.Println(len(data))
					for dataIndex := 0; dataIndex < len(data); dataIndex++ {
						var anonObj jsonUtil.JsonObject
						anonObj, err = data.Object(dataIndex)
						if err == nil {

							//For each object in data, iterate over the "images" array
							var imagesArray jsonUtil.JsonArray

							imagesArray, err = anonObj.Array("images")
							if err == nil {
								//Get the source for each image
								var image jsonUtil.JsonObject
								image, err = imagesArray.Object(1)

								if err == nil {
									var photoSource string
									photoSource, err = image.String("source")

									if err == nil {
										returned = append(returned, photoSource)
									}
								}
							}

						}

					}

				}
				return returned
			}
		}

	}

	if err != nil {
		returned := make([]string, 1)
		returned[0] = err.Error()
		return returned
	}

	return make([]string, 0)
}

func GetPhotos(token *AccessToken) []string {
	fmt.Println("Getting photos")
	response, err := GetUncachedResponse("https://graph.facebook.com/me/albums?access_token=" + token.Token + "&fields=id")

	if err == nil && response != nil {
		var jsonBlob interface{}

		responseBody := readHttpBody(response)

		if responseBody != "" {
			err = json.Unmarshal([]byte(responseBody), &jsonBlob)

			if err == nil {
				jsonObj := jsonBlob.(map[string]interface{})

				dataArray := jsonObj["data"].([]interface{})

				first := dataArray[0].(map[string]interface{})

				firstAlbumId := first["id"].(string)

				//Feed albumId into GetAlbumPhotos
				return GetAlbumPhotos(token, firstAlbumId)

			}
		}
	}
	fmt.Println("Failed to GetPhotos because " + err.Error())
	return nil
}
