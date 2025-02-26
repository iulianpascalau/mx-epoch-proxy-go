package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/process"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath      = "logs"
	logFilePrefix        = "epoch-proxy"
	logFileLifeSpanInSec = 86400 // 24h
	logFileLifeSpanInMB  = 1024  // 1GB
	configFile           = "config/config.toml"
	swaggerPath          = "./swagger/"
)

// appVersion should be populated at build time using ldflags
// Usage examples:
// linux/mac:
//
//	go build -v -ldflags="-X main.appVersion=$(git describe --all | cut -c7-32)
var appVersion = "undefined"

var (
	proxyHelpTemplate = `NAME:
   {{.Name}} - {{.Usage}}
USAGE:
   {{.HelpName}} {{if .VisibleFlags}}[global options]{{end}}
   {{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}{{if .Commands}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}
VERSION:
   {{.Version}}
   {{end}}
`

	log = logger.GetOrCreate("proxy")

	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level. However, if set to *:INFO,api:DEBUG" +
			" the logs for all packages will have the INFO level, excepting the api package which will receive a DEBUG" +
			" log level.",
		Value: "*:" + logger.LogInfo.String(),
	}
	// logFile is used when the log output needs to be logged in a file
	logSaveFile = cli.BoolFlag{
		Name:  "log-save",
		Usage: "Boolean option for enabling log saving. If set, it will automatically save all the logs into a file.",
	}
	// workingDirectory defines a flag for the path for the working directory.
	workingDirectory = cli.StringFlag{
		Name:  "working-directory",
		Usage: "This flag specifies the `directory` where the node will store databases and logs.",
		Value: "",
	}
)

func main() {
	app := cli.NewApp()
	cli.AppHelpTemplate = proxyHelpTemplate
	app.Name = "Multiversx Epoch Proxy CLI App"
	app.Version = fmt.Sprintf("%s/%s/%s-%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	app.Usage = "This is the entry point for starting a new Multiversx epoch proxy"
	app.Flags = []cli.Flag{
		logLevel,
		logSaveFile,
		workingDirectory,
	}
	app.Authors = []cli.Author{
		{
			Name:  "The Multiversx Team",
			Email: "contact@multiversx.com",
		},
	}

	app.Action = run

	err := app.Run(os.Args)
	if err != nil {
		log.Error(err.Error())
		os.Exit(1)
	}
}

func run(ctx *cli.Context) error {
	saveLogFile := ctx.GlobalBool(logSaveFile.Name)
	workingDir := ctx.GlobalString(workingDirectory.Name)

	err := logger.SetLogLevel(ctx.GlobalString(logLevel.Name))
	if err != nil {
		return err
	}

	fileLogging, err := attachFileLogger(log, saveLogFile, workingDir)
	if err != nil {
		return err
	}

	if !check.IfNil(fileLogging) {
		timeLogLifeSpan := time.Second * time.Duration(logFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(logFileLifeSpanInMB)
		err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	log.Info("starting epoch proxy", "version", appVersion, "pid", os.Getpid())

	cfg, err := loadConfig(configFile)
	if err != nil {
		return err
	}

	hostFinder, err := process.NewHostsFinder(cfg.Gateways)
	if err != nil {
		return err
	}

	loadedGateways := hostFinder.LoadedGateways()
	for _, host := range loadedGateways {
		log.Info("Loaded gateway",
			"name", host.Name,
			"URL", host.URL,
			"nonces", fmt.Sprintf("%s - %s", host.NonceStart, host.NonceEnd),
			"epochs", fmt.Sprintf("%s - %s", host.EpochStart, host.EpochEnd),
		)
	}

	tester := api.NewGatewayTester()
	err = tester.TestGateways(loadedGateways)
	if err != nil {
		return err
	}

	accessChecker, err := process.NewAccessChecker(cfg.AccessKeys)
	if err != nil {
		return err
	}

	requestsProcessor, err := process.NewRequestsProcessor(
		hostFinder,
		accessChecker,
		cfg.ClosedEndpoints,
	)
	if err != nil {
		return err
	}

	handlers := map[string]http.Handler{
		"*": requestsProcessor,
	}

	fs := http.FS(os.DirFS(swaggerPath))
	demuxer := process.NewDemuxer(handlers, http.FileServer(fs))
	engine, err := api.NewAPIEngine(fmt.Sprintf(":%d", cfg.Port), demuxer)
	if err != nil {
		return err
	}

	log.Info("Serving requests", "interface", engine.Address())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on all subcomponents...")
	err = engine.Close()

	return err
}

func attachFileLogger(log logger.Logger, saveLogFile bool, workingDir string) (common.FileLoggingHandler, error) {
	var fileLogging common.FileLoggingHandler
	var err error
	if saveLogFile {
		argsFileLogging := file.ArgsFileLogging{
			WorkingDir:      workingDir,
			DefaultLogsPath: defaultLogsPath,
			LogFilePrefix:   logFilePrefix,
		}
		fileLogging, err = file.NewFileLogging(argsFileLogging)
		if err != nil {
			return nil, fmt.Errorf("%w creating a log file", err)
		}
	}

	err = logger.SetDisplayByteSlice(logger.ToHex)
	log.LogIfError(err)

	return fileLogging, nil
}

func loadConfig(filepath string) (config.Config, error) {
	cfg := config.Config{}
	err := core.LoadTomlFile(&cfg, filepath)
	if err != nil {
		return config.Config{}, err
	}

	return cfg, nil
}
