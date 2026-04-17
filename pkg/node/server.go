package node

import (
	"github.com/aziskebanaran/bvm-core/api"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x" // 🚩 Impor Interface Pusat
	"net/http"
	"time"
)

// StartFullNode: Sekarang menggunakan Interface (Loose Coupling)
func StartFullNode(
	k x.BVMKeeper,       // 🚩 Gunakan Interface, bukan *keeper.Keeper
	mp x.MempoolKeeper,  // 🚩 Gunakan Interface, bukan *mempool.Mempool
	p2p x.P2PKeeper,     // 🚩 Tambahkan P2P jika diperlukan di API
	store storage.BVMStore, 
	nodeAddr string,
) {
	port := ":8080"

	// 🚩 Router juga harus disesuaikan agar menerima Interface
	router := api.NewRouter(k, mp, store, nodeAddr)

	server := &http.Server{
		Addr:         port,
		Handler:      router,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	logger.Success("NETWORK", "🌐 API Server aktif di http://localhost"+port)

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		logger.Error("SERVER", "Gagal menjalankan server: "+err.Error())
	}
}
