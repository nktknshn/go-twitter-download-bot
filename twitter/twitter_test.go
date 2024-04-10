package twitter

// testify
import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestParseUrl(t *testing.T) {
	t.Run("url1", func(t *testing.T) {
		url := "https://x.com/contextdogs/status/1742878545549087076?s=35"
		id, err := ParseTwitterURL(url)
		require.NoError(t, err)
		require.Equal(t, TwitterURL{
			User: "contextdogs",
			ID:   "1742878545549087076",
		}, id)
	})

	t.Run("url2", func(t *testing.T) {
		url := "https://twitter.com/contextdogs/status/1742878545549087076"
		id, err := ParseTwitterURL(url)
		require.NoError(t, err)
		require.Equal(t, TwitterURL{
			User: "contextdogs",
			ID:   "1742878545549087076",
		}, id)
	})

	t.Run("invalid", func(t *testing.T) {
		url := "https://google.com"
		_, err := ParseTwitterURL(url)
		require.Error(t, err)
	})

}
