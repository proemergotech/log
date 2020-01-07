# Release Notes

## v1.0.0 / 2020-01-07
- release version v1.0.0
- add verify script and ci job

## v0.3.0 / 2019-12-12
- use go modules
- update to echo v4
- remove gin

## v0.2.2 / 2019-03-26
- added http log for adding logging capability to go sdk http client
- modified elasticlog package to only contain a logger for elastic client errors

## v0.2.1 / 2019-01-22
- use "error" string for zap encoder output for all levels above error

## 0.2.0 / 2018-09-11
- add GlobalLogger function
- remove echo ErrorMiddleware, because in our services the http status code is set after this middleware running 

## 0.1.8 / 2018-07-30
- add development zap encoder, it adds stacktrace from the error object, indent the fields, adds color output(just for the levels by default)

## 0.1.7 / 2018-07-19
- fixed echo debug middleware in case trace middleware runs after log middleware

## 0.1.6 / 2018-06-21
- added Echo web framework logger

## 0.1.5 / 2018-05-25
- removed version contraint from elastic library

## 0.1.4 / 2018-05-23
- added elastic logger

## 0.1.3 / 2018-04-19
- added IsDebug() to global logger

## 0.1.2 / 2018-04-03
- removed version contraint from geb-client dependency

## 0.1.1 / 2018-03-12
- added special character escape for log message and special log fields for zap encoder

## 0.1.0 / 2018-03-12
- project created
