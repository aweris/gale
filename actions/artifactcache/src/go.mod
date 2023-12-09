module github.com/aweris/gale/services/artifactcache

go 1.21.3

require (
	github.com/aweris/gale/common v0.0.0-00010101000000-000000000000
	github.com/caarlos0/env/v9 v9.0.0
	github.com/julienschmidt/httprouter v1.3.0
	go.etcd.io/bbolt v1.3.7
)

require (
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/sys v0.13.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/aweris/gale/common => ../../../common
