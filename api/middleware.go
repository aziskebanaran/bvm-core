package api

import (
    "context"
    "net/http"
    "github.com/aziskebanaran/bvm-core/x"
    "github.com/aziskebanaran/bvm-core/x/storage/keeper"
)

// 🚀 Jenderal, JWT_SECRET SUDAH DIBUANG. 
// Sekarang kita menggunakan Kekuatan Kriptografi murni.

func AuthenticateBVMCloud(sk *keeper.StorageKeeper, k x.BVMKeeper) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

            // 1. Ambil Identitas & Bukti (Signature) dari Header
            appID := r.Header.Get("X-BVM-App-ID")
            address := r.Header.Get("X-BVM-Address")   // Alamat dompet bvmf...
            signature := r.Header.Get("X-BVM-Signature") // Tanda tangan digital
            message := r.Header.Get("X-BVM-Message")     // Pesan yang ditandatangani

            // 2. Cek apakah App ID terdaftar di StorageKeeper
            appData, err := sk.GetAppMetadata(appID)
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "App ID Tidak Terdaftar"}`))
                return
            }

            // 3. VERIFIKASI KRIPTOGRAFI (Pengganti JWT)
            // Mesin akan mengecek apakah Signature benar-benar dibuat oleh Address tersebut
            if address == "" || signature == "" || message == "" {
                http.Error(w, `{"error": "Bukti identitas (Address/Signature/Msg) diperlukan"}`, 401)
                return
            }

            isValid := k.GetAuth().VerifyManualSignature(address, message, signature)
            if !isValid {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "Tanda Tangan Digital Tidak Sah!"}`))
                return
            }

            // 4. SINKRONISASI CONTEXT
            // Sekarang kita menggunakan 'address' sebagai identitas utama, bukan lagi username JWT
            ctx := r.Context()
            ctx = context.WithValue(ctx, "app_id", appID)
            ctx = context.WithValue(ctx, "user_address", address)
            ctx = context.WithValue(ctx, "app_metadata", appData)

            // Teruskan ke handler dengan identitas yang sudah terverifikasi
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
