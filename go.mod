module github.com/aziskebanaran/bvm-core

go 1.26.2

require (
	github.com/aziskebanaran/bvm-lib v0.0.0-00010101000000-000000000000
	github.com/cbergoon/merkletree v0.2.0
	github.com/golang-jwt/jwt/v5 v5.3.1
	github.com/joho/godotenv v1.5.1
	github.com/spf13/cobra v1.10.2
	github.com/syndtr/goleveldb v1.0.1-0.20210819022825-2ae1ddf74ef7
	github.com/tetratelabs/wazero v1.11.0
	github.com/tyler-smith/go-bip39 v1.1.0
	github.com/vmihailenco/msgpack/v5 v5.4.1
)

require (
	github.com/golang/snappy v0.0.4 // indirect
	github.com/inconshreveable/mousetrap v1.1.0 // indirect
	github.com/spf13/pflag v1.0.10 // indirect
	github.com/stretchr/testify v1.11.1 // indirect
	github.com/vmihailenco/tagparser/v2 v2.0.0 // indirect
	golang.org/x/crypto v0.50.0 // indirect
	golang.org/x/net v0.53.0 // indirect
	golang.org/x/sys v0.43.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)

replace github.com/aziskebanaran/bvm-lib => ../bvm-lib

replace github.com/google/generative-ai-go => github.com/google/generative-ai-go v0.19.0
