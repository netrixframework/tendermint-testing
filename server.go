package main

import "github.com/netrixframework/tendermint-testing/cmd"

func main() {
	c := cmd.RootCmd()
	c.Execute()
}
