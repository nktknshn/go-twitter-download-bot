package twitter

import (
	"fmt"
	"regexp"
)

func ParseURLFilename(url string) string {
	re := regexp.MustCompile(`[^/]+$`)
	fn := re.FindString(url)
	return regexp.MustCompile(`\?.*$`).ReplaceAllString(fn, "")
}

type VideoVariant struct {
	Bitrate     int    `json:"bitrate"`
	ContentType string `json:"content_type"`
	VideoURL    string `json:"url"`
}

func (vd VideoVariant) URL() string {
	return vd.VideoURL
}

func (vd VideoVariant) Filename() string {
	return ParseURLFilename(vd.VideoURL)
}

type Photo struct {
	MediaURLHttps string `json:"media_url_https"`
}

func (pd Photo) Filename() string {
	return ParseURLFilename(pd.MediaURLHttps)
}

func (pd Photo) URL() string {
	return pd.MediaURLHttps
}

type Video struct {
	MediaKey string        `json:"media_key"`
	Variants VideoVariants `json:"video_variants"`
}

type VideoVariants []VideoVariant

func (vv VideoVariants) VideoBestBitrate() (VideoVariant, bool) {
	if len(vv) == 0 {
		return VideoVariant{}, false
	}
	best := vv[0]

	for _, v := range vv {
		if v.Bitrate > best.Bitrate {
			best = v
		}
	}
	return best, true
}

type TweetData struct {
	Url      TwitterURL
	FullText string
	Text     string
	Videos   []Video
	Photos   []Photo
}

func (td *TweetData) NoMedia() bool {
	return len(td.Videos) == 0 && len(td.Photos) == 0
}

func (td *TweetData) IsEmpty() bool {
	return td.NoMedia() && td.Text == "" && td.FullText == ""
}

func (td *TweetData) BestBitrateVideos() []VideoVariant {
	bestVariants := make([]VideoVariant, 0)
	for _, vv := range td.Videos {
		best, ok := vv.Variants.VideoBestBitrate()
		if !ok {
			continue
		}
		bestVariants = append(bestVariants, best)
	}
	return bestVariants
}

func (td *TweetData) String() string {
	return fmt.Sprintf("Videos: %v, Photos: %v, FullText: %s, Text: %s", td.BestBitrateVideos(), td.Photos, td.CleanText(), td.Text)
}

// strip https://t.co/* in the end
func (td *TweetData) CleanText() string {
	return regexp.MustCompile(`https://t.co/\w+$`).ReplaceAllString(td.FullText, "")
}

func (td *TweetData) TweetText() string {
	if td.Text != "" {
		return td.Text
	}
	return td.CleanText()
}

func (td *TweetData) AddPhoto(pd Photo) {

	for _, p := range td.Photos {
		if p.MediaURLHttps == pd.MediaURLHttps {
			return
		}
	}

	td.Photos = append(td.Photos, pd)
}

func (td *TweetData) AddVideo(vd Video) {
	for _, v := range td.Videos {
		if v.MediaKey == vd.MediaKey {
			return
		}
	}
	td.Videos = append(td.Videos, vd)
}

func (td *TweetData) HasVideos() bool {
	return len(td.Videos) > 0
}

func (td *TweetData) Photo() (Photo, bool) {
	if len(td.Photos) == 0 {
		return Photo{}, false
	}

	return td.Photos[0], true
}

func (td *TweetData) PhotoCount() int {
	return len(td.Photos)
}

func (td *TweetData) HasPhotos() bool {
	return len(td.Photos) > 0
}
