## Deeplinks Library

This library is used to create deep links for apps on Windows.

### Usage

```go
package main

import (
    "github.com/kaazedev/deeplink"
)

func main() {
    // Pass an scheme to the deeplink, and port for redirecting messages on already running app
    dl := deeplink.NewDeeplink("rocketapp", 5945)
	
    // Method to redirect messages to the app, this should be called before Register().
    dl.Prepare()
    
    // Register a scheme and set a callback function
    dl.Register(func(msg string) {
        println(msg)
    })
}
```