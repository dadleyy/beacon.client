package defs

const (
	// APIRegistrationEndpoint is used to open the websocket connection with the beacon api.
	APIRegistrationEndpoint = "register"

	// APIAuthorizationHeader is used during the registration process.
	APIAuthorizationHeader = "x-device-auth"

	// APIFeedbackEndpoint is the value sent in the content-type header when sending feedback data to the api.
	APIFeedbackEndpoint = "/device-feedback"

	// APIFeedbackContentTypeHeader is the value sent in the content-type header when sending feedback data to the api.
	APIFeedbackContentTypeHeader = "application/octet-stream"

	// APIReportMessageLabel is the label used when signing digests to the api during report feedback.
	APIReportMessageLabel = "report"
)
