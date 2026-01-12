package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"path"
	"runtime"
	"strconv"
	"syscall"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/storage"
	"github.com/joho/godotenv"
	"github.com/multiversx/mx-chain-core-go/core"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
	"github.com/multiversx/mx-chain-logger-go/file"
	"github.com/urfave/cli"
)

const (
	defaultLogsPath            = "logs"
	defaultDataPath            = "data"
	dbFile                     = "sqlite.db"
	logFilePrefix              = "epoch-proxy"
	logFileLifeSpanInSec       = 86400 // 24h
	logFileLifeSpanInMB        = 1024  // 1GB
	configFile                 = "./config.toml"
	envFile                    = "./.env"
	emailTemplateFile          = "./activation_email.html"
	emailChangeTemplateFile    = "./change_email.html"
	swaggerPath                = "./swagger/"
	envFileVarJwtKey           = "JWT_KEY"
	envFileVarInitialAdminUser = "INITIAL_ADMIN_USER"
	envFileVarInitialAdminPass = "INITIAL_ADMIN_PASSWORD"
	envFileVarInitialAdminKey  = "INITIAL_ADMIN_KEY"
	envFileVarSmtpHost         = "SMTP_HOST"
	envFileVarSmtpPort         = "SMTP_PORT"
	envFileVarSmtpFrom         = "SMTP_FROM"
	envFileVarSmtpPassword     = "SMTP_PASSWORD"
)

// appVersion should be populated at build time using ldflags
// Usage examples:
// Linux/macOS:
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

	envFileVars     = []string{envFileVarJwtKey, envFileVarInitialAdminUser, envFileVarInitialAdminPass, envFileVarInitialAdminKey, envFileVarSmtpHost, envFileVarSmtpPort, envFileVarSmtpFrom, envFileVarSmtpPassword}
	envFileContents = make(map[string]string)
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
	if fileLogging != nil {
		defer func() {
			_ = fileLogging.Close()
		}()
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

	err = readEnvFile(envFileContents)
	if err != nil {
		return err
	}

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

	sqlitePath := path.Join(workingDir, defaultDataPath, dbFile)
	sqliteWrapper, err := storage.NewSQLiteWrapper(sqlitePath)
	if err != nil {
		return err
	}

	defer func() {
		_ = sqliteWrapper.Close()
	}()

	err = ensureAdmin(sqliteWrapper)
	if err != nil {
		return err
	}

	if cfg.FreeAccount.ClearPeriodInSeconds == 0 {
		return fmt.Errorf("can not start as the config contains a 0 value for FreeAccount.ClearPeriodInSeconds")
	}

	keyCounter := common.NewKeyCounter()
	limitPeriod := time.Duration(cfg.FreeAccount.ClearPeriodInSeconds) * time.Second
	common.CronJobStarter(context.Background(), func() {
		log.Debug("Clearing the keys counters")
		keyCounter.Clear()
	}, limitPeriod)

	accessChecker, err := process.NewAccessChecker(
		sqliteWrapper,
		keyCounter,
		cfg.FreeAccount.MaxCalls,
	)
	if err != nil {
		return err
	}

	requestsProcessor, err := process.NewRequestsProcessor(
		hostFinder,
		accessChecker,
		sqliteWrapper,
		cfg.ClosedEndpoints,
	)
	if err != nil {
		return err
	}

	authenticator := api.NewJWTAuthenticator(envFileContents[envFileVarJwtKey])

	accessKeysHandler, err := api.NewAccessKeysHandler(sqliteWrapper, authenticator)
	if err != nil {
		return err
	}

	usersHandler, err := api.NewUsersHandler(sqliteWrapper, authenticator)
	if err != nil {
		return err
	}

	loginHandler, err := api.NewLoginHandler(sqliteWrapper, authenticator)
	if err != nil {
		return err
	}

	performanceHandler, err := api.NewPerformanceHandler(sqliteWrapper, authenticator)
	if err != nil {
		return err
	}

	smtpPort, err := strconv.Atoi(envFileContents[envFileVarSmtpPort])
	if err != nil {
		return fmt.Errorf("invalid SMTP port: %w", err)
	}
	emailSender := process.NewSmtpSender(process.ArgsSmtpSender{
		SmtpPort: smtpPort,
		SmtpHost: envFileContents[envFileVarSmtpHost],
		From:     envFileContents[envFileVarSmtpFrom],
		Password: envFileContents[envFileVarSmtpPassword],
	})

	if len(cfg.AppDomains.Backend) == 0 || len(cfg.AppDomains.Frontend) == 0 {
		return fmt.Errorf("the AppDomains section is not correctly configured in config.toml file")
	}

	emailTemplateBytes, err := os.ReadFile(emailTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read email template file: %w", err)
	}

	captchaWrapper := process.NewCaptchaWrapper()

	registrationHandler, err := api.NewRegistrationHandler(
		sqliteWrapper,
		emailSender,
		cfg.AppDomains,
		captchaWrapper,
		string(emailTemplateBytes),
	)
	if err != nil {
		return err
	}

	captchaHandler, err := api.NewCaptchaHandler(captchaWrapper)
	if err != nil {
		return err
	}

	changeEmailTemplateBytes, err := os.ReadFile(emailChangeTemplateFile)
	if err != nil {
		return fmt.Errorf("failed to read email change template file: %w", err)
	}

	userCredentialsHandler, err := api.NewUserCredentialsHandler(
		sqliteWrapper,
		emailSender,
		cfg.AppDomains,
		string(changeEmailTemplateBytes),
		authenticator,
	)
	if err != nil {
		return err
	}

	handlers := map[string]http.Handler{
		api.EndpointApiAccessKeys:         accessKeysHandler,
		api.EndpointApiAdminUsers:         usersHandler,
		api.EndpointApiLogin:              loginHandler,
		api.EndpointApiPerformance:        performanceHandler,
		api.EndpointApiRegister:           registrationHandler,
		api.EndpointApiActivate:           registrationHandler,
		api.EndpointApiChangePassword:     userCredentialsHandler,
		api.EndpointApiRequestEmailChange: userCredentialsHandler,
		api.EndpointApiConfirmEmailChange: userCredentialsHandler,
		api.EndpointCaptchaSingle:         http.HandlerFunc(captchaHandler.GenerateCaptchaHandler),
		api.EndpointCaptchaMultiple:       http.HandlerFunc(captchaHandler.ServeCaptchaImageHandler),
		api.EndpointAppInfo: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(map[string]string{
				"version": appVersion,
				"backend": cfg.AppDomains.Backend,
			})
		}),
		api.EndpointSwagger: http.StripPrefix(api.EndpointSwagger, http.FileServer(http.Dir(swaggerPath))),
		api.EndpointRoot:    http.RedirectHandler(api.EndpointSwagger, http.StatusFound),
		"*":                 requestsProcessor,
	}

	demuxer := process.NewDemuxer(handlers, nil)
	engine, err := api.NewAPIEngine(fmt.Sprintf(":%d", cfg.Port), api.CORSMiddleware(demuxer))
	if err != nil {
		return err
	}

	log.Info("Serving requests", "interface", engine.Address())

	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	<-sigs

	log.Info("application closing, calling Close on all subcomponents...")
	err = engine.Close()
	if sqliteWrapper != nil {
		err = sqliteWrapper.Close()
	}

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

func readEnvFile(m map[string]string) error {
	err := godotenv.Load(envFile)
	if err != nil {
		return err
	}

	for _, v := range envFileVars {
		val := os.Getenv(v)
		if len(val) == 0 {
			return fmt.Errorf("%s is not set in the .env file", v)
		}

		m[v] = val
	}

	return nil
}

func ensureAdmin(sqliteWrapper api.KeyAccessProvider) error {
	foundAdmin, err := checkAdminIsPresent(sqliteWrapper)
	if err != nil {
		return err
	}

	if foundAdmin {
		log.Info("admin user found, skipping adding new admin")
		return nil
	}

	log.Info("creating admin user from .env file")

	err = sqliteWrapper.AddUser(
		envFileContents[envFileVarInitialAdminUser],
		envFileContents[envFileVarInitialAdminPass],
		true,
		0,
		"premium",
		true,
		"")
	if err != nil {
		return err
	}

	return sqliteWrapper.AddKey(
		envFileContents[envFileVarInitialAdminUser],
		envFileContents[envFileVarInitialAdminKey])
}

func checkAdminIsPresent(sqliteWrapper api.KeyAccessProvider) (bool, error) {
	users, err := sqliteWrapper.GetAllUsers()
	if err != nil {
		return false, err
	}

	for _, userDetails := range users {
		if userDetails.IsAdmin {
			return true, nil
		}
	}

	return false, nil
}
