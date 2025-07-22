package service

import (
	"bufio"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"jstarpl/jpm/api"
	"jstarpl/jpm/service/executor"
	"log"
	"math/rand"
	"runtime"
	"time"

	"fyne.io/systray"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/pkg/browser"
	"github.com/valyala/fasthttp"

	ipc "github.com/james-barrow/golang-ipc"
)

const logProps = log.Lmicroseconds | log.Ltime | log.Ldate | log.LUTC

//go:embed icons/icon.ico
var iconIco []byte

//go:embed icons/icon.svg
var iconSvg []byte

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
		run()
	} else {
		systray.Run(onReady, onExit)
	}
}

func run() {
	startIPCServer()
	startHTTPServer()
}

func onReady() {
	if runtime.GOOS == "windows" {
		systray.SetIcon(iconIco)
	} else {
		systray.SetIcon(iconSvg)
	}

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

	run()
}

func startHTTPServer() {
	logger := log.New(log.Default().Writer(), "http: ", logProps)

	app := fiber.New(fiber.Config{
		ServerHeader: "JPM/0.1",
		TrustProxyConfig: fiber.TrustProxyConfig{
			LinkLocal: false,
			Loopback:  false,
			Private:   false,
		},
	})

	app.Use(recover.New())

	app.Use(func(c fiber.Ctx) error {
		logger.Printf("%v %s %s \"%s %s %s\"", c.IP(), "-", "-", c.Method(), c.OriginalURL(), c.Protocol())

		return c.Next()
	})

	apiRouter := app.Group("/api")

	apiRouter.Use(func(c fiber.Ctx) error {
		if config.Start.Token != "" {
			headers := c.GetReqHeaders()
			if len(headers[fiber.HeaderAuthorization]) == 0 {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
			auth := headers[fiber.HeaderAuthorization][0]
			if auth != fmt.Sprintf("Bearer %s", config.Start.Token) {
				return c.SendStatus(fiber.StatusUnauthorized)
			}
		}

		return c.Next()
	})

	apiRouter.Get("/", func(c fiber.Ctx) error {
		fmt.Println("🥇 First handler")
		c.Status(fiber.StatusOK)
		c.Set(fiber.HeaderContentType, "text/html")
		return c.SendString("Hello World!")
	})

	apiRouter.Get("/processes", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "application/json")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Status(fiber.StatusOK)
		list := executor.ListProcesses()
		res := api.Response{Header: "2.0", Result: &api.ResponseResult{ProcessList: list}, MsgID: 0}
		data, _ := json.Marshal(res)
		return c.Send(data)
	})

	apiRouter.Post("/processes/start", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Get("/processes/:id", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Post("/processes/:id/stop", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Post("/processes/:id/restart", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Delete("/processes/:id", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Patch("/processes/:id", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Get("/events", func(c fiber.Ctx) error {
		c.Set(fiber.HeaderContentType, "text/event-stream")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set(fiber.HeaderConnection, "keep-alive")

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
					logger.Printf("Error while flushing: %v. Closing http connection.", err)

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

	logger.Printf("JPM Console at http://%s", config.Start.Listen)

	go (func() {
		log.Fatal(app.Listen(config.Start.Listen, fiber.ListenConfig{
			DisableStartupMessage: true,
		}))
	})()
}

func startIPCServer() {
	server, err := ipc.StartServer(api.IPCName, nil)
	if err != nil {
		panic("Could not open `jpm-ipc` IPC channel. Check if the service isn't already running.")
	}

	logger := log.New(log.Default().Writer(), "ipc: ", logProps)

	go (func() {
		for {
			data, err := server.Read()
			if err != nil {
				logger.Fatalf("Error reading from IPC: %v", err)
			}

			// logger.Printf("Message received: %d %v Length %d %v", data.MsgType, data.Status, len(data.Data), string(data.Data[:]))

			if data.MsgType > 0 {
				var e api.Request
				err = json.Unmarshal(data.Data, &e)

				if err != nil {
					logger.Printf("Unknown message received: %v", data.Data)
					errorMsg, _ := api.NewErrorResponse(e.MsgID, int(api.ParseError), "Parse error")
					server.Write(api.MsgType, errorMsg)
					continue
				}

				logger.Printf("Method requested %s", e.Method)
				switch e.Method {
				case api.ListProcesses:
					list := executor.ListProcesses()
					res, _ := api.NewSuccessResponse(e.MsgID, &api.ResponseResult{
						ProcessList: list,
					})
					server.Write(api.MsgType, res)
				case api.StartProcess:
					var params api.RequestStartProcessParams
					json.Unmarshal(e.Params, &params)
					proc, err := executor.StartProcess(params.Name, params.Exec, params.Arg, params.Dir, params.Env)
					if err != nil {
						res, _ := api.NewErrorResponse(e.MsgID, 501, fmt.Sprintf("Could not start process: %v", err))
						server.Write(api.MsgType, res)
						continue
					}

					res, _ := api.NewSuccessResponse(e.MsgID, &api.ResponseResult{Success: stringPtr("Process started"), ProcessId: &proc.Id})
					server.Write(api.MsgType, res)
				case api.StopProcess:
					var params api.RequestStopProcessParams
					json.Unmarshal(e.Params, &params)
					err := executor.StopProcess(params.Id)
					if err != nil {
						res, _ := api.NewErrorResponse(e.MsgID, 501, fmt.Sprintf("Could not stop process: %v", err))
						server.Write(api.MsgType, res)
						continue
					}

					res, _ := api.NewSuccessResponse(e.MsgID, &api.ResponseResult{Success: stringPtr("Process stopped")})
					server.Write(api.MsgType, res)
				case api.DeleteProcess:
					var params api.RequestStopProcessParams
					json.Unmarshal(e.Params, &params)
					err := executor.DeleteProcess(params.Id)
					if err != nil {
						res, _ := api.NewErrorResponse(e.MsgID, 501, fmt.Sprintf("Could not delete process: %v", err))
						server.Write(api.MsgType, res)
						continue
					}

					res, _ := api.NewSuccessResponse(e.MsgID, &api.ResponseResult{Success: stringPtr("Process deleted")})
					server.Write(api.MsgType, res)
				default:
					errorMsg, _ := api.NewErrorResponse(e.MsgID, int(api.MethodNotFound), "Method not found")
					server.Write(api.MsgType, errorMsg)
				}
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

func stringPtr(s string) *string {
	return &s
}
