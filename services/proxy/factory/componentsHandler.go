package factory

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/api"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/common"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/config"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/process"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/serviceWrappers"
	"github.com/iulianpascalau/mx-epoch-proxy-go/services/proxy/storage"
	"github.com/multiversx/mx-chain-core-go/core/check"
	logger "github.com/multiversx/mx-chain-logger-go"
)

var log = logger.GetOrCreate("factory")

type componentsHandler struct {
	config               config.Config
	hostFinder           process.HostFinder
	tester               GatewayTester
	countersCache        storage.CountersCache
	sqliteWrapper        SQLiteWrapper
	keyCounter           process.KeyCounter
	accessChecker        process.AccessChecker
	requestsProcessor    RequestsProcessor
	jwtAuthenticator     api.Authenticator
	emailSender          api.EmailSender
	cryptoPaymentClient  api.CryptoPaymentClient
	requestsSynchronizer RequestsSynchronizer
	apiEngine            APIEngine
	captchaWrapper       api.CaptchaHandler

	accessKeysHandler      http.Handler
	usersHandler           http.Handler
	loginHandler           http.Handler
	performanceHandler     http.Handler
	registrationHandler    http.Handler
	captchaHandler         CaptchaHTTPHandler
	userCredentialsHandler http.Handler
	cryptoPaymentHandler   http.Handler
	demuxer                http.Handler
}

// NewComponentsHandler creates a new instance of the components handler holding all high-level components
func NewComponentsHandler(
	cfg config.Config,
	sqlitePath string,
	jwtKey string,
	emailsConfig config.EmailsConfig,
	appVersion string,
	swaggerPath string,
	emailSender api.EmailSender,
	captchaWrapper api.CaptchaHandler,
) (*componentsHandler, error) {
	if cfg.FreeAccount.ClearPeriodInSeconds == 0 {
		return nil, fmt.Errorf("can not start as the config contains a 0 value for FreeAccount.ClearPeriodInSeconds")
	}
	if len(cfg.AppDomains.Backend) == 0 || len(cfg.AppDomains.Frontend) == 0 {
		return nil, fmt.Errorf("the AppDomains section is not correctly configured in config.toml file")
	}
	if cfg.UpdateContractDBInSeconds == 0 {
		return nil, fmt.Errorf("can not start as the config contains a 0 value for UpdateContractDBInSeconds")
	}
	if check.IfNil(emailSender) {
		return nil, errNilEmailSender
	}
	if check.IfNil(captchaWrapper) {
		return nil, errNilCaptchaWrapper
	}

	ch := &componentsHandler{
		config:         cfg,
		emailSender:    emailSender,
		captchaWrapper: captchaWrapper,
	}
	var err error
	defer func() {
		if err != nil {
			ch.Close()
		}
	}()

	ch.hostFinder, err = process.NewHostsFinder(cfg.Gateways)
	if err != nil {
		return nil, err
	}

	loadedGateways := ch.hostFinder.LoadedGateways()
	for _, host := range loadedGateways {
		log.Info("Loaded gateway",
			"name", host.Name,
			"URL", host.URL,
			"nonces", fmt.Sprintf("%s - %s", host.NonceStart, host.NonceEnd),
			"epochs", fmt.Sprintf("%s - %s", host.EpochStart, host.EpochEnd),
		)
	}

	ch.tester = api.NewGatewayTester()
	err = ch.tester.TestGateways(loadedGateways)
	if err != nil {
		return nil, err
	}

	ch.countersCache, err = storage.NewCountersCache(time.Duration(cfg.CountersCacheTTLInSeconds) * time.Second)
	if err != nil {
		return nil, err
	}

	ch.sqliteWrapper, err = storage.NewSQLiteWrapper(
		sqlitePath,
		ch.countersCache,
	)
	if err != nil {
		return nil, err
	}

	ch.keyCounter = common.NewKeyCounter()

	ch.accessChecker, err = process.NewAccessChecker(
		ch.sqliteWrapper,
		ch.keyCounter,
		cfg.FreeAccount.MaxCalls,
	)
	if err != nil {
		return nil, err
	}

	ch.requestsProcessor, err = process.NewRequestsProcessor(
		ch.hostFinder,
		ch.accessChecker,
		ch.sqliteWrapper,
		cfg.ClosedEndpoints,
	)
	if err != nil {
		return nil, err
	}

	ch.jwtAuthenticator = api.NewJWTAuthenticator(jwtKey)

	ch.accessKeysHandler, err = api.NewAccessKeysHandler(ch.sqliteWrapper, ch.jwtAuthenticator)
	if err != nil {
		return nil, err
	}

	ch.usersHandler, err = api.NewUsersHandler(ch.sqliteWrapper, ch.jwtAuthenticator)
	if err != nil {
		return nil, err
	}

	ch.loginHandler, err = api.NewLoginHandler(ch.sqliteWrapper, ch.jwtAuthenticator)
	if err != nil {
		return nil, err
	}

	ch.performanceHandler, err = api.NewPerformanceHandler(ch.sqliteWrapper, ch.jwtAuthenticator)
	if err != nil {
		return nil, err
	}

	ch.registrationHandler, err = api.NewRegistrationHandler(
		ch.sqliteWrapper,
		ch.emailSender,
		cfg.AppDomains,
		ch.captchaWrapper,
		string(emailsConfig.RegistrationEmailBytes),
	)
	if err != nil {
		return nil, err
	}

	ch.captchaHandler, err = api.NewCaptchaHandler(ch.captchaWrapper)
	if err != nil {
		return nil, err
	}

	ch.userCredentialsHandler, err = api.NewUserCredentialsHandler(
		ch.sqliteWrapper,
		ch.emailSender,
		cfg.AppDomains,
		string(emailsConfig.ChangeEmailBytes),
		ch.jwtAuthenticator,
	)
	if err != nil {
		return nil, err
	}

	httpRequester := process.NewHttpRequester(time.Duration(cfg.CryptoPayment.TimeoutInSeconds) * time.Second)
	ch.cryptoPaymentClient, err = serviceWrappers.NewCryptoPaymentClient(httpRequester, cfg.CryptoPayment)
	if err != nil {
		return nil, err
	}

	ch.requestsSynchronizer, err = process.NewRequestsSynchronizer(ch.sqliteWrapper, ch.cryptoPaymentClient)
	if err != nil {
		return nil, err
	}

	mutexManager := process.NewUserMutexManager()
	ch.cryptoPaymentHandler, err = api.NewCryptoPaymentHandler(
		ch.cryptoPaymentClient,
		ch.sqliteWrapper,
		ch.jwtAuthenticator,
		mutexManager,
	)
	if err != nil {
		return nil, err
	}

	handlers := map[string]http.Handler{
		api.EndpointApiAccessKeys:         ch.accessKeysHandler,
		api.EndpointApiAdminUsers:         ch.usersHandler,
		api.EndpointApiLogin:              ch.loginHandler,
		api.EndpointApiPerformance:        ch.performanceHandler,
		api.EndpointApiRegister:           ch.registrationHandler,
		api.EndpointApiActivate:           ch.registrationHandler,
		api.EndpointApiChangePassword:     ch.userCredentialsHandler,
		api.EndpointApiRequestEmailChange: ch.userCredentialsHandler,

		api.EndpointApiConfirmEmailChange:         ch.userCredentialsHandler,
		api.EndpointApiCryptoPaymentConfig:        ch.cryptoPaymentHandler,
		api.EndpointApiCryptoPaymentCreateAddress: ch.cryptoPaymentHandler,
		api.EndpointApiCryptoPaymentAccount:       ch.cryptoPaymentHandler,
		api.EndpointApiAdminCryptoPaymentAccount:  ch.cryptoPaymentHandler,
		api.EndpointCaptchaSingle:                 http.HandlerFunc(ch.captchaHandler.GenerateCaptchaHandler),
		api.EndpointCaptchaMultiple:               http.HandlerFunc(ch.captchaHandler.ServeCaptchaImageHandler),
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
		"*":                 ch.requestsProcessor,
	}

	ch.demuxer = process.NewDemuxer(handlers, nil)
	ch.apiEngine, err = api.NewAPIEngine(fmt.Sprintf(":%d", cfg.Port), api.CORSMiddleware(ch.demuxer))
	if err != nil {
		return nil, err
	}

	return ch, nil
}

// StartCronJobs starts all defined cron jobs
func (ch *componentsHandler) StartCronJobs(ctx context.Context) {
	if ch == nil {
		return
	}

	common.CronJobStarter(ctx, func() {
		log.Debug("Sweeping the counters cache")
		ch.countersCache.Sweep()
	}, time.Duration(ch.config.CountersCacheTTLInSeconds)*time.Second)

	limitPeriod := time.Duration(ch.config.FreeAccount.ClearPeriodInSeconds) * time.Second
	common.CronJobStarter(ctx, func() {
		log.Debug("Clearing the keys counters")
		ch.keyCounter.Clear()
	}, limitPeriod)

	common.CronJobStarter(ctx, func() {
		log.Debug("Synchronizing user max requests")
		ch.requestsSynchronizer.Process()
	}, time.Duration(ch.config.UpdateContractDBInSeconds)*time.Second)

}

// GetSQLiteWrapper returns the SQLiteWrapper instance
func (ch *componentsHandler) GetSQLiteWrapper() api.KeyAccessProvider {
	return ch.sqliteWrapper
}

// GetAPIEngine returns the APIEngine instance
func (ch *componentsHandler) GetAPIEngine() APIEngine {
	return ch.apiEngine
}

// Close closes all the components held by the handler
func (ch *componentsHandler) Close() {
	if ch == nil {
		return
	}

	if ch.sqliteWrapper != nil {
		err := ch.sqliteWrapper.Close()
		log.LogIfError(err)
	}

	if ch.apiEngine != nil {
		err := ch.apiEngine.Close()
		log.LogIfError(err)
	}
}
