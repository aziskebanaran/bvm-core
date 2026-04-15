package main

import (
    "bvm.core/pkg/sdk"
)

//export handle
func handle() {
    method := sdk.GetMethod() // Ambil metode dari payload (misal: "transfer")
    
    switch method {
    case "transfer":
        to := sdk.GetArgString("to")
        amount := sdk.GetArgUint64("amount")
        // Logika internal atau panggil host bank
        sdk.Transfer(to, amount)
        
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
