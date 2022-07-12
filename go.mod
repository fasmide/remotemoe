module github.com/fasmide/remotemoe

go 1.17

require (
	github.com/fasmide/hostkeys v0.0.0-20211023164018-0a66d786b24e
	github.com/fatih/color v1.13.0
	github.com/spf13/cobra v1.5.0
	github.com/spf13/pflag v1.0.5
	golang.org/x/crypto v0.0.0-20220622213112-05595931fe9d
	golang.org/x/sync v0.0.0-20220601150217-0de741cfad7f
	golang.org/x/term v0.0.0-20220526004731-065cf7ba2467
)

require (
	github.com/inconshreveable/mousetrap v1.0.0 // indirect
	github.com/mattn/go-colorable v0.1.12 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	golang.org/x/net v0.0.0-20220708220712-1185a9018129 // indirect
	golang.org/x/sys v0.0.0-20220712014510-0a85c31ab51e // indirect
	golang.org/x/text v0.3.7 // indirect
)

// Remove this once once https://github.com/golang/crypto/pull/211 or similar gets merged
// this is to support openssh clients version 8.8 and above using RSA keys with rsa-sha2-512 signatures
replace golang.org/x/crypto => github.com/drakkan/crypto v0.0.0-20220615080207-8cff98973996
