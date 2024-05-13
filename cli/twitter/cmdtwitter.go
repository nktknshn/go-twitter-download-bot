package twitter

import (
	"fmt"

	"github.com/nktknshn/go-twitter-download-bot/twitter"
	"github.com/spf13/cobra"
)

var (
	flagSaveData bool
)

func init() {
	Cmd.AddCommand(cmdGetTokens)
	Cmd.AddCommand(cmdGetData)

	cmdGetData.PersistentFlags().BoolVarP(&flagSaveData, "save-data", "s", false, "save data to file")
}

var (
	Cmd = &cobra.Command{
		Use:   "twitter",
		Short: "twitter is a command line interface for twitter",
		Args:  cobra.MinimumNArgs(1),
	}
	cmdGetTokens = &cobra.Command{
		Use:   "get-tokens",
		Short: "get-tokens <url>",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetTokens,
	}
	cmdGetData = &cobra.Command{
		Use:   "get-data",
		Short: "get-data <url>",
		Args:  cobra.ExactArgs(1),
		RunE:  runGetData,
	}
)

func runGetTokens(cmd *cobra.Command, args []string) error {
	twitter := twitter.NewTwitter()
	bt, err := twitter.GetTokens(cmd.Context(), args[0])
	if err != nil {
		return err
	}
	fmt.Printf("Bearer %s", bt)

	return nil
}

func runGetData(cmd *cobra.Command, args []string) error {
	tw := twitter.NewTwitter(
		twitter.WithSaveData(flagSaveData),
	)

	td, err := tw.GetTwitterData(cmd.Context(), args[0])

	if err != nil {
		return err
	}
	fmt.Println(td)
	return nil
}
