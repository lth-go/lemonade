package main

import (
	"fmt"
	"net/http"
	"os"
	"path"

	"github.com/ProtonMail/go-autostart"
	"github.com/getlantern/systray"
	"github.com/hanxi/lemonade/client"
	"github.com/hanxi/lemonade/icon"
	"github.com/hanxi/lemonade/lemon"
	"github.com/hanxi/lemonade/server"
	log "github.com/inconshreveable/log15"
	"github.com/skratchdot/open-golang/open"
)

var logLevelMap = map[int]log.Lvl{
	0: log.LvlDebug,
	1: log.LvlInfo,
	2: log.LvlWarn,
	3: log.LvlError,
	4: log.LvlCrit,
}

var srv *http.Server
var cli *lemon.CLI
var logger log.Logger

func getBinPath() string {
	e, err := os.Executable()
	if err != nil {
		panic(err)
	}
	path := path.Dir(e)
	return path
}

func onReady() {
	systray.SetIcon(icon.Data)
	systray.SetTitle("lemonade")
	tips := fmt.Sprintf("lemonade running.\nport:%d\nallow:'%s'", cli.Port, cli.Allow)
	systray.SetTooltip(tips)

	mChecked := systray.AddMenuItem("Auto Startup", "Auto Startup lemonade on boot")
	filename := os.Args[0] // get command line first parameter
	app := &autostart.App{
		Name:        "lemonade",
		DisplayName: "lemonade",
		Exec:        []string{filename},
	}
	if app.IsEnabled() {
		mChecked.Check()
	}

	mOpenConfig := systray.AddMenuItem("OpenConfig", "Open lemonade Config file")

	go func() {
		for {
			select {
			case <-mChecked.ClickedCh:
				if mChecked.Checked() {
					if err := app.Disable(); err != nil {
						logger.Error("Disable Autostart Failed.", "err", err)
					} else {
						mChecked.Uncheck()
					}
				} else {
					if err := app.Enable(); err != nil {
						logger.Error("Enable Autostart Failed.", "err", err)
					} else {
						mChecked.Check()
					}
				}
			case <-mOpenConfig.ClickedCh:
				{
					confPath, err := cli.GetConfPath()
					if err == nil {
						logger.Error("conf file path", "confPath", confPath)
						open.Run(confPath)
					}
				}
			}
		}
	}()

	mQuit := systray.AddMenuItem("Quit", "Quit lemonade")
	go func() {
		<-mQuit.ClickedCh
		systray.Quit()
	}()

	go startServer()
}

func onExit() {
	// clean up here
	stopServer()
}

func startServer() {
	logger.Debug("Starting Server")
	var err error
	srv, err = server.Serve(cli, logger)
	if err != nil {
		logger.Error("StartServer", "err", err)
	}
}

func stopServer() {
	if srv == nil {
		logger.Error("Server not running")
		return
	}
	if err := srv.Shutdown(nil); err != nil {
		panic(err) // failure/timeout shutting down the server gracefully
	}
	logger.Error("Server stoped.")
	srv = nil
}

func main() {
	cli = &lemon.CLI{
		In:  os.Stdin,
		Out: os.Stdout,
		Err: os.Stderr,
	}
	var err error
	if err = cli.FlagParse(os.Args, false); err != nil {
		writeError(cli, err)
		os.Exit(lemon.FlagParseError)
	}

	logger = log.New()
	logLevel := logLevelMap[cli.LogLevel]
	logger.SetHandler(log.LvlFilterHandler(logLevel, log.StdoutHandler))

	if len(os.Args) == 1 {
		fmt.Fprintln(cli.Err, lemon.Usage)
		systray.Run(onReady, onExit)
	} else {
		if cli.Type == lemon.SERVER {
			startServer()
		} else {
			lc := client.New(cli, logger)
			switch cli.Type {
			case lemon.OPEN:
				logger.Debug("Opening URL")
				err = lc.Open(cli.DataSource, cli.TransLocalfile, cli.TransLoopback)
			case lemon.COPY:
				logger.Debug("Copying text")
				err = lc.Copy(cli.DataSource)
			case lemon.PASTE:
				logger.Debug("Pasting text")
				var text string
				text, err = lc.Paste()
				cli.Out.Write([]byte(text))
			}
			if err != nil {
				writeError(cli, err)
				os.Exit(lemon.RPCError)
			}
		}
	}

	if cli.Help {
		fmt.Fprintln(cli.Err, lemon.Usage)
		os.Exit(lemon.Help)
	}
}

func writeError(c *lemon.CLI, err error) {
	fmt.Fprintln(c.Err, err.Error())
}
