module bvm.core/mining

go 1.25.0 // Gunakan versi yang terinstall di Termux/Server Sultan

require bvm.core v0.0.0

require (
	github.com/cbergoon/merkletree v0.2.0 // indirect
	github.com/tyler-smith/go-bip39 v1.1.0 // indirect
	golang.org/x/crypto v0.0.0-20220518034528-6f7dac969898 // indirect
)

replace bvm.core => ../
