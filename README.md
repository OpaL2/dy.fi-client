# dy.fi client

Golang client for dy.fi dynamic dns -service

Client takes in configuration file as first argument and template config.yaml is included in repository. Log is printed to stdout.

Client checks hourly if update dns record update is required and schedules update if required. Update is executed in 15 minutes from scheduling. 