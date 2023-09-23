package deeplink

import (
	"fmt"
	"golang.org/x/sys/windows/registry"
	"log"
	"net"
	"os"
	"path/filepath"
	"strings"
)

type DeepLink struct {
	// Scheme is the name of the scheme to register, for example "myapp" will be accessible via "myapp://"
	Scheme string

	// Host is the host to listen on, used for checking is the app is already running
	Port uint

	// OnMessage is called when a message is received, the message is passed as a string
	OnMessage func(string)
}

// NewDeeplink creates a new DeepLink instance
// scheme is the name of the scheme to register, for example "myapp" will be accessible via "myapp://"
func NewDeeplink(scheme string, port uint) *DeepLink {
	return &DeepLink{
		Scheme: scheme,
		Port:   port,
	}
}

// Register registers the scheme in the registry, this is required for the scheme to work
func (dl *DeepLink) Register(callback func(string)) (bool, error) {
	exe, _ := os.Executable()
	exePath := filepath.ToSlash(exe)

	// reformat exePath to windows format
	exePath = strings.Replace(exePath, "/", "\\", -1)

	k, err := registry.OpenKey(registry.CURRENT_USER, `Software\Classes\`+dl.Scheme, registry.ALL_ACCESS)
	if err != nil {
		k, _, err = registry.CreateKey(registry.CURRENT_USER, `Software\Classes\`+dl.Scheme, registry.ALL_ACCESS)
		if err != nil {
			log.Fatal("registry.CreateKey", err)
		}
	}
	defer k.Close()

	err = k.SetStringValue("", "URL:"+dl.Scheme)
	if err != nil {
		log.Fatal("k.SetStringValue", err)
	}

	err = k.SetStringValue("URL Protocol", "")
	if err != nil {
		log.Fatal("k.SetStringValue URL Protocol", err)
	}

	k2, _, err := registry.CreateKey(k, `DefaultIcon`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal("registry.CreateKey DefaultIcon", err)
	}

	err = k2.SetStringValue("", fmt.Sprintf("%s,0", exePath))
	if err != nil {
		log.Fatal("k2.SetStringValue", err)
	}

	k2.Close()

	k3, _, err := registry.CreateKey(k, `shell\open\command`, registry.ALL_ACCESS)
	if err != nil {
		log.Fatal("registry.CreateKey shell\\open\\command", err)
	}

	err = k3.SetStringValue("", fmt.Sprintf("%s \"%%1\"", exePath))
	if err != nil {
		log.Fatal("k3.SetStringValue", err)
	}

	k3.Close()

	dl.OnMessage = callback

	return true, nil
}

// Unregister unregisters the scheme from the registry
func (dl *DeepLink) Unregister() (bool, error) {
	err := registry.DeleteKey(registry.CURRENT_USER, `Software\Classes\`+dl.Scheme)
	if err != nil {
		log.Fatal("registry.DeleteKey", err)
	}

	return true, nil

}

// Prepare prepares the DeepLink instance for receiving messages, this is required to redirect message to already running app
func (dl *DeepLink) Prepare() {
	go func() {
		listener, err := net.Listen("tcp", fmt.Sprintf("127.0.0.1:%d", dl.Port))
		if err != nil {
			if strings.Contains(err.Error(), "Only one usage of each socket address") {
				conn, err := net.Dial("tcp", fmt.Sprintf("127.0.0.1:%d", dl.Port))
				if err != nil {
					os.Exit(1)
				}

				conn.Write([]byte(os.Args[len(os.Args)-1]))
				conn.Close()

				os.Exit(0)
			}
		}

		// listen for incoming connections
		for {
			message, err := listener.Accept()
			if err != nil {
				os.Exit(0)
			}

			buf := make([]byte, 1024)
			n, err := message.Read(buf)
			if err != nil {
				os.Exit(0)
			}

			var msg string
			if n > 0 {
				msg = string(buf[:n])

				if dl.OnMessage != nil {
					dl.OnMessage(msg)
				}
			}
		}
	}()
}
