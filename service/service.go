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

	"fyne.io/systray"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/static"
	"github.com/pkg/browser"
	"github.com/valyala/fasthttp"

	ipc "github.com/james-barrow/golang-ipc"
)

const (
	logProps = log.Lmicroseconds | log.Ltime | log.Ldate | log.LUTC
)

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
		Logs      string `name:"logs" help:"Path where the output from processes should be put." default:"<none>"`
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
		log.Default().Printf("Shutdown requested from systray")
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
		return c.SendStatus(fiber.StatusNoContent)
	})

	apiRouter.Get("/processes", func(c fiber.Ctx) error {
		list := executor.ListProcesses()
		res := api.Response{Header: "2.0", Result: &api.ResponseResult{ProcessList: list}, MsgID: 0}

		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Status(fiber.StatusOK)

		return c.JSON(res)
	})

	apiRouter.Post("/processes/start", func(c fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNotImplemented)
	})

	apiRouter.Get("/processes/:id", func(c fiber.Ctx) error {
		list := executor.ListProcesses()

		c.Set(fiber.HeaderCacheControl, "no-cache")

		for _, process := range *list {
			if process.Id == c.Params("id") {
				res := api.Response{Header: "2.0", Result: &api.ResponseResult{Process: &process}, MsgID: 0}
				c.Status(fiber.StatusOK)

				return c.JSON(res)
			}
		}

		c.Status(fiber.StatusNotFound)

		res, _ := api.NewErrorResponse(0, 404, "Process not found")
		return c.Send(res)
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

	apiRouter.Get("/processes/:id/stdouterr", func(c fiber.Ctx) error {
		stdOutErrBroadcast, err := executor.GetProcessStdStreamRelay(c.Params("id"))
		if err != nil {
			c.Set(fiber.HeaderContentType, "application/json")
			c.Set(fiber.HeaderCacheControl, "no-cache")
			c.Status(fiber.StatusNotFound)

			res, _ := api.NewErrorResponse(0, 404, "Process not found")
			return c.Send(res)
		}

		c.Set(fiber.HeaderContentType, "text/event-stream")
		c.Set(fiber.HeaderCacheControl, "no-cache")
		c.Set(fiber.HeaderConnection, "keep-alive")

		l := stdOutErrBroadcast.Listener(1)

		c.Status(fiber.StatusOK).SendStreamWriter(fasthttp.StreamWriter(func(w *bufio.Writer) {
			for n := range l.Ch() {
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", n.StreamType, n.Data)

				err := w.Flush()
				if err != nil {
					logger.Printf("Error while flushing: %v. Closing http connection.", err)

					l.Close()
					break
				}
			}
		}))

		return nil
	})

	apiRouter.Post("/processes/:id/stdin", func(c fiber.Ctx) error {
		stdInBroadcast, err := executor.GetProcessStdStreamRelay(c.Params("id"))
		if err != nil {
			c.Set(fiber.HeaderContentType, "application/json")
			c.Set(fiber.HeaderCacheControl, "no-cache")
			c.Status(fiber.StatusNotFound)

			res, _ := api.NewErrorResponse(0, 404, "Process not found")
			return c.Send(res)
		}

		stdInBroadcast.Broadcast(api.StdStreamMessage{
			Data: c.BodyRaw(),
		})

		c.Status(fiber.StatusOK)
		res, _ := api.NewSuccessResponse(0, &api.ResponseResult{Success: stringPtr("Sent")})
		return c.Send(res)
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
					proc, err := executor.StartProcess(params.Name, params.Namespace, params.Exec, params.Arg, params.Dir, params.Env)
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
				case api.RequestStopService:
					res, _ := api.NewSuccessResponse(e.MsgID, &api.ResponseResult{Success: stringPtr("Shutting down")})
					server.Write(api.MsgType, res)

					log.Default().Fatalf("Shutdown requested over IPC")
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
