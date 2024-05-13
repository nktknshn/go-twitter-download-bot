package twitter

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"

	"github.com/go-faster/errors"
	"github.com/go-resty/resty/v2"
	"go.uber.org/zap"
)

func saveBody(resp *resty.Response, path string) error {
	return os.WriteFile(path, resp.Body(), 0644)
}

var logger *zap.Logger = zap.Must(zap.NewDevelopmentConfig().Build())

type TwitterURL struct {
	User string
	ID   string
}

func (tu *TwitterURL) String() string {
	return fmt.Sprintf("https://twitter.com/%s/status/%s", tu.User, tu.ID)
}

var rexURL = regexp.MustCompile(`https://(?:www\.)?(twitter|x)\.com/(?P<user>[^/]+)/status/(?P<id>\d+)`)

func IsValidTwitterURL(url string) bool {
	_, err := ParseTwitterURL(url)
	return err == nil
}

func ParseTwitterURL(url string) (TwitterURL, error) {
	match := rexURL.FindStringSubmatch(url)

	if match == nil {
		return TwitterURL{}, fmt.Errorf("invalid url")
	}

	return TwitterURL{User: match[2], ID: match[3]}, nil
}

type Twitter struct {
	logger     *zap.Logger
	httpClient *resty.Client
	saveData   bool
}

type Options struct {
	httpClient *resty.Client
	retryCount int
	saveData   bool
}

type Option func(*Options)

func DefaultResty() *resty.Client {
	r := resty.New()
	r.SetHeader("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/115.0")
	return r
}

func WithRestyClient(r *resty.Client) Option {
	return func(o *Options) {
		o.httpClient = r
	}
}

func WithRetryCount(n int) Option {
	return func(o *Options) {
		o.retryCount = n
	}
}

func WithSaveData(sd bool) Option {
	return func(o *Options) {
		o.saveData = sd
	}
}

func NewTwitter(opts ...Option) *Twitter {

	options := &Options{
		retryCount: 3,
		httpClient: DefaultResty(),
	}

	for _, opt := range opts {
		opt(options)
	}

	options.httpClient.SetRetryCount(options.retryCount)

	return &Twitter{
		httpClient: options.httpClient,
		logger:     logger.Named("twitter"),
		saveData:   options.saveData,
	}
}

type Tokens struct {
	Bearer     string
	GuestToken string
}

func (t Tokens) String() string {
	return fmt.Sprintf("Bearer %s, GuestToken %s", t.Bearer, t.GuestToken)
}

func (t *Twitter) GetTokens(ctx context.Context, url string) (Tokens, error) {
	res := Tokens{}

	turl, err := ParseTwitterURL(url)

	if err != nil {
		return res, errors.Wrap(err, "failed to parse twitter url")
	}

	resp, err := t.httpClient.R().SetContext(ctx).Get(turl.String())

	if err != nil {
		return res, errors.Wrap(err, "failed to get twitter url")
	}

	if t.saveData {
		if err := saveBody(resp, "samples/twitter.html"); err != nil {
			return res, errors.Wrap(err, "failed to save twitter html")
		}
	}

	// https://abs.twimg.com/responsive-web/client-web-legacy/main.3ba1b53a.js
	// https://abs.twimg.com/responsive-web/client-web/main.3b59231a.js
	rextGuestToken := regexp.MustCompile(`cookie="gt=(\d+)`)
	rexMainJsURL := regexp.MustCompile(`https://abs.twimg.com/responsive-web/client-web-legacy/main\.[a-f0-9]+\.js`)

	matchGuestToken := rextGuestToken.FindStringSubmatch(string(resp.Body()))

	if matchGuestToken == nil {
		return res, errors.New("failed to find guest token")
	}

	res.GuestToken = matchGuestToken[1]

	matchMainJsURL := rexMainJsURL.FindStringSubmatch(string(resp.Body()))

	if matchMainJsURL == nil {
		return res, errors.New("failed to find main js url")
	}

	t.logger.Debug("main js url", zap.String("url", matchMainJsURL[0]))

	resp, err = t.httpClient.R().SetContext(ctx).Get(matchMainJsURL[0])

	if err != nil {
		return res, errors.Wrap(err, "failed to get main js url")
	}

	rexBearerToken := regexp.MustCompile(`Bearer ([a-zA-Z0-9%]+)`)

	bearerMatches := rexBearerToken.FindAllStringSubmatch(string(resp.Body()), -1)

	if len(bearerMatches) == 0 {
		return res, errors.New("failed to find bearer token")
	}

	lastMatch := bearerMatches[len(bearerMatches)-1]
	res.Bearer = lastMatch[1]

	return res, nil
}

func (t *Twitter) GetURLJSON(ctx context.Context, posturl string) ([]byte, error) {

	tu, err := ParseTwitterURL(posturl)

	if err != nil {
		return nil, errors.Wrap(err, "failed to parse twitter url")
	}

	bt, err := t.GetTokens(ctx, posturl)

	t.logger.Debug("tokens", zap.Any("tokens", bt))

	if err != nil {
		return nil, errors.Wrap(err, "failed to get bearer token")
	}

	variable := url.QueryEscape(fmt.Sprintf(`{"tweetId":"%s","withCommunity":false,"includePromotedContent":false,"withVoice":false}`, tu.ID))
	features := url.QueryEscape(`{"creator_subscriptions_tweet_preview_api_enabled":true,"communities_web_enable_tweet_community_results_fetch":true,"c9s_tweet_anatomy_moderator_badge_enabled":true,"articles_preview_enabled":true,"tweetypie_unmention_optimization_enabled":true,"responsive_web_edit_tweet_api_enabled":true,"graphql_is_translatable_rweb_tweet_is_translatable_enabled":true,"view_counts_everywhere_api_enabled":true,"longform_notetweets_consumption_enabled":true,"responsive_web_twitter_article_tweet_consumption_enabled":true,"tweet_awards_web_tipping_enabled":false,"creator_subscriptions_quote_tweet_preview_enabled":false,"freedom_of_speech_not_reach_fetch_enabled":true,"standardized_nudges_misinfo":true,"tweet_with_visibility_results_prefer_gql_limited_actions_policy_enabled":true,"tweet_with_visibility_results_prefer_gql_media_interstitial_enabled":true,"rweb_video_timestamps_enabled":true,"longform_notetweets_rich_text_read_enabled":true,"longform_notetweets_inline_media_enabled":true,"rweb_tipjar_consumption_enabled":true,"responsive_web_graphql_exclude_directive_enabled":true,"verified_phone_label_enabled":false,"responsive_web_graphql_skip_user_profile_image_extensions_enabled":false,"responsive_web_graphql_timeline_navigation_enabled":true,"responsive_web_enhance_cards_enabled":false}`)
	fields := url.QueryEscape(`{"withArticleRichContentState":true,"withArticlePlainText":false}`)

	graphqlURL := fmt.Sprintf("https://api.twitter.com/graphql/7xflPyRiUxGVbJd4uWmbfg/TweetResultByRestId?variables=%s&features=%s&fieldToggles=%s", variable, features, fields)

	// guest_id_marketing=v1%3A171561022759809470; guest_id_ads=v1%3A171561022759809470; personalization_id="v1_f7qgca5qLTAZPj0EXeaJcA=="; guest_id=v1%3A171561022759809470; gt=1790025150592598265

	t.logger.Debug("cookie", zap.Any("value", t.httpClient.Cookies))

	resp, err := t.httpClient.R().
		SetCookie(&http.Cookie{
			Name:  "gt", // guest token
			Value: bt.GuestToken,
		}).
		SetContext(ctx).
		SetHeader("authorization", "Bearer "+bt.Bearer).
		SetHeader("x-guest-token", bt.GuestToken).
		SetHeader("Accept", "*/*").
		SetHeader("X-Twitter-Active-User", "yes").
		SetHeader("X-Twitter-Client-Language", "en").
		SetHeader("Content-Type", "application/json").
		SetHeader("Referer", "https://twitter.com/").
		SetHeader("Origin", "https://twitter.com").
		Get(graphqlURL)

	if err != nil {
		return nil, errors.Wrap(err, "failed to get graphql url")
	}

	if t.saveData {
		if err := saveBody(resp, "samples/twitter.json"); err != nil {
			return nil, errors.Wrap(err, "failed to save twitter json")
		}
	}

	t.logger.Debug("response", zap.Any("status", resp.Status()), zap.String("body", string(resp.Body())))

	return resp.Body(), nil
}

func (t *Twitter) GetTwitterData(ctx context.Context, url string) (*TweetData, error) {
	p := TwitterParser{}
	turl, err := ParseTwitterURL(url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse twitter url")
	}
	body, err := t.GetURLJSON(ctx, url)
	if err != nil {
		return nil, errors.Wrap(err, "failed to get url json")
	}
	var jsonBody interface{}
	if err := JsonDecodeWithNumberBytes(body, &jsonBody); err != nil {
		return nil, errors.Wrap(err, "failed to unmarshal json")
	}
	td := p.Parse(jsonBody)
	td.Url = turl
	return &td, nil
}
