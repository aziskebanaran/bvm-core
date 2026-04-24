package types

import (
        "bytes"
        "encoding/binary"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"
)

// CalculateHash: Menghasilkan ID unik untuk seluruh isi blok
func (b *Block) CalculateHash() ([]byte, error) {
    buf := new(bytes.Buffer)

    // --- KELOMPOK 1: ANGKA (Tetap/Fixed) ---
    // Gunakan tipe data eksplisit (int64/int32) agar konsisten di semua OS
    binary.Write(buf, binary.LittleEndian, int32(b.Version))
    binary.Write(buf, binary.LittleEndian, int64(b.Index))
    binary.Write(buf, binary.LittleEndian, b.Timestamp)
    binary.Write(buf, binary.LittleEndian, int32(b.Difficulty))
    binary.Write(buf, binary.LittleEndian, b.Nonce)
    binary.Write(buf, binary.LittleEndian, b.Reward)
    binary.Write(buf, binary.LittleEndian, b.TotalFee)

    // --- KELOMPOK 2: STRING (Dinamis/Variable) ---
    writeStringWithLength(buf, b.PrevHash)
    writeStringWithLength(buf, b.MerkleRoot)
    writeStringWithLength(buf, b.Miner)
    writeStringWithLength(buf, b.MinerName)
    writeStringWithLength(buf, b.StateRoot)

    hash := sha256.Sum256(buf.Bytes())
    return hash[:], nil
}

// Helper writeStringWithLength (Tetap seperti milik Sultan, sudah bagus)
func writeStringWithLength(buf *bytes.Buffer, s string) {
    // Menulis panjang string sebagai int32 sebelum isinya
    binary.Write(buf, binary.LittleEndian, int32(len(s)))
    buf.WriteString(s)
}


// CalculateBlockHash: Mengembalikan string Hex dari hash blok
func (b *Block) CalculateBlockHash() string {
	h, _ := b.CalculateHash()
	return hex.EncodeToString(h)
}

// --- RUMUS CEK TARGET (NOL) ---
func (b Block) HasValidTarget(hash string) bool {
	target := strings.Repeat("0", int(b.Difficulty))
	return strings.HasPrefix(hash, target)
}

// --- RUMUS MERKLE ROOT (VERSI OPTIMAL) ---
func CalculateMerkleRoot(txs []Transaction) string {
    if len(txs) == 0 {
        return "0000000000000000000000000000000000000000000000000000000000000000"
    }

    var hashes [][]byte // Gunakan byte slice agar lebih cepat dan akurat
    for _, tx := range txs {
        h, _ := tx.CalculateHash()
        hashes = append(hashes, h)
    }

    for len(hashes) > 1 {
        if len(hashes)%2 != 0 {
            hashes = append(hashes, hashes[len(hashes)-1])
        }
        var nextLevel [][]byte
        for i := 0; i < len(hashes); i += 2 {
            // Gabungkan byte, bukan string hex, untuk keamanan kriptografi
            combined := append(hashes[i], hashes[i+1]...)
            hash := sha256.Sum256(combined)
            nextLevel = append(nextLevel, hash[:])
        }
        hashes = nextLevel
    }
    return hex.EncodeToString(hashes[0])
}

// NewMiningBlock: Menambahkan validasi Index
func NewMiningBlock(status NodeStatus, txs []Transaction, minerAddr string, minerName string) Block {
    parentHash := status.LatestHash
    // Pastikan parentHash tidak kosong
    if parentHash == "" || parentHash == "0" || len(parentHash) < 64 {
        parentHash = strings.Repeat("0", 64)
    }

    var totalFee uint64 = 0
    for _, tx := range txs {
        totalFee += tx.Fee
    }

    return Block{
        Version:      status.Version,
        Index:        int64(status.Height + 1), // Blok selanjutnya
        Timestamp:    time.Now().Unix(),
        PrevHash:     parentHash,
        MerkleRoot:   CalculateMerkleRoot(txs),
        Miner:        minerAddr,
        MinerName:    minerName,
        Difficulty:   status.Difficulty,
        Reward:       status.Reward,
        TotalFee:     totalFee,
        StateRoot:    status.StateRoot, // State saat ini
        Transactions: txs,
        Nonce:        0,
    }
}
