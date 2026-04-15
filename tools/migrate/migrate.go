package migrate 

import (
	"bvm.core/pkg/storage"
	"bvm.core/x/bvm/types"
	"encoding/json"
	"fmt"
	"os" // GUNAKAN OS SEBAGAI PENGGANTI IOUTIL
	"path/filepath"
	"sort"
)

// RunMigration sekarang bisa di-import oleh cmd/bvm
func RunMigration(store storage.BVMStore) {
	fmt.Println("🔄 [MIGRATION ENGINE] JSON -> LevelDB...")

	blockDir := "data/blocks"
	files, err := os.ReadDir(blockDir)
	if err != nil {
		fmt.Println("ℹ️ Folder lama 'data/blocks' tidak ditemukan. Migrasi dilewati.")
		return
	}

	var allBlocks []types.Block

	for _, f := range files {
		if filepath.Ext(f.Name()) == ".json" {
			content, err := os.ReadFile(filepath.Join(blockDir, f.Name()))
			if err != nil {
				continue
			}
			var b types.Block
			if err := json.Unmarshal(content, &b); err == nil {
				allBlocks = append(allBlocks, b)
			}
		}
	}

	// URUTKAN AGAR TIDAK BERANTAKAN DI DATABASE
	sort.Slice(allBlocks, func(i, j int) bool {
		return allBlocks[i].Index < allBlocks[j].Index
	})

	for _, b := range allBlocks {
		// GUNAKAN STORE YANG SUDAH TERKONEKSI LEVELDB
		err := store.SaveBlock(b) 
		if err != nil {
			fmt.Printf("❌ Gagal migrasi blok %d: %v\n", b.Index, err)
		} else {
			fmt.Printf("✅ Blok %d berhasil dipindahkan ke LevelDB.\n", b.Index)
		}
	}

	fmt.Println("✨ Migrasi Selesai! Data Sultan sekarang aman di LevelDB.")
}
