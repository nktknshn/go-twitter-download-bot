package main

import "github.com/nktknshn/go-twitter-download-bot/cli"

func main() {
	if err := cli.CmdRoot.Execute(); err != nil {
		panic(err)
	}
}
