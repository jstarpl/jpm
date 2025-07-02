package main

import (
	"bufio"
	"embed"
	"fmt"
	"io/fs"
	"log"
	"time"

	"fyne.io/systray"
	"fyne.io/systray/example/icon"

	"github.com/alecthomas/kong"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/pkg/browser"
	"github.com/valyala/fasthttp"
)

//go:embed webgui/dist/*
var webgui embed.FS

var CLI struct {
	Service struct {
		Start struct {
			NoSystray bool   `name:"no-systray" help:"Do not show an icon in systray" default:"true"`
			Listen    string `name:"listen" help:"Address to listen for API connections." default:"127.0.0.1:3000"`
			Token     string `help:"Bearer Token to use to authorize API requests." default:"<random>"`
		} `cmd:"" help:"Start the service."`
		Stop struct{} `cmd:"" help:"Stop the service."`
	} `cmd:"" help:"Manage the JPM service."`

	Ps struct{} `cmd:"" help:"List running processes."`

	Start struct {
		Args []string `arg:""`
	} `cmd:"" help:"Start a new process."`

	Stop struct{} `cmd:"" help:"Stop specified process"`
}

func main() {
	ctx := kong.Parse(&CLI)
	switch ctx.Command() {
	case "service start":
		systray.Run(onReady, onExit)
	case "ps":
	default:
		panic(ctx.Error)
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
			browser.OpenURL("http://127.0.0.1:3000")
		}
	})()

	app := fiber.New(fiber.Config{
		AppName:      "JPM",
		ServerHeader: "JPM/0.1",
	})

	app.Use(recover.New())

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
					fmt.Printf("Error while flushing: %v. Closing http connection.\n", err)

					break
				}

				time.Sleep(2 * time.Second)
			}
		}))

		return nil
	})

	webguiDist, error := fs.Sub(webgui, "webgui/dist")

	if error != nil {
		panic("Could not get webgui/dist")
	}

	app.Use("/", static.New("", static.Config{
		Browse: true,
		FS:     webguiDist,
	}))

	log.Fatal(app.Listen("127.0.0.1:3000", fiber.ListenConfig{
		DisableStartupMessage: true,
	}))
}

func onExit() {
	// clean up here
}
