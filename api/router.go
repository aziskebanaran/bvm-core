package api

import (
	"github.com/aziskebanaran/BVM.core/pkg/storage"
	"github.com/aziskebanaran/BVM.core/x" // 🚩 WAJIB: Gunakan interface pusat
	"github.com/aziskebanaran/BVM.core/x/storage/keeper"
	"time"
	"encoding/json"
	"net/http"
)

// NewRouter: Sekarang hanya butuh 4 argumen (BC dan MP sudah ada di dalam K)
func NewRouter(k x.BVMKeeper, mp x.MempoolKeeper, store storage.BVMStore, nodeAddr string) http.Handler {
	mux := http.NewServeMux()

	// --- 1. JALUR MINING ---
	mux.HandleFunc("/api/mine", HandleMine(k))
        mux.HandleFunc("/api/getwork", HandleGetWork(k))

	// --- 2. JALUR EKONOMI ---
	// 🚩 PERHATIKAN: Kita tidak lagi mengoper 'bc' karena 'k' sudah punya akses ke sana
	mux.HandleFunc("/api/send", HandleSend(k)) 
	mux.HandleFunc("/api/balance", HandleBalance(k))
	mux.HandleFunc("/api/account", HandleGetAccount(k))
	mux.HandleFunc("/api/mempool", HandleMempool(k))
	mux.HandleFunc("/api/mempool/stats", HandleMempoolStats(k))

	mux.HandleFunc("/api/nonce", HandleNonce(k))

	// --- 3. INFORMASI JARINGAN ---
	mux.HandleFunc("/api/params", HandleParams(k))
	mux.HandleFunc("/api/stats", HandleStats(k)) // Cukup kirim 'k'
	mux.HandleFunc("/api/status", HandleStatus(k))
	mux.HandleFunc("/api/validators", HandleValidators(k))

	// --- 4. EXPLORER & UTILITY ---
	mux.HandleFunc("/api/history", HandleAddressHistory(store))
	mux.HandleFunc("/api/search", HandleSearchUser(k))
	mux.HandleFunc("/api/inspect", HandleInspect(nodeAddr))
        mux.HandleFunc("/api/events", HandleGetEvents)
        mux.HandleFunc("/api/upgrades", HandleGetUpgrades(k))

        // Untuk block, kita bisa ambil dari Keeper
	mux.HandleFunc("/api/explorer/stream", HandleRealTimeExplorer(k))

	mux.HandleFunc("/api/block/", HandleGetBlockByHeight(k))

	// --- 6. TOKEN FACTORY (The Forge) ---
	mux.HandleFunc("/api/token/create", HandleCreateToken(k))

	mux.HandleFunc("/api/staking/join", HandleJoinValidator(k))
	mux.HandleFunc("/api/staking/unstake", HandleUnstake(k))

	// --- 4. EXPLORER & UTILITY ---
	mux.HandleFunc("/api/peers", HandlePeers(k))

        // --- 7. SISTEM AUDIT & TRANSPARANSI ---
        // Loket untuk mengecek supply koin dan pembakaran fee
        mux.HandleFunc("/api/audit/supply", HandleAuditSupply(k))
        // Loket untuk verifikasi kode sumber Smart Contract (WASM)
        mux.HandleFunc("/api/audit/verify", HandleVerifyContract)

        // --- 8. BVM CLOUD STORAGE ---
        storageK, _ := k.GetCloudStorage().(*keeper.StorageKeeper)

        // Pastikan nama variabel di sini...
        authMiddleware := AuthenticateBVMCloud(storageK)

        // ...sama dengan yang digunakan di sini
        mux.HandleFunc("/api/app/register", HandleAppRegister(storageK, k))
        mux.Handle("/api/storage/put", authMiddleware(http.HandlerFunc(HandleAppPut(storageK, k))))
        mux.Handle("/api/storage/get", authMiddleware(http.HandlerFunc(HandleAppGet(storageK))))

        // --- 9. IDENTITY & AUTH SYSTEM ---
        mux.HandleFunc("/api/login", HandleLogin(k)) // Pintu masuk utama untuk JWT

	// --- 5. INFO NODE (Root API) ---
	mux.HandleFunc("/api/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/" || r.URL.Path == "/api" {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(map[string]interface{}{
				"node_identity": nodeAddr,
				"version":       "v1.5-Keeper-SDK",
				"network":       "BVM-Mainnet-Beta",
				"status":        "Operational",
				"binary_codec":  "Msgpack/v5", // Informasi tambahan untuk klien
				"time":          time.Now().Format(time.RFC1123),
			})
			return
		}
		http.NotFound(w, r)
	})

	return enableCORS(mux)
}

// enableCORS: Izin akses untuk Web Dashboard / Explorer Sultan
func enableCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
		if r.Method == "OPTIONS" {
			return
		}
		next.ServeHTTP(w, r)
	})
}
