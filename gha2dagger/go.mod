module github.com/aweris/gale/gha2dagger

go 1.21

require (
	github.com/aweris/gale/common v0.0.0
	github.com/dave/jennifer v1.7.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/kr/text v0.2.0 // indirect
	golang.org/x/text v0.13.0
)

replace github.com/aweris/gale/common => ../common
