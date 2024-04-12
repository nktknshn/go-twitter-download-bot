package cmdbot

import (
	"fmt"

	"github.com/nktknshn/go-twitter-download-bot/bot"
	"github.com/nktknshn/go-twitter-download-bot/cli/logging"
	"github.com/spf13/cobra"
)

var (
	logger             = logging.GetLogger().Named("cmdbot")
	flagDownloadFolder string
	flagAdminID        int64
	flagForwardTo      int64
	flagDebug          bool
)

func init() {
	CmdBot.AddCommand(cmdStart)

	cmdStart.PersistentFlags().StringVarP(&flagDownloadFolder, "download-folder", "d", "", "download folder")
	// cmdStart.PersistentFlags().Int64VarP(&flagAdminID, "admin-id", "a", 0, "admin id")
	cmdStart.PersistentFlags().Int64VarP(&flagForwardTo, "forward-to", "f", 0, "forward media that was sent to a user to a channel")
	cmdStart.PersistentFlags().BoolVarP(&flagDebug, "debug-telegram", "D", false, "debug log")

}

var CmdBot = &cobra.Command{
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
		flagAdminID,
		flagDownloadFolder,
		bot.WithForwardTo(flagForwardTo),
		bot.WithDebugTelegram(flagDebug),
	)
}
