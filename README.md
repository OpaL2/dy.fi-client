# dy.fi client

Golang client for dy.fi dynamic dns -service

Client is runned inside docker and credentials are provided as environment variables.

Following compose file can be used as base:

```yaml
version: '3'

services:
  client:
    build: github.com/OpaL2/dy.fi-client
    environment:
      - HOSTNAME=
      - USERNAME=
      - PASSWORD
    restart: always
```