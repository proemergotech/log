package zaplog

import (
	"go.uber.org/zap/buffer"
	"go.uber.org/zap/zapcore"

	"strings"
	"time"
)

const EncoderType = "dliver"

type Encoder struct {
	zapcore.Encoder
	pool        buffer.Pool
	specialKeys map[string]struct{}
}

// NewEncoder create a new zapcore.Encoder configured for the dliver system needs.
// During encoding field names matching a specialKeys entry will be added to the log message separately from the other fields.
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
	buf.AppendString(entry.Level.String())
	buf.AppendString(" ")
	if entry.LoggerName != "" {
		buf.AppendString(entry.LoggerName)
		buf.AppendString(" - ")
	}

	buf.AppendString(entry.Message)
	buf.AppendString(" ##<")

	for i := len(fields) - 1; i >= 0; i-- {
		field := fields[i]
		if _, ok := de.specialKeys[field.Key]; ok {
			buf.AppendString(field.Key + "=")
			buf.AppendString(strings.Replace(field.String, ";", "", -1))
			buf.AppendString(";")
			fields = append(fields[0:i], fields[i+1:]...)
		}
	}
	buf.AppendString(">##")

	fieldsBuf, err := de.Encoder.EncodeEntry(entry, fields)
	if err != nil {
		return nil, err
	}

	buf.Write(fieldsBuf.Bytes())

	return buf, nil
}
