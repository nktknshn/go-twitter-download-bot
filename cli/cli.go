package cli

import (
	"github.com/nktknshn/go-twitter-download-bot/cli/bot"
	"github.com/nktknshn/go-twitter-download-bot/cli/twitter"
	"github.com/spf13/cobra"
)

func init() {
	CmdRoot.AddCommand(twitter.Cmd)
	CmdRoot.AddCommand(bot.Cmd)
}

var CmdRoot = &cobra.Command{
	Use:   "cli",
	Short: "cli is a command line interface for twitter",
}
