package bot

import (
	"fmt"

	"github.com/nktknshn/go-twitter-download-bot/bot"
	"github.com/nktknshn/go-twitter-download-bot/cli/logging"
	"github.com/spf13/cobra"
)

var (
	logger                = logging.GetLogger().Named("cmdbot")
	flagDownloadFolder    string
	flagAdminID           int64
	flagRestrictToAdminID bool
	flagForwardTo         int64
	flagDebug             bool
	flagSessionFile       string = "twitter-downloader-session.json"
	flagUseLimiter        bool

	flagIncludeText    bool
	flagIncludeURL     bool
	flagIncludeBotName bool

	flagLimitPending int = 1
	flagLimitPerDay  int = 30
)

func init() {
	Cmd.AddCommand(cmdStart)

	cmdStart.PersistentFlags().Int64VarP(&flagAdminID, "admin-id", "a", 0, "admin id")
	cmdStart.PersistentFlags().BoolVarP(&flagRestrictToAdminID, "restrict-to-admin-id", "r", flagRestrictToAdminID, "Restrict usage to admin id")

	cmdStart.PersistentFlags().StringVarP(&flagSessionFile, "session-file", "s", flagSessionFile, "session file")
	cmdStart.PersistentFlags().StringVarP(&flagDownloadFolder, "download-folder", "d", "", "download folder")
	cmdStart.PersistentFlags().Int64VarP(&flagForwardTo, "forward-to", "f", 0, "forward media that was sent to a user to a channel")
	cmdStart.PersistentFlags().BoolVarP(&flagDebug, "debug-telegram", "D", false, "debug log")
	cmdStart.PersistentFlags().BoolVarP(&flagUseLimiter, "use-limiter", "l", true, "use rate limiter for telegram api calls")

	cmdStart.PersistentFlags().BoolVarP(&flagIncludeText, "include-text", "T", false, "post will include text")
	cmdStart.PersistentFlags().BoolVarP(&flagIncludeURL, "include-url", "U", false, "post will include tweet url")
	cmdStart.PersistentFlags().BoolVarP(&flagIncludeBotName, "include-bot-name", "B", false, "post will include bot name")

	cmdStart.PersistentFlags().IntVarP(&flagLimitPending, "limit-pending", "p", flagLimitPending, "limit pending requests from a user (admin has no limit)")

	cmdStart.PersistentFlags().IntVarP(&flagLimitPerDay, "limit-per-day", "L", flagLimitPerDay, "limit requests per day (admin has no limit). Will reset after restart or next day.")

}

var Cmd = &cobra.Command{
	Use:   "bot",
	Short: "bot is a command line interface for telegram bot",
	Args:  cobra.MinimumNArgs(1),
}

var cmdStart = &cobra.Command{
	Use:   "start",
	Short: "start",
	Args:  cobra.ExactArgs(0),
	RunE:  runStart,
}

func runStart(cmd *cobra.Command, args []string) error {
	if flagDownloadFolder == "" {
		return fmt.Errorf("download folder is required")
	}

	logger.Info("Starting bot")

	return bot.Run(
		cmd.Context(),
		flagDownloadFolder,
		bot.WithAdmin(flagAdminID, flagRestrictToAdminID),
		bot.WithForwardTo(flagForwardTo),
		bot.WithDebugTelegram(flagDebug),
		bot.WithRateLimiter(flagUseLimiter),
		bot.WithSessionFile(flagSessionFile),
		bot.WithPostSettings(flagIncludeText, flagIncludeURL, flagIncludeBotName),
		bot.WithLimits(flagLimitPerDay, flagLimitPending),
	)
}
