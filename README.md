AWS tidy can be used to enforce policies on AWS resource tagging, etc

You can find an example config [here](config.yml)

Running:
```
go build
AWS_PROFILE=<your profile> ./reaper -c config.yml
```
