# Log package for the dliver system

Log packages for the different frameworks we use in dliver.

## Installation

1. Add dependency to dep
2. Run dep ensure

#### Gopkg.toml

```toml
[[constraint]]
  name = "gitlab.com/proemergotech/log-go"
  source = "git@gitlab.com:proemergotech/log-go.git"
  version = "0.1.0"
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

## Documentation

Private repos don't show up on godoc.org so you have to run it locally.

```
godoc -http=":6060"
```

Then open http://localhost:6060/pkg/gitlab.com/proemergotech/log-go/

## Development

- install go
- check out project to: $GOPATH/src/gitlab.com/proemergotech/log-go
- install dep
- run dep ensure
