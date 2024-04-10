package main

import "github.com/nktknshn/go-twitter-fun/cli"

func main() {
	if err := cli.CmdRoot.Execute(); err != nil {
		panic(err)
	}
}
