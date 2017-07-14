package defs

import "log"

const (
	// DebugLogLevelTag is used for debugf logger calls
	DebugLogLevelTag = "debug"

	// InfoLogLevelTag is used for infof logger calls
	InfoLogLevelTag = "info"

	// WarnLogLevelTag is used for errorf logger calls
	WarnLogLevelTag = "warn"

	// ErrorLogLevelTag is used for errorf logger calls
	ErrorLogLevelTag = "error"

	// RuntimeLoggerPrefix is the logger prefix used in the main runtime loop
	RuntimeLoggerPrefix = "[client runtime] "

	// DebugStateLoggerPrefix is the logger prefix used by the state logger
	DebugStateLoggerPrefix = "[state logger] "

	// FeedbackProcessorLoggerPrefix is used by the command processor
	FeedbackProcessorLoggerPrefix = "[feedback processor] "

	// HeartbeatProcessorLoggerPrefix is used by the command processor
	HeartbeatProcessorLoggerPrefix = "[heartbeat processor] "

	// CommandProcessorLoggerPrefix is used by the command processor
	CommandProcessorLoggerPrefix = "[command processor] "

	// DefaultLogFlags is a shared bitmask for default log.Logger flags
	DefaultLogFlags = log.Ldate | log.Ltime
)
