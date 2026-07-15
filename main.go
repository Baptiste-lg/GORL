package main

import "github.com/Baptiste-lg/GORL/game"

func main() {
	g := game.New()
	g.Run()

	// Block forever to keep the Go runtime alive in the browser.
	select {}
}
