package api

import (
    "context"
    "net/http"
    "strings"
    "github.com/golang-jwt/jwt/v5"
    "github.com/aziskebanaran/bvm-core/x/storage/keeper" // Sesuaikan dengan path Sultan
)


// 🚩 TAMBAHKAN INI DI SINI
var JWT_SECRET = []byte("BVM-SULTAN-RAHASIA-2026")

func AuthenticateBVMCloud(sk *keeper.StorageKeeper) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { // 🚩 SUDAH DIPERBAIKI

            // 1. Cek API Key Aplikasi via StorageKeeper (sk)
            appID := r.Header.Get("X-BVM-App-ID")

            appData, err := sk.GetAppMetadata(appID) 
            if err != nil {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "App ID Tidak Terdaftar"}`))
                return
            }

            // 2. Cek JWT User
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, `{"error": "Authorization Header Diperlukan"}`, 401)
                return
            }
            tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

            token, _ := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
                return JWT_SECRET, nil
            })

            if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
                username, _ := claims["username"].(string)

                // 3. SINKRONISASI CONTEXT (Gunakan "app_metadata")
                ctx := r.Context()
                ctx = context.WithValue(ctx, "app_id", appID)
                ctx = context.WithValue(ctx, "user_id", username)
                ctx = context.WithValue(ctx, "app_metadata", appData) 

                // Teruskan ke handler
                next.ServeHTTP(w, r.WithContext(ctx))
            } else {
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusUnauthorized)
                w.Write([]byte(`{"error": "Token User Tidak Valid"}`))
            }
        }) // 🚩 TUTUP HANDLERFUNC
    } // 🚩 TUTUP MIDDLEWARE RETURN
}
