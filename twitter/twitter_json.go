package twitter

import (
	"encoding/json"
	"fmt"

	"github.com/go-faster/errors"
)

//{
//	"bitrate": 632000,
//  "content_type": "video/mp4",
//	"url": "https://video.twimg.com/ext_tw_video/1742758051872960512/pu/vid/avc1/320x568/39j_OueTsS4i2p8-.mp4?tag=12"
//}

type VideoData struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	URL         string `json:"url"`
}

type PhotoData struct {
	MediaURLHttps string `json:"media_url_https"`
}

type TweetData struct {
	Url    TwitterURL
	Videos []VideoData
	Photos []PhotoData
}

func (td *TweetData) AddPhoto(pd PhotoData) {

	for _, p := range td.Photos {
		if p.MediaURLHttps == pd.MediaURLHttps {
			return
		}
	}

	td.Photos = append(td.Photos, pd)
}

func (td *TweetData) AddVideo(vd VideoData) {
	td.Videos = append(td.Videos, vd)
}

func (td *TweetData) HasVideos() bool {
	return len(td.Videos) > 0
}

func (td *TweetData) Photo() (PhotoData, bool) {
	if len(td.Photos) == 0 {
		return PhotoData{}, false
	}

	return td.Photos[0], true
}

func (td *TweetData) HasPhotos() bool {
	return len(td.Photos) > 0
}

func (td *TweetData) VideoBestBitrate() (VideoData, bool) {
	if len(td.Videos) == 0 {
		return VideoData{}, false
	}

	best := td.Videos[0]

	for _, v := range td.Videos {
		if v.Bitrate > best.Bitrate {
			best = v
		}
	}

	return best, true
}

func hasKey(aMap map[string]interface{}, key string) bool {
	_, ok := aMap[key]
	return ok
}

func tryGetKeyString(aMap map[string]interface{}, key string) (string, bool) {
	val, ok := aMap[key]
	if ok {
		str, ok := val.(string)

		if ok {
			return str, true
		}
	}
	return "", false
}

func tryGetKeyInt(aMap map[string]interface{}, key string) (int, bool) {
	val, ok := aMap[key]
	if ok {
		num, ok := val.(json.Number)

		n, err := num.Int64()

		if err != nil {
			return 0, false
		}

		if ok {
			return int(n), true
		}
	}
	return 0, false

}

func tryParsePhotoData(aMap map[string]interface{}) (PhotoData, bool) {
	id := PhotoData{}

	if mediaURLHttps, ok := tryGetKeyString(aMap, "media_url_https"); ok {
		id.MediaURLHttps = mediaURLHttps
	} else {
		return id, false
	}

	if typ, ok := tryGetKeyString(aMap, "type"); ok {
		if typ != "photo" {
			return id, false
		}
	}

	return id, true
}

func tryParseVideoData(aMap map[string]interface{}) (VideoData, bool) {
	vd := VideoData{}

	if bitrate, ok := tryGetKeyInt(aMap, "bitrate"); ok {
		vd.Bitrate = bitrate
	} else {
		return vd, false
	}

	if contentType, ok := tryGetKeyString(aMap, "content_type"); ok {
		vd.ContentType = contentType
	} else {
		return vd, false
	}

	if url, ok := tryGetKeyString(aMap, "url"); ok {
		vd.URL = url
	} else {
		return vd, false
	}

	return vd, true
}

func parseMap(d *TweetData, aMap map[string]interface{}) {

	if vd, ok := tryParseVideoData(aMap); ok {
		d.AddVideo(vd)
		return
	}

	if pd, ok := tryParsePhotoData(aMap); ok {
		d.AddPhoto(pd)
		return
	}

	for _, val := range aMap {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			//fmt.Println(key)
			parseMap(d, concreteVal)
		case []interface{}:
			//fmt.Println(key)
			parseArray(d, concreteVal)
		default:
			//fmt.Println(key, ":", concreteVal)
		}
	}
}

func parseArray(d *TweetData, anArray []interface{}) {
	for _, val := range anArray {
		switch concreteVal := val.(type) {
		case map[string]interface{}:
			//fmt.Println("Index:", i)
			parseMap(d, concreteVal)
		case []interface{}:
			//fmt.Println("Index:", i)
			parseArray(d, concreteVal)
		default:
			//fmt.Println("Index", i, ":", concreteVal)

		}
	}
}

func ParseData(json any) (*TweetData, error) {

	data := &TweetData{}

	switch concreteVal := json.(type) {
	case map[string]interface{}:
		parseMap(data, concreteVal)
	case []interface{}:
		parseArray(data, concreteVal)
	default:
		return nil, errors.New("Invalid JSON")
	}

	return data, nil
}

func (td *TweetData) String() string {
	return fmt.Sprintf("Videos: %v, Photos: %v", td.Videos, td.Photos)
}
