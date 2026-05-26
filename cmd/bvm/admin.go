package main

import (
    "fmt"
    "github.com/aziskebanaran/bvm-core/pkg/wallet"
    "github.com/spf13/cobra"
)

var adminCmd = &cobra.Command{
    Use:   "admin",
    Short: "Pusat komando administratif jaringan",
}

var setVaultsCmd = &cobra.Command{
    Use:   "set-vaults [addr1,addr2,addr3]",
    Short: "Mendaftarkan daftar alamat Vault ke sistem storage",
    Args:  cobra.ExactArgs(1),
    Run: func(cmd *cobra.Command, args []string) {
        h, _ := cmd.Flags().GetString("home")
        walletFile := fmt.Sprintf("%s/node_wallet.json", h)

        w, err := wallet.LoadWallet(walletFile)
        if err != nil {
            fmt.Println("❌ Wallet tidak ditemukan!")
            return
        }

        // 1. Tanda tangani pesan untuk keamanan
        msg := "UPDATE_VAULTS_" + args[0]
        sig, _ := w.SignMessage(msg)

        // 2. Kirim ke client (menggunakan bvmClient yang sudah ada di main.go)
        err = bvmClient.SetVaults(args[0], bvmClient.Token, sig)
        if err != nil {
            fmt.Printf("❌ Gagal mendaftarkan Vault: %v\n", err)
            return
        }
        fmt.Println("✅ Daftar Vault berhasil disinkronisasi ke seluruh jaringan!")
    },
}

func init() {
    adminCmd.AddCommand(setVaultsCmd)
    // Langsung daftarkan ke rootCmd yang ada di main.go
    rootCmd.AddCommand(adminCmd)
}
