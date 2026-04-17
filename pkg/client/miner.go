package client

import (
	"bytes"
	"encoding/json"
	"github.com/aziskebanaran/bvm-core/x/bvm/types"
	"fmt"
	"net/http"
)

func (c *BVMClient) GetNodeStatus(minerAddr string) (*types.NodeStatus, error) {
        resp, err := c.HTTP.Get(fmt.Sprintf("%s/api/status?address=%s", c.BaseURL, minerAddr))
        if err != nil {
                return nil, err
        }
        defer resp.Body.Close()

        var status types.NodeStatus
        err = json.NewDecoder(resp.Body).Decode(&status)
        if err != nil {
                return nil, fmt.Errorf("gagal decode status: %v", err)
        }

        return &status, nil
}


func (c *BVMClient) SubmitBlock(block types.Block) error {
    payload, err := json.Marshal(block)
    if err != nil { return err }

    req, _ := http.NewRequest("POST", c.BaseURL+"/api/mine", bytes.NewBuffer(payload))
    req.Header.Set("Content-Type", "application/json")

    resp, err := c.HTTP.Do(req)
    if err != nil { 
        return fmt.Errorf("🌐 Jaringan Terputus/Timeout: %v", err) 
    }
    defer resp.Body.Close()

    if resp.StatusCode == http.StatusOK || resp.StatusCode == http.StatusAccepted {
        return nil 
    }

    if resp.StatusCode == http.StatusConflict {
        return fmt.Errorf("⏩ Blok #%d sudah diproses oleh Kernel sebelumnya", block.Index)
    }

    return fmt.Errorf("❌ Kernel Menolak (Status: %d)", resp.StatusCode)
}


func (c *BVMClient) GetMempoolTxs() ([]types.Transaction, error) {
    resp, err := c.HTTP.Get(c.BaseURL + "/api/mempool")
    if err != nil { return nil, err }
    defer resp.Body.Close()

    var wrapper struct {
        Txs []types.Transaction `json:"txs"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&wrapper); err != nil {
        return []types.Transaction{}, nil 
    }

    if wrapper.Txs == nil {
        return []types.Transaction{}, nil
    }

    return wrapper.Txs, nil
}

// pkg/client/client.go (atau di mana BVMClient didefinisikan)

func (c *BVMClient) GetWork(minerAddr string, minerName string) (interface{}, error) {
    // 🚩 Hubungi API Kernel untuk mendapatkan paket kerja terbaru
    url := fmt.Sprintf("%s/api/getwork?address=%s&miner=%s", c.BaseURL, minerAddr, minerName)

    resp, err := http.Get(url)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        return nil, fmt.Errorf("server return status: %d", resp.StatusCode)
    }

    var block types.Block
    if err := json.NewDecoder(resp.Body).Decode(&block); err != nil {
        return nil, err
    }

    // Pastikan block yang diterima tidak kosong
    return block, nil
}
