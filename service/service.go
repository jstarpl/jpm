package service

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"jstarpl/jpm/api"
	"log"
	"math/rand"
	"time"

	"fyne.io/systray"
	"fyne.io/systray/example/icon"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/pkg/browser"
	"github.com/valyala/fasthttp"

	ipc "github.com/james-barrow/golang-ipc"
)

//go:embed webgui/dist/*
var embedFS embed.FS
var webgui *fs.FS

func init() {
	webguiDist, err := fs.Sub(embedFS, "webgui/dist")
	webgui = &webguiDist

	if err != nil {
		panic("Could not get webgui/dist")
	}
}

const randomTokenLength = 32

type Service struct {
	Start struct {
		NoSystray bool   `name:"no-systray" help:"Do not show an icon in systray" default:"false"`
		Listen    string `name:"listen" help:"Address to listen for API connections." default:"127.0.0.1:3000"`
		Token     string `name:"token" help:"Bearer Token to use to authorize API requests." default:"<random>"`
	} `cmd:"" help:"Start the service."`
	Stop struct{} `cmd:"" help:"Stop the service."`
}

var config *Service

func StartService(cli *Service) {
	if cli.Start.Token == "<random>" {
		cli.Start.Token = generateRandomBase36(randomTokenLength)
	}

	config = cli

	if cli.Start.NoSystray {
		startHTTPServer()
	} else {
		systray.Run(onReady, onExit)
	}
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("JPM")
	systray.SetTooltip("2 of 2 apps working")
	mOpen := systray.AddMenuItem("Open JPM Console", "Open JPM management console in your Web Browser")
	systray.AddSeparator()
	mExit := systray.AddMenuItem("Exit", "Exit JPM and shut down all apps")

	go (func() {
		<-mExit.ClickedCh
		systray.Quit()
	})()

	go (func() {
		for {
			<-mOpen.ClickedCh
			browser.OpenURL(fmt.Sprintf("http://%s/#token=%s", config.Start.Listen, config.Start.Token))
		}
	})()

	startHTTPServer()
	startIPCServer()
}

func startHTTPServer() {
	app := fiber.New(fiber.Config{
		AppName:      "JPM",
		ServerHeader: "JPM/0.1",
		TrustProxyConfig: fiber.TrustProxyConfig{
			LinkLocal: false,
			Loopback:  false,
			Private:   false,
		},
	})

	app.Use(recover.New())

	app.Use(func(c fiber.Ctx) error {
		log.Default().Printf("%v %s %s \"%s %s %s\"", c.IP(), "-", "-", c.Method(), c.OriginalURL(), c.Protocol())

		return c.Next()
	})

	api := app.Group("/api")

	api.Get("/", func(c fiber.Ctx) error {
		fmt.Println("🥇 First handler")
		c.Status(200)
		c.Set("content-type", "text/html")
		return c.SendString("Hello World!")
	})

	api.Get("/events", func(c fiber.Ctx) error {
		c.Set("Content-Type", "text/event-stream")
		c.Set("Cache-Control", "no-cache")
		c.Set("Connection", "keep-alive")
		c.Set("Transfer-Encoding", "chunked")

		// c.Status(fiber.StatusOK).Context().SetBodyStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
		// 	fmt.Println("WRITER")
		// 	var i int
		// 	for {
		// 		i++

		// 		var msg string

		// 		// if there are messages that have been sent to the `/publish` endpoint
		// 		// then use these first, otherwise just send the current time
		// 		if len(sseMessageQueue) > 0 {
		// 			msg = fmt.Sprintf("%d - message recieved: %s", i, sseMessageQueue[0])
		// 			// remove the message from the buffer
		// 			sseMessageQueue = sseMessageQueue[1:]
		// 		} else {
		// 			msg = fmt.Sprintf("%d - the time is %v", i, time.Now())
		// 		}

		// 		fmt.Fprintf(w, "data: Message: %s\n\n", msg)
		// 		fmt.Println(msg)

		// 		err := w.Flush()
		// 		if err != nil {
		// 			// Refreshing page in web browser will establish a new
		// 			// SSE connection, but only (the last) one is alive, so
		// 			// dead connections must be closed here.
		// 			fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

		// 			break
		// 		}
		// 		time.Sleep(2 * time.Second)
		// 	}
		// }))

		c.Status(fiber.StatusOK).SendStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			for {
				var msg = fmt.Sprintf("the time is %v", time.Now())
				fmt.Fprintf(w, "data: %s\n\n", msg)

				err := w.Flush()
				if err != nil {
					log.Default().Printf("Error while flushing: %v. Closing http connection.", err)

					break
				}

				time.Sleep(2 * time.Second)
			}
		}))

		return nil
	})

	app.Use("/", static.New("", static.Config{
		Browse: true,
		FS:     *webgui,
	}))

	log.Default().Printf("JPM Console at http://%s", config.Start.Listen)

	go (func() {
		log.Fatal(app.Listen(config.Start.Listen, fiber.ListenConfig{
			DisableStartupMessage: true,
		}))
	})()
}

func startIPCServer() {
	server, err := ipc.StartServer("jpm-ipc", nil)
	if err != nil {
		panic("Could not open `jpm-ipc` IPC channel. Check if the service isn't already running.")
	}

	go (func() {
		for {
			data, err := server.Read()
			if err != nil {
				log.Default().Fatalf("Error reading from IPC: %v", err)
			}

			log.Default().Printf("Message received: %d Length %d %v", data.MsgType, len(data.Data), string(data.Data[:]))

			if data.MsgType > 0 {
				var e api.Request
				err = json.Unmarshal(data.Data, &e)

				if err != nil {
					log.Default().Printf("Unknown message received: %v", data.Data)
					continue
				}

				errorMsg, _ := api.NewErrorResponse(e.MsgID, int(api.MethodNotFound), "Method not found")
				server.Write(1, errorMsg)
			}
		}
	})()
}

func onExit() {
	// clean up here
}

const charset = "0123456789abcdefghijklmnopqrstuvwxyz"

// generateRandomBase36 returns a random string of the given length using base-36 characters.
func generateRandomBase36(length int) string {
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[rand.Intn(len(charset))]
	}
	return string(result)
}
