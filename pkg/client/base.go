package client

import (
	"bytes"
	"io"
	"mime/multipart"
 	"os"
	"path/filepath"
	 "encoding/json"
	"net/http"
	"time"
	"fmt"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"github.com/aziskebanaran/bvm-core/pkg/storage"
)


type BVMClient struct {
    BaseURL string
    Token   string // 🚩 Tambahkan ini sebagai kantong penyimpanan token
    HTTP    *http.Client
}

func NewBVMClient(url string) *BVMClient {
	return &BVMClient{
		BaseURL: url,
		HTTP:    &http.Client{Timeout: 120 * time.Second},
	}
}

// GetNetworkInfo: Mengambil data gabungan (Params + Realtime Status)
func (c *BVMClient) GetNetworkInfo() (*types.NetworkResponse, error) {
    resp, err := c.HTTP.Get(c.BaseURL + "/api/params")
    if err != nil { 
        return nil, fmt.Errorf("🌐 Jaringan Offline: %v", err) 
    }
    defer resp.Body.Close()

    var info types.NetworkResponse
    
    // Bongkar JSON langsung ke Struct agar aman dan cepat
    if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
        return nil, fmt.Errorf("❌ Format data tidak dikenal: %v", err)
    }
    
    return &info, nil
}

// MempoolResponse: Struktur data untuk menangkap antrean dari API
type MempoolResponse struct {
    Count   int                 `json:"count"`
    Txs     []types.Transaction `json:"txs"`
    Status  string              `json:"status"`
    Message string              `json:"message"`
}

// GetMempool: Fungsi kurir untuk mengambil antrean transaksi di RAM
func (c *BVMClient) GetMempool() (*MempoolResponse, error) {
    // Tembak ke endpoint /api/mempool sesuai Router Sultan
    resp, err := c.HTTP.Get(c.BaseURL + "/api/mempool")
    if err != nil {
        return nil, fmt.Errorf("🌐 Node Offline: %v", err)
    }
    defer resp.Body.Close()

    var result MempoolResponse
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, fmt.Errorf("❌ Gagal bongkar data Mempool: %v", err)
    }

    return &result, nil
}

// UploadToCloud: Versi 3 Parameter agar Sinkron dengan Nexus
func (c *BVMClient) UploadToCloud(filePath string, owner string, apiKey string) (string, error) {
    // 1. Buka file snapshot
    file, err := os.Open(filePath)
    if err != nil { return "", err }
    defer file.Close()

    // 2. Siapkan wadah Multipart
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)

    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil { return "", err }
    io.Copy(part, file)

    // Tambahkan info ke form
    writer.WriteField("owner", owner)
    writer.WriteField("app_id", "Nexus-Alpha") 
    writer.Close()

    // 3. Tembak ke Core
    req, err := http.NewRequest("POST", c.BaseURL+"/api/storage/put", body)
    if err != nil { return "", err }
    
    // 🚩 PENTING: Masukkan API Key ke Header agar Core mengizinkan akses
    req.Header.Set("Content-Type", writer.FormDataContentType())
    req.Header.Set("X-BVM-API-KEY", apiKey) 

    resp, err := c.HTTP.Do(req)
    if err != nil { return "", fmt.Errorf("🛰️ Cloud Offline: %v", err) }
    defer resp.Body.Close()

    // 4. Tangkap Respon
    var result struct {
        Status    string `json:"status" json:"status"`
        StorageID string `json:"storage_id" json:"storage_id"`
        Message   string `json:"message" json:"message"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("❌ Gagal baca respon Cloud: %v", err)
    }

    // Karena di Core kita menggunakan "success" (huruf kecil), sesuaikan di sini
    if result.Status != "success" && result.Status != "SUCCESS" {
        return "", fmt.Errorf("🚨 Core Menolak: %s", result.Message)
    }

    return result.StorageID, nil
}

func (c *BVMClient) RegisterApp(appID, owner string) (string, error) {
    reqBody, _ := json.Marshal(map[string]string{
        "app_id": appID,
        "owner":  owner,
    })

    resp, err := c.HTTP.Post(c.BaseURL+"/api/storage/register", "application/json", bytes.NewBuffer(reqBody))
    if err != nil { return "", err }
    defer resp.Body.Close()

    var res struct {
        Status string `json:"status"`
        ApiKey string `json:"api_key"`
    }
    json.NewDecoder(resp.Body).Decode(&res)

    return res.ApiKey, nil
}

// GetBlockByHeight: Menjemput satu blok spesifik dari Core
func (c *BVMClient) GetBlockByHeight(height uint64) (*types.Block, error) {
    // Tembak ke endpoint /api/block/[height]
    url := fmt.Sprintf("%s/api/block/%d", c.BaseURL, height)

    resp, err := c.HTTP.Get(url)
    if err != nil {
        return nil, fmt.Errorf("🌐 Gagal kontak Core: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        if resp.StatusCode == http.StatusNotFound {
            return nil, fmt.Errorf("🚨 Blok #%d belum ada (Core belum sampai sana)", height)
        }
        return nil, fmt.Errorf("🚨 Core Error (Status: %d)", resp.StatusCode)
    }

    var block types.Block
    if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
        return nil, fmt.Errorf("❌ Format blok tidak valid: %v", err)
    }

    return &block, nil
}

func (c *BVMClient) FastSync(start uint64, target uint64, store storage.BVMStore) error {
    fmt.Printf("⚡ [FASTSYNC] Memulai akselerasi dari #%d ke #%d...\n", start, target)
    
    // Kita tarik dalam kelompok per 100 blok agar tidak membebani RAM
    const batchSize = 100 
    
    for current := start + 1; current <= target; {
        end := current + batchSize
        if end > target {
            end = target
        }

        fmt.Printf("📥 Menarik batch blok #%d sampai #%d...\n", current, end)
        
        // Panggil endpoint Nexus yang sudah kita buat
        for h := current; h <= end; h++ {
            block, err := c.GetBlockByHeight(h)
            if err != nil {
                return fmt.Errorf("Gagal di blok %d: %v", h, err)
            }
            
            // Simpan ke database lokal perangkat baru
            store.SaveBlock(*block)
            store.Put("m:height", h)
        }
        
        current = end + 1
    }
    
    fmt.Println("✅ [FASTSYNC] Sinkronisasi Selesai! Perangkat sudah mutakhir.")
    return nil
}
