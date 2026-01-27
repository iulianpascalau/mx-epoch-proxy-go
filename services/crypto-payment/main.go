package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"path"
	"path/filepath"
	"runtime"
	"syscall"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/crypto-payment/factory"
	"github.com/joho/godotenv"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/multiversx/mx-sdk-go/blockchain"
	sdkCore "github.com/multiversx/mx-sdk-go/core"
	"github.com/multiversx/mx-sdk-go/interactors"
	"github.com/pelletier/go-toml"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath       = "logs"
	defaultDataPath       = "data"
	dbFile                = "sqlite.db"
	logFilePrefix         = "crypto-payment"
	logFileLifeSpanInSec  = 86400 // 24h
	logFileLifeSpanInMB   = 1024  // 1GB
	envFile               = "./.env"
	pemFilesSearchPattern = "*.pem"
)

// appVersion should be populated at build time using ldflags
var appVersion = "undefined"

var (
	helpTemplate = `NAME:
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

	log = logger.GetOrCreate("crypto-payment")

	// logLevel defines the logger level
	logLevel = cli.StringFlag{
		Name: "log-level",
		Usage: "This flag specifies the logger `level(s)`. It can contain multiple comma-separated value. For example" +
			", if set to *:INFO the logs for all packages will have the INFO level.",
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
	cli.AppHelpTemplate = helpTemplate
	app.Name = "Crypto Payment Service"
	app.Version = fmt.Sprintf("%s/%s/%s-%s", appVersion, runtime.Version(), runtime.GOOS, runtime.GOARCH)
	app.Usage = "Entry point for the Crypto Payment Service"
	app.Flags = []cli.Flag{
		logLevel,
		logSaveFile,
		workingDirectory,
	}
	app.Authors = []cli.Author{
		{
			Name:  "Iulian Pascalau",
			Email: "iulian.pascalau@gmail.com",
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
	if fileLogging != nil {
		defer func() {
			_ = fileLogging.Close()
		}()
	}

	if fileLogging != nil {
		timeLogLifeSpan := time.Second * time.Duration(logFileLifeSpanInSec)
		sizeLogLifeSpanInMB := uint64(logFileLifeSpanInMB)
		err = fileLogging.ChangeFileLifeSpan(timeLogLifeSpan, sizeLogLifeSpanInMB)
		if err != nil {
			return err
		}
	}

	log.Info("starting crypto-payment service", "version", appVersion, "pid", os.Getpid())

	err = godotenv.Load(envFile)
	if err != nil {
		log.Warn("load env file", "error", err)
	}

	mnemonics := os.Getenv("MNEMONICS")
	if len(mnemonics) == 0 {
		return fmt.Errorf("missing MNEMONICS environment variable")
	}

	sqlitePath := path.Join(workingDir, defaultDataPath, dbFile)

	cfg, err := loadConfig(workingDir)
	if err != nil {
		return err
	}

	proxyArgs := blockchain.ArgsProxy{
		ProxyURL:            cfg.ProxyURL,
		SameScState:         false,
		ShouldBeSynced:      false,
		FinalityCheck:       true,
		AllowedDeltaToFinal: 7,
		CacheExpirationTime: time.Second * 600,
		EntityType:          sdkCore.Proxy,
	}
	proxy, err := blockchain.NewProxy(proxyArgs)
	if err != nil {
		return err
	}

	relayersKeys, err := loadPemFiles(workingDir)
	if err != nil {
		return err
	}

	components, err := factory.NewComponentsHandler(mnemonics, sqlitePath, proxy, *cfg, relayersKeys)
	if err != nil {
		return err
	}
	defer components.Close()

	time.Sleep(time.Second)

	ctxRun, cancel := context.WithCancel(context.Background())
	defer cancel()

	components.StartCronJobs(ctxRun)

	log.Info("Service is running... Press Ctrl+C to stop")

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing")

	return nil
}

// FileLoggingHandler interface for the logger
type FileLoggingHandler interface {
	ChangeFileLifeSpan(newDuration time.Duration, newSizeInMB uint64) error
	Close() error
}

func attachFileLogger(log logger.Logger, saveLogFile bool, workingDir string) (FileLoggingHandler, error) {
	var fileLogging FileLoggingHandler
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

func loadConfig(workingDir string) (*config.Config, error) {
	configFile := filepath.Join(workingDir, "config.toml")
	_, err := os.Stat(configFile)
	if os.IsNotExist(err) {
		return nil, fmt.Errorf("config file not found: %s", configFile)
	}

	var cfg config.Config
	tree, err := toml.LoadFile(configFile)
	if err != nil {
		return nil, err
	}
	err = tree.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func loadPemFiles(workingDir string) ([][]byte, error) {
	pemFiles, err := filepath.Glob(filepath.Join(workingDir, pemFilesSearchPattern))
	if err != nil {
		return nil, err
	}

	allPemBytes := make([][]byte, 0, len(pemFiles))
	wallet := interactors.NewWallet()
	for _, pemFile := range pemFiles {
		pemBytes, errRead := wallet.LoadPrivateKeyFromPemFile(pemFile)
		if errRead != nil {
			return nil, fmt.Errorf("%w for file %s", errRead, pemFile)
		}

		log.Info("loaded pem file", "filename", pemFile)

		allPemBytes = append(allPemBytes, pemBytes)
	}

	return allPemBytes, nil
}
