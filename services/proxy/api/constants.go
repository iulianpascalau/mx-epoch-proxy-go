package api

// Constants for endpoint namings
const (
	EndpointApiAccessKeys         = "/api/admin-access-keys"
	EndpointApiAdminUsers         = "/api/admin-users"
	EndpointApiLogin              = "/api/login"
	EndpointApiRegister           = "/api/register"
	EndpointApiActivate           = "/api/activate"
	EndpointCaptchaMultiple       = "/api/captcha/"
	EndpointCaptchaSingle         = "/api/captcha"
	EndpointAppInfo               = "/api/app-info"
	EndpointApiPerformance        = "/api/performance"
	EndpointApiChangePassword     = "/api/change-password"
	EndpointApiRequestEmailChange = "/api/request-email-change"
	EndpointApiConfirmEmailChange = "/api/confirm-email-change"

	EndpointApiCryptoPaymentConfig        = "/api/crypto-payment/config"
	EndpointApiCryptoPaymentCreateAddress = "/api/crypto-payment/create-address"
	EndpointApiCryptoPaymentAccount       = "/api/crypto-payment/account"
	EndpointApiAdminCryptoPaymentAccount  = "/api/admin-crypto-payment/account"

	EndpointSwagger = "/swagger/"
	EndpointRoot    = "/"

	EndpointFrontendLogin = "/#/login"
)
