package twitter

type TwitterParser struct {
	d TweetData
}

func (tp *TwitterParser) Parse(a interface{}) TweetData {
	parse(tp, a)
	return tp.d
}

func (tp *TwitterParser) ParseMap(aMap map[string]interface{}) {

	if vd, ok := tryParseVideo(aMap); ok {
		tp.d.AddVideo(vd)
		return
	}

	if pd, ok := tryParsePhotoData(aMap); ok {
		tp.d.AddPhoto(pd)
		return
	}

	if ft, ok := tryParseFullText(aMap); ok {
		tp.d.FullText = ft
	}

	if ft, ok := tryParseText(aMap); ok {
		tp.d.Text = ft
	}

	// go deeper
	parseMap(tp, aMap)
}

// parse video with variants
func tryParseVideo(aMap map[string]interface{}) (Video, bool) {
	res := Video{}

	if v, _ := tryGetKeyString(aMap, "type"); v != "video" {
		return res, false
	}

	if v, ok := tryGetKeyString(aMap, "media_key"); ok {
		res.MediaKey = v
	} else {
		return res, false
	}

	if !hasKey(aMap, "video_info") {
		return res, false
	}

	if m, ok := aMap["video_info"].(map[string]interface{}); ok {
		vp := variantsParser{}
		vp.ParseMap(m)
		res.Variants = vp.variants
		return res, true
	}

	return res, false
}

type variantsParser struct {
	variants VideoVariants
}

func (vp *variantsParser) ParseMap(aMap map[string]interface{}) {
	if vd, ok := tryParseVideoData(aMap); ok {
		vp.variants = append(vp.variants, vd)
		return
	}
	parseMap(vp, aMap)
}

func (vp *variantsParser) ParseArray(anArray []interface{}) {
	parseArray(vp, anArray)
}

func tryParsePhotoData(aMap map[string]interface{}) (Photo, bool) {
	id := Photo{}

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

func tryParseVideoData(aMap map[string]interface{}) (VideoVariant, bool) {
	vd := VideoVariant{}

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
		vd.VideoURL = url
	} else {
		return vd, false
	}

	return vd, true
}

func tryParseFullText(aMap map[string]interface{}) (string, bool) {
	if fullText, ok := tryGetKeyString(aMap, "full_text"); ok {
		return fullText, true
	}

	return "", false
}
func tryParseText(aMap map[string]interface{}) (string, bool) {
	if fullText, ok := tryGetKeyString(aMap, "text"); ok {
		return fullText, true
	}

	return "", false
}

func (tp *TwitterParser) ParseArray(anArray []interface{}) {
	parseArray(tp, anArray)
}
