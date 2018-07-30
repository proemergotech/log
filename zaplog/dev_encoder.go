package zaplog

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/pkg/errors"
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"
)

const DevEncoderType = "dliver-dev"

type DevEncoderOption func(de *DevEncoderOptions)

type DevEncoderOptions struct {
	timeLayout    string
	indentFields  bool
	excludeFilter map[string]struct{}
}

type devEncoder struct {
	*zapcore.MapObjectEncoder
	options DevEncoderOptions
	pool    buffer.Pool
}

var (
	levelToColor = map[zapcore.Level]Color{
		zapcore.DebugLevel:  Magenta,
		zapcore.InfoLevel:   Blue,
		zapcore.WarnLevel:   Yellow,
		zapcore.ErrorLevel:  Red,
		zapcore.DPanicLevel: Red,
		zapcore.PanicLevel:  Red,
		zapcore.FatalLevel:  Red,
	}
	unknownLevelColor = Red

	levelToCapitalColorString = make(map[zapcore.Level]string, len(levelToColor))
)

func init() {
	for level, color := range levelToColor {
		levelToCapitalColorString[level] = color.Add(level.CapitalString())
	}
}

// NewDevelopmentEncoder create a new zapcore.Encoder configured for development.
func NewDevelopmentEncoder(options ...DevEncoderOption) func(zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		do := DevEncoderOptions{
			timeLayout:    "15:04:05.999999",
			indentFields:  true,
			excludeFilter: make(map[string]struct{}),
		}

		for _, option := range options {
			option(&do)
		}

		return &devEncoder{
			MapObjectEncoder: zapcore.NewMapObjectEncoder(),
			options:          do,
			pool:             buffer.NewPool(),
		}, nil
	}
}

func (de *devEncoder) Clone() zapcore.Encoder {
	// when zap creates a new encoder, for example to add new base fields, it will call this method
	// we need to copy the previously added base fields
	enc := zapcore.NewMapObjectEncoder()
	for k, v := range de.MapObjectEncoder.Fields {
		enc.Fields[k] = v
	}

	return &devEncoder{
		MapObjectEncoder: enc,
		options:          de.options,
		pool:             buffer.NewPool(),
	}
}

func (de *devEncoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := de.pool.Get()

	for filter := range de.options.excludeFilter {
		if strings.Contains(entry.Message, filter) {
			return buf, nil
		}
	}

	// make it thread safe, the 'de' instance MapObjectEncoder contains only the base fields, we need to include them in all log entry
	// do not modify the de.MapObjectEncoder directly
	enc := zapcore.NewMapObjectEncoder()
	for k, v := range de.MapObjectEncoder.Fields {
		enc.Fields[k] = v
	}

	buf.AppendString(entry.Time.Format(de.options.timeLayout))
	buf.AppendString(" ")

	level, ok := levelToCapitalColorString[entry.Level]
	if !ok {
		level = unknownLevelColor.Add(entry.Level.CapitalString())
	}
	buf.AppendString(level)

	buf.AppendString(" ")
	if entry.LoggerName != "" {
		buf.AppendString(entry.LoggerName)
		buf.AppendString(" - ")
	}

	buf.AppendString(entry.Message)
	buf.AppendString("\n")

	errWithStack := ""

	for _, f := range fields {
		err, ok := f.Interface.(error)
		if ok {
			type stackTracer interface {
				StackTrace() errors.StackTrace
			}

			type causer interface {
				Cause() error
			}

			for err != nil {
				_, ok := err.(stackTracer)
				if ok {
					errWithStack = fmt.Sprintf("%+v\n", err)
					// errors package will format the nested errors too with their stack traces
					// so don't need to check the underlying errors
					break
				}

				cause, ok := err.(causer)
				if !ok {
					break
				}
				err = cause.Cause()
			}

			// if it wasn't a stackTracer, just add the error message to the fields
			if errWithStack == "" {
				f.AddTo(enc)
			}
		} else {
			f.AddTo(enc)
		}
	}

	var b []byte
	var err error
	if de.options.indentFields {
		b, err = json.MarshalIndent(enc.Fields, "", "  ")
	} else {
		b, err = json.Marshal(enc.Fields)
	}

	if err != nil {
		panic(err)
	}

	buf.Write(b)
	buf.AppendString("\n")
	buf.AppendString(errWithStack)

	return buf, nil
}

func TimeLayout(layout string) DevEncoderOption {
	return func(de *DevEncoderOptions) {
		de.timeLayout = layout
	}
}

func IndentFields(indent bool) DevEncoderOption {
	return func(de *DevEncoderOptions) {
		de.indentFields = indent
	}
}

func ExcludeFilter(filters ...string) DevEncoderOption {
	return func(de *DevEncoderOptions) {
		for _, f := range filters {
			de.excludeFilter[f] = struct{}{}
		}
	}
}
