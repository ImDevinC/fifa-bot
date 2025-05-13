module github.com/imdevinc/fifa-bot

go 1.23.0

toolchain go1.24.3

require (
	github.com/getsentry/sentry-go v0.15.0
	github.com/imdevinc/go-fifa v0.1.2
	github.com/redis/go-redis/v9 v9.8.0
	github.com/sirupsen/logrus v1.9.0
	github.com/stretchr/testify v1.8.1
	golang.org/x/sync v0.14.0
)

require (
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/google/go-querystring v1.1.0 // indirect
	github.com/kelseyhightower/envconfig v1.4.0 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/sys v0.2.0 // indirect
	golang.org/x/text v0.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/imdevinc/go-fifa => ../go-fifa
