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
	"github.com/aziskebanaran/bvm-core/pkg/types"
)

type BVMClient struct {
	BaseURL string
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

// UploadToCloud: Mengirim snapshot database Nexus ke BVM Cloud Storage Core
func (c *BVMClient) UploadToCloud(filePath string, owner string) (string, error) {
    // 1. Buka file snapshot
    file, err := os.Open(filePath)
    if err != nil { return "", err }
    defer file.Close()

    // 2. Siapkan wadah Multipart (Form Data)
    body := &bytes.Buffer{}
    writer := multipart.NewWriter(body)
    
    // Masukkan file ke form
    part, err := writer.CreateFormFile("file", filepath.Base(filePath))
    if err != nil { return "", err }
    io.Copy(part, file)
    
    // Tambahkan informasi pemilik agar Core tahu siapa yang menitip
    writer.WriteField("owner", owner)
    writer.Close()

    // 3. Tembak ke Endpoint /api/storage/put Sultan
    req, err := http.NewRequest("POST", c.BaseURL+"/api/storage/put", body)
    if err != nil { return "", err }
    req.Header.Set("Content-Type", writer.FormDataContentType())

    resp, err := c.HTTP.Do(req)
    if err != nil { return "", fmt.Errorf("🛰️ Cloud Offline: %v", err) }
    defer resp.Body.Close()

    // 4. Tangkap Storage_ID dari Core
    var result struct {
        Status    string `json:"status"`
        StorageID string `json:"storage_id"`
        Message   string `json:"message"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return "", fmt.Errorf("❌ Gagal baca respon Cloud: %v", err)
    }

    if result.Status != "success" {
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
