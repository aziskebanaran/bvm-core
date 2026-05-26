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

func StartFullNode(k x.BVMKeeper, mp x.MempoolKeeper, p2pMgr x.P2PKeeper, store storage.BVMStore, nodeAddr string, apiPort int, p2pPort int) {
    // 🚩 Jalur 1: P2P Engine
    go func() {
        logger.Success("P2P", fmt.Sprintf("📡 P2P Engine Aktif di port %d", p2pPort))
        p2p.StartNode(p2pPort)
    }()

    // 🚩 Jalur 2: API Server
    go func() {
        port := fmt.Sprintf(":%d", apiPort)
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

	select {}
}
