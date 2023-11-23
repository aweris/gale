module gale

go 1.21.3

require (
	github.com/99designs/gqlgen v0.17.31
	github.com/Khan/genqlient v0.6.0
	github.com/aweris/gale/common v0.0.0-00010101000000-000000000000
	github.com/google/uuid v1.4.0
	golang.org/x/exp v0.0.0-20231006140011-7918f672742d
	golang.org/x/sync v0.4.0
)

require github.com/kr/text v0.2.0 // indirect

require (
	github.com/vektah/gqlparser/v2 v2.5.6 // indirect
	gopkg.in/yaml.v3 v3.0.1
)

replace github.com/aweris/gale/common => ../../common
