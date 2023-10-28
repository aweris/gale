module actions-generator

go 1.21

require (
	github.com/99designs/gqlgen v0.17.39
	github.com/Khan/genqlient v0.6.0
	github.com/aweris/gale/common v0.0.0
	github.com/dave/jennifer v1.7.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/sync v0.4.0
	golang.org/x/text v0.13.0
	gopkg.in/yaml.v3 v3.0.1
)

require (
	github.com/google/uuid v1.3.1 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/sosodev/duration v1.2.0 // indirect
	github.com/vektah/gqlparser/v2 v2.5.10 // indirect
)

replace github.com/aweris/gale/common => ../../../common
