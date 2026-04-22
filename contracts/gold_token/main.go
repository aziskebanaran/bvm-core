package main

import (
    "github.com/aziskebanaran/bvm-core/pkg/sdk"
)

//export handle
func handle() {
    method := sdk.GetMethod() // Ambil metode dari payload (misal: "transfer")
    
    switch method {
	// Di dalam switch case "transfer"
case "transfer":
    to := sdk.GetArgString("to")
    amount := sdk.GetArgUint64("amount")
    sender := sdk.GetCaller() // Ambil siapa yang memanggil contract

    // 🚩 PERBAIKAN: Kirim 4 argumen sesuai SDK
    sdk.Transfer(sender, to, amount, "GOLD")



    case "mint":
        // Cek Security: Hanya Owner yang bisa Mint
        if sdk.GetCaller() == "bvmfa15608e2b0225f96b915" {
             target := sdk.GetArgString("target")
             amt := sdk.GetArgUint64("amount")
             sdk.Mint(target, amt, "GOLD")
        }
    }
}

func main() {}
