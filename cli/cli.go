package cli

import (
	"github.com/nktknshn/go-twitter-fun/cli/cmdbot"
	"github.com/nktknshn/go-twitter-fun/cli/cmdtwitter"
	"github.com/spf13/cobra"
)

func init() {
	CmdRoot.AddCommand(cmdtwitter.CmdTwitter)
	CmdRoot.AddCommand(cmdbot.CmdBot)
}

var CmdRoot = &cobra.Command{
	Use:   "cli",
	Short: "cli is a command line interface for twitter",
}
