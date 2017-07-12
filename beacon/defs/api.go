package defs

const (
	// APIRegistrationEndpoint is used to open the websocket connection with the beacon api.
	APIRegistrationEndpoint = "register"

	// APIAuthorizationHeader is used during the registration process.
	APIAuthorizationHeader = "x-device-auth"

	// APIReportMessageLabel is the label used when signing digests to the api during report feedback.
	APIReportMessageLabel = "report"
)
