package zaplog

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"github.com/proemergotech/log/v3"

	"strings"
	"time"
)

const (
	EncoderType          = "dliver"
	messageTruncateLimit = 500
)

type Encoder struct {
	zapcore.Encoder
	pool        buffer.Pool
	specialKeys map[string]struct{}
}

var msgReplacer = strings.NewReplacer("\n", "\\n", "\r", "\\r")
var keyReplacer = strings.NewReplacer("\n", "\\n", "\r", "\\r", "=", "", ";", "")
var valueReplacer = keyReplacer

// NewEncoder create a new zapcore.Encoder configured for the dliver system needs.
// During encoding field names matching a specialKeys entry will be added to the log message separately from the other fields.
// Special field keys and values may not include ';' and '='. These characters will be replaced with empty string.
func NewEncoder(specialKeys []string) func(zapcore.EncoderConfig) (zapcore.Encoder, error) {
	return func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
		// copy the struct
		jsonCfg := cfg
		jsonCfg.MessageKey = ""
		jsonCfg.LevelKey = ""
		jsonCfg.TimeKey = ""
		jsonCfg.NameKey = ""
		jsonCfg.CallerKey = ""
		jsonCfg.StacktraceKey = ""
		jsonCfg.LineEnding = ""

		k := make(map[string]struct{}, len(specialKeys))
		for _, v := range specialKeys {
			k[v] = struct{}{}
		}

		return &Encoder{
			Encoder:     zapcore.NewJSONEncoder(jsonCfg),
			pool:        buffer.NewPool(),
			specialKeys: k,
		}, nil
	}
}

func (de *Encoder) Clone() zapcore.Encoder {
	return &Encoder{
		Encoder:     de.Encoder.Clone(),
		pool:        buffer.NewPool(),
		specialKeys: de.specialKeys,
	}
}

func (de *Encoder) EncodeEntry(entry zapcore.Entry, fields []zapcore.Field) (*buffer.Buffer, error) {
	buf := de.pool.Get()

	buf.AppendString(entry.Time.Format(time.RFC3339Nano))
	buf.AppendString(" ")
	buf.AppendString(levelToString(entry.Level))
	buf.AppendString(" ")
	if entry.LoggerName != "" {
		buf.AppendString(entry.LoggerName)
		buf.AppendString(" - ")
	}

	buf.AppendString(log.Truncate(msgReplacer.Replace(entry.Message), messageTruncateLimit, "..."))
	buf.AppendString(" ##<")

	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		if _, ok := de.specialKeys[field.Key]; ok {
			buf.AppendString(keyReplacer.Replace(field.Key) + "=")
			buf.AppendString(valueReplacer.Replace(field.String))
			buf.AppendString(";")
			fields = append(fields[0:i], fields[i+1:]...)
		}
	}
	buf.AppendString(">##")

	fieldsBuf, err := de.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	_, _ = buf.Write(fieldsBuf.Bytes())

	return buf, nil
}

func levelToString(lvl zapcore.Level) string {
	switch lvl {
	case zapcore.DebugLevel:
		return "debug"
	case zapcore.InfoLevel:
		return "info"
	case zapcore.WarnLevel:
		return "warn"
	default:
		return "error"
	}
}
