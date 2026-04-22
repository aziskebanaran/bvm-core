package node

import (
	"fmt"
	"net/http"
	"time"
	"github.com/aziskebanaran/bvm-core/api"
	"github.com/aziskebanaran/bvm-core/pkg/logger"
	"github.com/aziskebanaran/bvm-core/pkg/p2p"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
	"github.com/aziskebanaran/bvm-core/x"
)

func StartFullNode(k x.BVMKeeper, mp x.MempoolKeeper, p2pMgr x.P2PKeeper, store storage.BVMStore, nodeAddr string) {
	// 🚩 Jalur 1: P2P Engine (Port 9090) - Mesin Antar Node
	go func() {
		p2pPort := 9090
		logger.Success("P2P", fmt.Sprintf("📡 P2P Engine Aktif di port %d", p2pPort))
		p2p.StartNode(p2pPort)
	}()

	// 🚩 Jalur 2: API Server (Port 8080) - Jalur Sultan & Aplikasi
	go func() {
		port := ":8080"
		router := api.NewRouter(k, mp, store, nodeAddr)
		server := &http.Server{
			Addr:         port,
			Handler:      router,
			ReadTimeout:  30 * time.Second,
			WriteTimeout: 30 * time.Second,
		}
		logger.Success("NETWORK", "🌐 API Server aktif di http://localhost"+port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("SERVER", "Gagal menjalankan server: "+err.Error())
		}
	}()

	select {} // Hold agar proses tidak mati
}
