module github.com/braswelljr/alGOrithm/web/runtime

go 1.27

require (
	github.com/braswelljr/alGOrithm v0.0.0-00010101000000-000000000000
	github.com/braswelljr/gnux v1.5.1
	github.com/klauspost/compress v1.20.3
)

require (
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/mattn/go-isatty v0.0.24 // indirect
	github.com/spf13/cobra v1.10.2 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
	golang.org/x/term v0.47.0 // indirect
)

// The library is this repository's own, read from the checkout.
replace github.com/braswelljr/alGOrithm => ../..
