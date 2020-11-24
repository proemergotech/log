# Log package for our systems

Log packages for the different frameworks we use in our systems.

## Installation

1. Add dependency to go mod
2. Run go build/run/tidy

```bash
go get -u github.com/proemergotech/log v0.3.0
```

## Usage

### Zap

```go
    err := zap.RegisterEncoder(zaplog.EncoderType, zaplog.NewEncoder([]string{
      trace.CorrelationIDField,
      trace.WorkflowIDField,
      log.AppName,
      log.AppVersion,
    }))
    if err != nil {
    	panic(fmt.Sprintf("couldn't create logger: %v", err))
    }

    zapConf := zap.NewProductionConfig()
    zapConf.Encoding = zaplog.EncoderType

    logLvl := new(zapcore.Level)
    err = logLvl.Set("debug")
    if err != nil {
    	panic(err)
    }
    zapConf.Level = zap.NewAtomicLevelAt(*logLvl)

    zapLog, err := zapConf.Build()
    if err != nil {
    	panic(fmt.Sprintf("couldn't create logger: %v", err))
    }
    zapLog = zapLog.With(
    	zap.String(log.AppName, config.Name),
    	zap.String(log.AppVersion, config.Version),
    )

    logger := zaplog.NewLogger(zapLog, trace.Mapper())
    log.SetGlobalLogger(logger)
    
    log.Info(context.Background(), "hello world", "world", "earth")
```

Zap logger will log error fields if the error implements the `fielder` interface:

```go
    type fielder interface {
      Fields() []interface{}
    }
```

If the error wrapped other errors and implements the `causer` interface, the nested errors and their fields will be logged too.

```go
    type causer interface {
      Cause() error
    }
```

For a complete example see [errorfields](./_examples/errorfields/main.go).

## Development

- install go
- check out project to: $GOPATH/src/github.com/proemergotech/log
