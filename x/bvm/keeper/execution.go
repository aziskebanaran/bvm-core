package keeper

import (
    "encoding/json" // 🚩 WAJIB ADA
    "strings"
	"time"
    "fmt"
    "github.com/aziskebanaran/bvm-core/pkg/logger"
    "github.com/aziskebanaran/bvm-core/pkg/ai"
    "github.com/aziskebanaran/bvm-core/x/bvm/types"
    bankkeeper "github.com/aziskebanaran/bvm-core/x/bank/keeper"
    banktypes "github.com/aziskebanaran/bvm-core/x/bank/types" // 🚩 WAJIB ADA
    "github.com/aziskebanaran/bvm-core/x/events"
    utxotypes "github.com/aziskebanaran/bvm-core/x/utxo/types"
)

func (k *Keeper) ExecuteBlock(block types.Block) error {
    batch := k.Store.NewBatch()
    var totalFees uint64 = 0

    // Kita gunakan Map sementara untuk BVM agar tidak menulis ke batch berulang kali
    // untuk alamat yang sama di dalam blok yang sama (Optimasi I/O)
    pendingChanges := make(map[string]int64)
    pendingNonces := make(map[string]uint64)

    // --- 🚩 THE SHIFT: AKTIVASI KONSTITUSI MODERN ---
    if block.Index == 6000 {
        logger.Info("SYSTEM", "⚔️ MEMASUKI ERA MODERN: Mengaktifkan Governance Contract...")
        err := k.InitializeGovernance()
        if err != nil {
            // Jika gagal muat kontrak, blok tidak boleh dieksekusi demi keamanan
            return fmt.Errorf("🚨 KRITIS: Gagal aktivasi Era Modern di blok 6000: %v", err)
        }
    }

    // --- 1. PROSES TRANSAKSI (BVM & WASM) ---
    for _, tx := range block.Transactions {


        k.Store.PutHistoryToBatch(batch, tx.From, tx)
        // Catat untuk Penerima
        k.Store.PutHistoryToBatch(batch, tx.To, tx)
        // Simpan detail TXID (untuk pencarian cepat)
        k.Store.PutToBatch(batch, "tx:"+tx.ID, tx)

        // A. Potong Fee dari Pengirim
        totalFees += tx.Fee
        pendingChanges[tx.From] -= int64(tx.Fee)

        switch tx.Type {

        case "mempool_report":
            // 📡 JALUR OTORITAS KHUSUS: Mempool Standalone mengirim laporan status / detak jantung
            logger.Info("CORE", fmt.Sprintf("📢 Laporan Jaringan Masuk dari Mempool Node: %s", tx.From[:12]))
            // Contoh pembongkaran payload jika mempool mengirim data status JSON
            var reportData map[string]interface{}
            if err := json.Unmarshal(tx.Payload, &reportData); err == nil {
                logger.Info("CORE", fmt.Sprintf("📊 Mempool Status -> Total Queue: %v", reportData["queue_count"]))
            }

            // Daftarkan kontribusi mempool secara otomatis ke disk metadata jika diperlukan
            k.Store.PutToBatch(batch, "mempool_active:"+tx.From, time.Now().Unix())
            logger.Success("CORE", fmt.Sprintf("✅ Laporan Mempool [%s] Berhasil Di-arsip!", tx.From[:8]))


    case "utxo_move":
        var utxoData utxotypes.UTXOData
        if err := json.Unmarshal(tx.Payload, &utxoData); err != nil {
            logger.Error("EXEC", "❌ Gagal bedah Payload UTXO")
            continue
        }

        // 🛡️ Safety check tambahan sebelum pembongkaran kepingan disk dilakukan
        if tx.ChainID > 0 && tx.ChainID != k.GetParamsData().GetChainID() {
            logger.Error("EXEC", fmt.Sprintf("🚨 Upaya eksekusi UTXO ilegal digagalkan! Tx ChainID Mismatch: %d", tx.ChainID))
            continue
        }

        // 1. Eksekusi UTXO (Hancurkan kepingan lama)
        err := k.UTXO.SendAset(tx.From, tx.To, tx.Symbol, tx.Amount, tx.Fee, tx.ID, tx.Payload, batch)

        if err != nil {
            logger.Error("EXEC", fmt.Sprintf("❌ Gagal Eksekusi UTXO: %v", err))
            continue
        }

        // 🚩 2. LOGIKA PENYAMBUNG JEMBATAN (The Bridge Back)
        // Jika target BUKAN 0x, berarti koin masuk ke Ruang Account (bvmf)
        if !strings.HasPrefix(tx.To, "0x") {
            if tx.Symbol == "BVM" || tx.Symbol == k.Params.NativeSymbol {
                // Tambahkan ke map pendingChanges agar nanti di-commit ke State Account
                pendingChanges[tx.To] += int64(tx.Amount)
                logger.Success("EXEC", fmt.Sprintf("🔓 %d %s Dicairkan ke Saldo Account @%s", tx.Amount, tx.Symbol, tx.To[:10]))
            } else {
                // Untuk Token selain BVM, gunakan Bank Keeper
                bank := k.GetBank()
                if b, ok := bank.(*bankkeeper.BankKeeper); ok { b.Batch = batch }
                bank.AddBalance(tx.To, tx.Amount, tx.Symbol)
            }
        }

        logger.Success("EXEC", fmt.Sprintf("💎 UTXO %s Berhasil Dipahat!", tx.ID[:8]))


        case "user_register":
            // 1. Bongkar Payload (JSON) untuk mengambil Username
            var data struct {
                Username string `json:"username"`
            }
            if err := json.Unmarshal(tx.Payload, &data); err != nil {
                logger.Error("AUTH", fmt.Sprintf("❌ Gagal Unmarshal Payload Registrasi dari %s", tx.From[:10]))
                continue
            }

            // 2. Eksekusi Pendaftaran melalui AuthKeeper
            // Ingat: tx.From adalah alamat dompet user
            err := k.Auth.RegisterUser(data.Username, tx.From)
            if err != nil {
                logger.Error("AUTH", fmt.Sprintf("❌ Registrasi Gagal [%s]: %v", data.Username, err))
                continue
            }

            // 3. Catat keberhasilan di Log
            logger.Success("AUTH", fmt.Sprintf("👤 User Terdaftar: @%s -> %s", data.Username, tx.From[:10]))

        case "bridge_out":
            // 1. Ekstrak tiket tujuan
            var bridgeData struct {
                TargetChainID uint64 `json:"target_chain_id"`
            }
            if err := json.Unmarshal(tx.Payload, &bridgeData); err != nil {
                logger.Error("BRIDGE", fmt.Sprintf("❌ Gagal membaca tiket tujuan dari %s", tx.From[:8]))
                continue
            }

            // 2. Kunci Saldo Pengirim ke Brankas Bridge!
            pendingChanges[tx.From] -= int64(tx.Amount)
            pendingChanges["bvm_bridge_vault_locked"] += int64(tx.Amount)

            // 3. Cetak Kuitansi (Event) agar dibaca oleh Nexus
            logger.Success("BRIDGE", fmt.Sprintf("🔒 %d %s Dikunci! Menunggu keberangkatan ke Chain %d", tx.Amount, tx.Symbol, bridgeData.TargetChainID))
            
            // Opsional: Rekam log spesifik untuk Nexus
            k.Store.PutToBatch(batch, "bridge_out_log:"+tx.ID, tx)

	// 🚩 TRIGGER RELAYER LANGSUNG DARI ENGINE (The Sultan Hook)
	events.EmitEvent("BRIDGE_OUT_READY", map[string]interface{}{
	    "tx_id":           tx.ID,
	    "target_chain_id": bridgeData.TargetChainID,
	    "amount":          tx.Amount,
	})


case "bridge_in":
    // 1. Verifikasi Identitas Relayer (Keamanan Utama)
    if !k.Bridge.IsAuthorizedRelayer(tx.From) {
        logger.Error("BRIDGE", "🚨 Relayer tidak sah!")
        continue
    }

    // 2. Verifikasi Tanda Tangan (Wajib untuk integritas)
    if !k.Bridge.VerifyRelayerSignature(tx.Payload, tx.Signature) {
        logger.Error("BRIDGE", "🚨 Tanda tangan Nexus palsu!")
        continue
    }

    // 3. Ekstrak Bukti TxID
    var inData struct {
        RefTxID string `json:"ref_tx_id"`
    }
    if err := json.Unmarshal(tx.Payload, &inData); err != nil {
        logger.Error("BRIDGE", "❌ Format payload rusak!")
        continue
    }

    // 4. Verifikasi Status Lock (Cek apakah proof sudah tercatat di SubmitProof)
    if !k.Bridge.VerifySourceChainLock(inData.RefTxID) {
        logger.Error("BRIDGE", "⚠️ Tx asal belum terkunci (Lock Proof hilang)!")
        continue
    }

    // 5. Cek Double-Spend (Idempotency)
    var isClaimed string
    errClaim := k.Store.Get("bridge_claimed:"+inData.RefTxID, &isClaimed)
    if errClaim == nil && isClaimed == "true" {
        logger.Error("BRIDGE", "⚠️ Transaksi sudah pernah diklaim!")
        continue
    }

    // 6. Eksekusi Saldo (Update state melalui pendingChanges)
    pendingChanges[tx.To] += int64(tx.Amount)
    k.Store.PutToBatch(batch, "bridge_claimed:"+inData.RefTxID, []byte("true"))

    logger.Success("BRIDGE", fmt.Sprintf("🌉 KEDATANGAN SUKSES! %d BVM ke %s", tx.Amount, tx.To[:8]))



case "submit_proof":
    // Hanya relayer yang boleh mengirim bukti kunci
    if !k.Bridge.IsAuthorizedRelayer(tx.From) { continue }

    // Simpan bukti ke KVStore agar nanti bisa dicek oleh bridge_in
    var proofData struct { RefTxID string }
    json.Unmarshal(tx.Payload, &proofData)
    k.Bridge.RecordLockProof(proofData.RefTxID)

    logger.Success("BRIDGE", "✅ Bukti kunci diterima: " + proofData.RefTxID[:8])


		case "contract_call":
		    // 🚩 TAMBAHKAN PENJAGA GERBANG DI SINI
		    if !k.IsFeatureActive("WASM_ENGINE", int64(block.Index)) {
			logger.Error("WASM", fmt.Sprintf("⚠️ Blok #%d: Fitur WASM belum aktif secara konstitusi.", block.Index))

		        // Jika belum aktif, kita skip transaksinya (atau anggap gagal)
		        // Fee tetap hangus (totalFees += tx.Fee tetap jalan di atas)
		        continue 
		    }

		    // Jika lolos (IsFeatureActive == true), baru eksekusi mesin beratnya
		    logger.Info("WASM", fmt.Sprintf("🚀 Kontrak Dipicu: %s", tx.To))
		    err := k.Wasm.ExecuteContractWithBatch(batch, tx.To, tx.From, tx.Payload)
		    if err != nil {
		        logger.Error("WASM", fmt.Sprintf("❌ Gagal: %v", err))
	        continue
	    }

        case "create_token":
		if !k.IsFeatureActive("TOKEN_FACTORY", int64(block.Index)) {
                logger.Error("BANK", fmt.Sprintf("⚠️ Blok #%d: Pabrik Token belum diizinkan beroperasi.", block.Index))
                continue // Transaksi dibatalkan, lanjut ke transaksi berikutnya
            }

            var msg banktypes.MsgCreateToken
            if err := json.Unmarshal(tx.Payload, &msg); err != nil {
                logger.Error("EXEC", "Gagal unmarshal MsgCreateToken")
                continue
            }

            bank := k.GetBank()

            // 🚩 Kita lakukan casting ke struct asli untuk akses field .Batch
            if b, ok := bank.(*bankkeeper.BankKeeper); ok {
                b.Batch = batch

                // Sekarang Go tidak akan protes lagi karena HandleMsgCreateToken 
                // sudah ada di struct dan interface
                err := b.HandleMsgCreateToken(msg) 
                if err != nil {
                    logger.Error("BANK", fmt.Sprintf("❌ Gagal membuat token %s: %v", msg.Symbol, err))
                    continue
                }
                logger.Success("BANK", fmt.Sprintf("💎 Token %s Berhasil Terbit!", msg.Symbol))
            }

        case "stake":
            // 1. Cek Aktivasi Fitur (Staking V2)
            if !k.IsFeatureActive("STAKING_V2", int64(block.Index)) {
                logger.Error("STAKING", fmt.Sprintf("⚠️ Blok #%d: Staking V2 belum aktif.", block.Index))
                continue
            }

            // 2. Eksekusi Native Staking
            // Kita ambil Keeper Staking Sultan
            staking := k.GetStaking()

            // Masukkan saldo ke gudang staking (Native)
            err := staking.Stake(tx.From, tx.Amount)
            if err != nil {
                logger.Error("STAKING", fmt.Sprintf("❌ Gagal Stake: %v", err))
                continue
            }

            // 3. 🚩 THE BRIDGE: Update Power ke WASM DPoS
            // Inilah saatnya dpos.go Sultan bekerja!
            if k.IsFeatureActive("WASM_ENGINE", int64(block.Index)) {
                payload := map[string]interface{}{
                    "method": "delegate",
                    "amount": tx.Amount,
                }
                payloadJSON, _ := json.Marshal(payload)

                // Panggil Kontrak DPoS
                k.Wasm.ExecuteContractWithBatch(batch, "dpos_contract", tx.From, payloadJSON)
                logger.Success("STAKING", fmt.Sprintf("⚡ Power %s diperbarui di WASM!", tx.From[:8]))
            }

	case "unstake":
	    // Panggil fungsi Unstake yang sudah Sultan buat di Staking Keeper
	    err := k.GetStaking().Unstake(tx.From, tx.Amount)
	    if err != nil {
	        logger.Error("STAKING", fmt.Sprintf("❌ Gagal Unstake: %v", err))
	        continue
	    }

	    // Jika WASM Aktif, kabari kontrak DPoS bahwa power berkurang
	    if k.IsFeatureActive("WASM_ENGINE", int64(block.Index)) {
	        payload, _ := json.Marshal(map[string]interface{}{
	            "method": "undelegate", // Nama fungsi di dpos.go Sultan
	            "amount": tx.Amount,
	        })
	        k.Wasm.ExecuteContractWithBatch(batch, "dpos_contract", tx.From, payload)
	    }
	    logger.Success("STAKING", fmt.Sprintf("🔓 %s menarik stake sebesar %d", tx.From[:8], tx.Amount))


default: // Transfer Standar
    if tx.Symbol == "BVM" || tx.Symbol == k.Params.NativeSymbol {
        pendingChanges[tx.From] -= int64(tx.Amount) 

        if strings.HasPrefix(tx.To, "0x") {
            // 📥 MASUK KE RUANG UTXO (0x)
            // Jika dikirim via API lokal dan ChainID tx kosong, kita otomatis wariskan ChainID node
            if tx.ChainID == 0 {
                tx.ChainID = k.GetParamsData().GetChainID()
            }
            k.UTXO.MintUTXO(tx.To, tx.To, "AUTO_USER", tx.Symbol, tx.Amount, tx.ID, batch)
            logger.Success("EXEC", fmt.Sprintf("📥 %d %s Berhasil Masuk Brankas UTXO 0x (ChainID: %d)", tx.Amount, tx.Symbol, tx.ChainID))
        } else {
            pendingChanges[tx.To] += int64(tx.Amount)
        }

    } else {
        // Logika untuk Token Factory tetap menggunakan Bank Keeper
        bank := k.GetBank()
        if b, ok := bank.(*bankkeeper.BankKeeper); ok { b.Batch = batch }

        bank.SubBalance(tx.From, tx.Amount, tx.Symbol)

        if strings.HasPrefix(tx.To, "0x") {
            // Token juga bisa dipindahkan ke ruang UTXO 0x!
            k.UTXO.MintUTXO(tx.To, tx.To, "AUTO_USER", tx.Symbol, tx.Amount, tx.ID, batch)
        } else {
            bank.AddBalance(tx.To, tx.Amount, tx.Symbol)
        }
    }
}

        // =========================================================================
        // 🚩 B. PEMBAGIAN FEE DENGAN DELEGASI KE REWARD KEEPER
        // =========================================================================
        actualFee := tx.Fee

        // Panggil fungsi pemotong jatah Mempool dari reward.go
        relayerCut := k.DistributeRelayerReward(tx.Relayer, tx.ID, tx.Fee, uint64(block.Index), pendingChanges, batch)
        actualFee -= relayerCut

        // Lanjutkan pembagian sisa fee untuk Miner tip & burn
        tip, burn := k.Params.DistributeFee(actualFee)

        if tx.Symbol == "BVM" || tx.Symbol == "" {
            pendingChanges[block.Miner] += int64(tip)
            k.TotalSupplyBVM -= burn
        } else {
            bank := k.GetBank()
            if b, ok := bank.(*bankkeeper.BankKeeper); ok { b.Batch = batch }

            if tip > 0 { bank.AddBalance(block.Miner, tip, tx.Symbol) }
            if burn > 0 { bank.Burn(k.Params.BurnAddress, burn, tx.Symbol) }
        }


        // C. Update Nonce
        if tx.Nonce >= pendingNonces[tx.From] {
            pendingNonces[tx.From] = tx.Nonce + 1
        }
    }

    // --- 2. DISTRIBUSI REWARD BLOK (BVM) REVOLUSI ADALAH NYATA ---
    k.PromoteMinerToValidator(block.Miner, batch)

    // 🎯 KODE BARU AMAN (Suntikkan block.Miner secara presisi):
    _, _, errBlock := k.DistributeBlockReward(int64(block.Index), block.Miner, totalFees, batch, pendingChanges)

    if errBlock != nil {
        logger.Error("EXEC", fmt.Sprintf("❌ Gagal distribusi block reward: %v", errBlock))
    }


    // --- 3. KOMITMEN BVM KE BATCH (Satu Pintu) ---
	for addr, change := range pendingChanges {
	    oldBal := k.GetBalanceBVM(addr)
    // Hitung saldo baru (bisa naik atau turun)
	    newBal := uint64(int64(oldBal) + change)

	    // Simpan langsung ke batch menggunakan helper keeper Sultan
	    k.SetBalanceBVM(addr, newBal, batch) 
	}

    // --- 4. KOMITMEN NONCE & METADATA ---
    for addr, nextNonce := range pendingNonces {
        k.NonceMgr.SetNonce(addr, nextNonce) 
        k.Store.PutToBatch(batch, "n:"+addr, nextNonce)
    }

    // Update Metadata Global
    absoluteSupply := k.Params.GetExpectedSupply(int64(block.Index))
    k.Store.PutToBatch(batch, k.keyMeta("height"), uint64(block.Index))
    k.Store.PutToBatch(batch, k.keyMeta("hash"), block.Hash)
    k.Store.PutToBatch(batch, k.keyMeta("supply"), absoluteSupply)
    k.Store.PutToBatch(batch, k.keyBlock(int64(block.Index)), block)

    // --- 5. EKSEKUSI PAHAT DISK (FINAL) ---
    if err := k.Store.WriteBatch(batch); err != nil {
        return fmt.Errorf("🚨 Gagal Pahat Disk: %v", err)
    }

    // --- 6. UPDATE RAM ---
    k.Blockchain.Height = int64(block.Index)
    k.Blockchain.LatestHash = block.Hash
    k.TotalSupplyBVM = absoluteSupply

    k.FinalizeBlock(block)

    logger.Success("EXECUTE", fmt.Sprintf("✅ Blok #%d Sukses!", block.Index))
    return nil
}


func (k *Keeper) CommitBlock(block types.Block) error {

    if len(block.Transactions) > 0 {
        k.Mempool.RemoveUsedTransactions(block.Transactions)
    }

    k.Blockchain.InFlight = 0

    vCount := k.GetValidatorCount()
    reward := k.GetSubsidiAtHeight(int64(block.Index), vCount)

    events.EmitEvent("NEW_BLOCK_COMMITTED", map[string]interface{}{
        "height": block.Index,
        "hash":   block.Hash,
        "miner":  block.Miner,
        "reward": k.Params.FormatDisplay(reward),
    })

    logger.Success("COMMIT", fmt.Sprintf("🧱 Blok #%d Sah & Bersih!", block.Index))

    // --- 🤖 AKTIVASI KOMANDAN AI (TAMBAHKAN DI SINI) ---
    // Gunakan goroutine (go ...) agar AI tidak menghambat kecepatan blockchain (Async)
    go func() {
        orchestrator := ai.NewOrchestrator()

        // Kita hitung jumlah wallet unik dalam transaksi blok ini
        activeWallets := k.extractUniqueWallets(block.Transactions)

        // Perintahkan AI untuk melapor
        orchestrator.ReportNewBlock(int64(block.Index), len(block.Transactions), activeWallets)
    }()

    return nil
}



func (k *Keeper) CreateNextBlock(minerAddr string) types.Block {
    // 1. Ambil transaksi yang benar-benar menunggu
    txs := k.Mempool.GetTransactions(100)
    if txs == nil {
        txs = []types.Transaction{}
    }

    status := k.GetStatus()
    lastHeight := k.GetLastHeight()
    nextHeight := lastHeight + 1

    status.Height = int64(lastHeight)
    status.Difficulty = int32(k.GetNextDifficultyForMiner(minerAddr))
    status.Reward = k.GetCurrentReward(status.Height + 1)

    // 2. Rakit Blok
    block := types.NewMiningBlock(status, txs, minerAddr, "BVM-NODE")

    // 3. LOGIKA NAVIGASI (Anchor vs Normal)
    if nextHeight > 1 && nextHeight % 10 == 1 {
        var anchor string
        err := k.Store.Get(k.keyMeta("cycle_anchor"), &anchor)
        if err == nil && anchor != "" {
            block.PrevHash = anchor
            logger.Info("MINER", fmt.Sprintf("🔗 SIKLUS BARU #%d: Menggunakan Anchor.", nextHeight))
        } else {
            block.PrevHash = k.Blockchain.LatestHash
        }
    } else {
        block.PrevHash = k.Blockchain.LatestHash
    }

    block.Hash = block.CalculateBlockHash()
    return block
}
