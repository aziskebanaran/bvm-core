package api

import (
    "context"
    "net/http"
    "bvm.core/x/storage/keeper" // Sesuaikan dengan path Sultan
)

func AuthenticateBVMCloud(k *keeper.StorageKeeper) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            appID := r.Header.Get("X-BVM-App-ID")
            apiKey := r.Header.Get("X-BVM-API-Key")

            if appID == "" || apiKey == "" {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "🚫 Kredensial Tidak Lengkap"}`))
                return
            }

            app, err := k.GetAppMetadata(appID)
            if err != nil || app.APIKey != apiKey {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusForbidden)
                w.Write([]byte(`{"error": "🚫 API Key atau App ID Salah"}`))
                return
            }

            // Simpan metadata ke context agar Handler (Put/Get) bisa tahu siapa yang memanggil
            ctx := context.WithValue(r.Context(), "app_metadata", app)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
