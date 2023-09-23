package main

import "github.com/kaazedev/deeplink"

func main() {
	dl := deeplink.NewDeeplink("resourcer", 8080)
	dl.Prepare()
	dl.Register(func(message string) {
		println("Message received: " + message)
	})

	for {
	}
}
