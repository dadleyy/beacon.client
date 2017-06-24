package defs

import "log"

const (
	// RuntimeLoggerPrefix is the logger prefix used in the main runtime loop
	RuntimeLoggerPrefix = "[client runtime] "

	// DebugStateLoggerPrefix is the logger prefix used by the state logger
	DebugStateLoggerPrefix = "[state logger] "

	// HeartbeatProcessorLoggerPrefix is used by the command processor
	HeartbeatProcessorLoggerPrefix = "[heartbeat processor] "

	// CommandProcessorLoggerPrefix is used by the command processor
	CommandProcessorLoggerPrefix = "[command processor] "

	// DefaultLogFlags is a shared bitmask for default log.Logger flags
	DefaultLogFlags = log.Ldate | log.Ltime
)
