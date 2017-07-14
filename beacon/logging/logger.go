package logging

import "os"
import "log"
import "fmt"
import "github.com/ttacon/chalk"
import "github.com/dadleyy/beacon.client/beacon/defs"

const (
	// Black = chalk.Black
	Black = iota
	// Red = chalk.Red
	Red
	// Green = chalk.Green
	Green
	// Yellow = chalk.Yellow
	Yellow
	// Blue = chalk.Blue
	Blue
	// Magenta = chalk.Magenta
	Magenta
	// Cyan = chalk.Cyan
	Cyan
	// White = chalk.White
	White
)

// New retrurns a new logger
func New(name string, colorFlag uint) Logger {
	prefix := color(colorFlag, name)
	writer := log.New(os.Stdout, prefix, defs.DefaultLogFlags)
	return &ColorLogger{writer}
}

// Logger defines an interface for leveled logging
type Logger interface {
	Errorf(string, ...interface{})
	Warnf(string, ...interface{})
	Debugf(string, ...interface{})
	Infof(string, ...interface{})
}

// ColorLogger wraps the golang log.Logger struct for coloring
type ColorLogger struct {
	*log.Logger
}

// Errorf sends the output colored
func (logger *ColorLogger) Errorf(format string, items ...interface{}) {
	logger.printfc(chalk.Red, defs.ErrorLogLevelTag, format, items...)
}

// Warnf sends the output colored
func (logger *ColorLogger) Warnf(format string, items ...interface{}) {
	logger.printfc(chalk.Yellow, defs.WarnLogLevelTag, format, items...)
}

// Infof sends the output colored
func (logger *ColorLogger) Infof(format string, items ...interface{}) {
	logger.printfc(chalk.Cyan, defs.InfoLogLevelTag, format, items...)
}

// Debugf sends the output colored
func (logger *ColorLogger) Debugf(format string, items ...interface{}) {
	logger.printfc(chalk.Blue, defs.DebugLogLevelTag, format, items...)
}

func (logger *ColorLogger) printfc(crayon chalk.Color, label string, format string, items ...interface{}) {
	labelTag := fmt.Sprintf("[%s]", label)
	formatted := fmt.Sprintf("%v %s", crayon.Color(labelTag), fmt.Sprintf(format, items...))
	logger.Printf("%s", formatted)
}

func color(colorFlag uint, text string) string {
	crayon := chalk.ResetColor

	switch colorFlag {
	case Black:
		crayon = chalk.Black
	case Red:
		crayon = chalk.Red
	case Green:
		crayon = chalk.Green
	case Yellow:
		crayon = chalk.Yellow
	case Blue:
		crayon = chalk.Blue
	case Magenta:
		crayon = chalk.Magenta
	case Cyan:
		crayon = chalk.Cyan
	case White:
		crayon = chalk.White
	}

	return crayon.Color(text)
}
