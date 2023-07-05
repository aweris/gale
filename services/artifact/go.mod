module github.com/aweris/gale/services/artifact

go 1.20

require (
	github.com/aweris/gale v0.0.0 // it's a local dependency version is not important
	github.com/julienschmidt/httprouter v1.3.0
	github.com/spf13/pflag v1.0.5
)

replace github.com/aweris/gale => ../..
