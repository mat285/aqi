package slack

const (
	// ParamTextKey is the slack text key
	ParamTextKey = "text"
	// ParamUserIDKey is the slack user key
	ParamUserIDKey = "user_id"

	// TimestampHeaderParam is the header for timestamp
	TimestampHeaderParam = "X-Slack-Request-Timestamp"
	// SignatureHeaderParam is the header for signature
	SignatureHeaderParam = "X-Slack-Signature"
)

const (
	// ResponseTypeInChannel is the in channel response type
	ResponseTypeInChannel = "in_channel"
	// ResponseTypeEphemeral is the ephemeral response type
	ResponseTypeEphemeral = "ephemeral"
)

const (
	// EnvVarSignatureSecret is the secret for the signature
	EnvVarSignatureSecret = "SLACK_SIGNATURE_SECRET"
)

const (
	// ErrInvalidDigest indicates the digest for a request is improperly formatted
	ErrInvalidDigest = "ErrInvalidDigest"
	// ErrSignatureInvalid indicates the signature of a request is invalid
	ErrSignatureInvalid = "ErrSignatureInvalid"
)
