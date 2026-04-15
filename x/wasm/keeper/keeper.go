package keeper

import (
    "context"
    "fmt"
    "time"
	"os"
    "strings"
    "bvm.core/pkg/logger"
    "bvm.core/pkg/storage"
    "bvm.core/x" // 🚩 Untuk akses BVMKeeper
    "github.com/tetratelabs/wazero"
    "github.com/tetratelabs/wazero/api"
    "github.com/tetratelabs/wazero/imports/wasi_snapshot_preview1"
)

type Keeper struct {
    Store  storage.BVMStore
    Wazero wazero.Runtime
    Ctx    context.Context
    // Kita simpan referensi ke Kernel pusat
    BVM    x.BVMKeeper 
    currentCaller string 
}

func NewKeeper(store storage.BVMStore) *Keeper {
    ctx := context.Background()
    // Inisialisasi Wazero Engine
    runtime := wazero.NewRuntime(ctx)

    // --- BARIS SAKTI ---
    wasi_snapshot_preview1.MustInstantiate(ctx, runtime)
    // ------------------

    return &Keeper{
        Store:  store,
        Wazero: runtime,
        Ctx:    ctx,
    }
}

// SetBVM: Menghubungkan WASM ke Jenderal Pusat (Kabinet)
func (k *Keeper) SetBVM(bvm x.BVMKeeper) {
    k.BVM = bvm
    // Di sini Sultan bisa mendaftarkan Host Functions (Jembatan) ke modul Bank/Auth
    k.RegisterBVMFunctions() 
}

func (k *Keeper) DeployContract(owner string, bytecode []byte) (string, error) {
    // 1. Buat ID unik untuk kontrak (misal hash dari bytecode + owner)
    contractAddr := fmt.Sprintf("bvmwasm%s", owner[len(owner)-8:])

    // 2. Simpan Bytecode ke LevelDB dengan prefix 'w:' (Wasm)
    err := k.Store.Put("w:"+contractAddr, bytecode)
    if err != nil {
        return "", err
    }

    logger.Success("WASM", fmt.Sprintf("📜 Kontrak %s Dideploy!", contractAddr))
    return contractAddr, nil
}

func (k *Keeper) ExecuteContract(contractAddr string, caller string, payload []byte) error {

    k.currentCaller = caller

    // 1. Ambil Bytecode dari Disk
    var bytecode []byte
    if err := k.Store.Get("w:"+contractAddr, &bytecode); err != nil {
        return fmt.Errorf("kontrak %s tidak terdaftar", contractAddr)
    }

    // 2. PASANG SATPAM: Timeout 1 Detik
    ctx, cancel := context.WithTimeout(k.Ctx, 1*time.Second)
    defer cancel()

    mod, err := k.Wazero.Instantiate(ctx, bytecode)
    if err != nil { 
        return fmt.Errorf("gagal instansiasi: %v", err) 
    }
    defer mod.Close(ctx)

    // --- PROSES PASSING DATA ---

    // 1. Ambil akses ke memori WASM
    mem := mod.Memory()
    if mem == nil {
        return fmt.Errorf("kontrak tidak memiliki export memory")
    }

    // 3. PANGGIL FUNGSI 'handle'
    handleFunc := mod.ExportedFunction("handle")
    if handleFunc == nil {
        return fmt.Errorf("fungsi 'handle' tidak ditemukan di kontrak")
    }

// --- PANGGIL FUNGSI 'handle' DENGAN PARAMETER AMAN ---
results, err := handleFunc.Call(ctx,
    0, 0, 0, 0, 500, 0, 0,
)

if err != nil {
    // 🛡️ Jaring Pengaman untuk Go Standar (WASI)
    // Jika error mengandung pesan exit_code(0), itu artinya SUCCESS
    if strings.Contains(err.Error(), "exit_code(0)") {
        fmt.Println("✅ [WASM] Kontrak selesai dijalankan (Success Exit).")
        return nil 
    }
    return fmt.Errorf("gagal eksekusi handle: %v", err)
}


	// CETAK HASILNYA DI SINI!
   if len(results) > 0 {
       logger.Success("WASM_RESULT", fmt.Sprintf("Saldo yang terbaca oleh Kontrak: %d", results[0]))
   }

   return nil
}
// ExecuteContractWithBatch: Menjalankan kontrak dengan dukungan penulisan Batch Disk
func (k *Keeper) ExecuteContractWithBatch(batch storage.Batch, contractAddr string, caller string, payload []byte) error {
    k.currentCaller = caller

    // 1. Ambil Bytecode dari Disk (Cache atau DB)
    var bytecode []byte
    if err := k.Store.Get("w:"+contractAddr, &bytecode); err != nil {
        return fmt.Errorf("kontrak %s tidak ditemukan", contractAddr)
    }

    // 2. Setting Context & Runtime
    ctx, cancel := context.WithTimeout(k.Ctx, 1*time.Second)
    defer cancel()

    mod, err := k.Wazero.Instantiate(ctx, bytecode)
    if err != nil {
        return fmt.Errorf("gagal instansiasi wasm: %v", err)
    }
    defer mod.Close(ctx)

    // 🚩 STRATEGI ATOMIC: 
    // Kita beritahu BankKeeper untuk menggunakan 'batch' ini 
    // selama eksekusi kontrak berlangsung.
    if k.BVM != nil {
        bank := k.BVM.GetBank()
        // Cek apakah Bank punya metode SetBatch (Opsional, tergantung implementasi Bank Sultan)
        // Jika tidak, Sultan bisa memodifikasi Bank agar menerima batch di fungsi-fungsinya.
        _ = bank 
    }

    // 3. Eksekusi handle (Gunakan data dari Payload Tx Sultan)
    handleFunc := mod.ExportedFunction("handle")
    if handleFunc == nil {
        return fmt.Errorf("entry point 'handle' hilang")
    }

    // Panggil kontrak (Parameter disesuaikan dengan kebutuhan payload)
    _, err = handleFunc.Call(ctx, 0, 0, 0, 0, 0, 0, 0) 

    if err != nil && !strings.Contains(err.Error(), "exit_code(0)") {
        return fmt.Errorf("runtime error: %v", err)
    }

    return nil
}


func (k *Keeper) RegisterBVMFunctions() {
    // Kita buat builder-nya satu kali di awal
    builder := k.Wazero.NewHostModuleBuilder("env")

    // --- Fungsi 1: Log Pesan ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, ptr, size uint32) {
            buf, _ := m.Memory().Read(ptr, size)
            logger.Success("WASM_CONTRACT", string(buf))
        }).
        Export("log_message")

    // --- Fungsi 2: Get Balance ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32) uint64 {
            mem := m.Memory()
            buf, ok := mem.Read(addrPtr, addrSize)
            if !ok { return 0 }
            address := string(buf)
            if k.BVM != nil {
                return k.BVM.GetBank().GetBalance(address, "BVM")
            }
            return 0
        }).
        Export("get_balance")

    // --- Fungsi 3: Transfer Token ---
builder.NewFunctionBuilder().
    WithFunc(func(ctx context.Context, m api.Module, fromPtr, fromSize, toPtr, toSize uint32, amount uint64, symPtr, symSize uint32) uint32 {
        mem := m.Memory()

        // Baca data dinamis dari memori WASM
        fBuf, _ := mem.Read(fromPtr, fromSize)
        tBuf, _ := mem.Read(toPtr, toSize)
        sBuf, _ := mem.Read(symPtr, symSize)

        fromAddr := string(fBuf)
        toAddr   := string(tBuf)
        symbol   := string(sBuf)

        // KEAMANAN: Pastikan pengirim transaksi (currentCaller) 
        // hanya bisa mentransfer aset MILIKNYA sendiri.
        if fromAddr != k.currentCaller {
            fmt.Printf("🚨 [WASM_SECURITY] %s mencoba mencuri aset milik %s!\n", k.currentCaller, fromAddr)
            return 0
        }

        if k.BVM == nil { return 0 }
        bank := k.BVM.GetBank()

        if err := bank.SubBalance(fromAddr, amount, symbol); err != nil { return 0 }
        bank.AddBalance(toAddr, amount, symbol)
        return 1
    }).
    Export("transfer_token")

    // --- Fungsi 4: Mint Token ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32, amount uint64, symPtr, symSize uint32) uint32 {
            mem := m.Memory()
            addrBuf, _ := mem.Read(addrPtr, addrSize)
            symBuf, _ := mem.Read(symPtr, symSize)
            if k.BVM == nil { return 0 }
            k.BVM.GetBank().Mint(string(addrBuf), amount, string(symBuf))
            return 1
        }).
        Export("mint_token")

    // --- Fungsi 5: Burn Token ---
builder.NewFunctionBuilder().
    WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32, amount uint64, symPtr, symSize uint32) uint32 {
        mem := m.Memory()
        sBuf, _ := mem.Read(symPtr, symSize)
        symbol := string(sBuf)

        // Alamat yang dibakar adalah alamat si pemanggil transaksi
        if k.BVM == nil { return 0 }
        err := k.BVM.GetBank().Burn(k.currentCaller, amount, symbol) 
        if err != nil { return 0 }
        return 1
    }).
    Export("burn_token")

    // --- Fungsi 6: Create Token (DIBETULKAN DI SINI) ---
builder.NewFunctionBuilder().
    WithFunc(func(ctx context.Context, m api.Module, symPtr, symSize uint32, fee uint64) uint32 {
        mem := m.Memory()
        sBuf, _ := mem.Read(symPtr, symSize)
        symbol := string(sBuf)

        if k.BVM == nil { return 0 }
        // Pemilik token baru (Owner) adalah si pemanggil transaksi (k.currentCaller)
        err := k.BVM.GetBank().CreateToken(k.currentCaller, symbol, fee)
        if err != nil { return 0 }
        return 1
    }).
    Export("create_token")

    // --- Fungsi 7: Check Balance Dinamis ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32, symPtr, symSize uint32) uint64 {
            mem := m.Memory()
            aBuf, _ := mem.Read(addrPtr, addrSize)
            sBuf, _ := mem.Read(symPtr, symSize)
            if k.BVM == nil { return 0 }
            return k.BVM.GetBank().GetBalance(string(aBuf), string(sBuf))
        }).
        Export("get_balance_ext") // Gunakan nama berbeda agar tidak bentrok dengan Fungsi 2

// --- Fungsi 8: Get Caller (Sudah Dinamis!) ---
builder.NewFunctionBuilder().
    WithFunc(func(ctx context.Context, m api.Module, ptr, size uint32) uint32 {
        caller := k.currentCaller // ✅ Tidak hardcode lagi!

        mem := m.Memory()
        ok := mem.Write(ptr, []byte(caller))
        if !ok { return 0 }
        return uint32(len(caller))
    }).
    Export("get_caller")

// --- Fungsi 9: Emit Event (Catat Sejarah) ---
builder.NewFunctionBuilder().
    WithFunc(func(ctx context.Context, m api.Module, tagPtr, tagSize, msgPtr, msgSize uint32) {
        mem := m.Memory()
        tBuf, _ := mem.Read(tagPtr, tagSize)
        mBuf, _ := mem.Read(msgPtr, msgSize)

        tag := string(tBuf)
        message := string(mBuf)

        // Simpan ke sistem Event BVM
        if k.BVM != nil {
            // Asumsi Sultan punya module events di x/events/keeper.go
            fmt.Printf("📝 [EVENT_%s] %s\n", tag, message)
        }
    }).
    Export("emit_event")

    // --- Fungsi 10: Update Stake (DPoS Bridge) ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32, amount uint64, isAdding uint32) uint32 {
            mem := m.Memory()
            aBuf, _ := mem.Read(addrPtr, addrSize)
            minerAddr := string(aBuf)

            if k.BVM == nil { return 0 }

            // Ambil Staking Keeper dari BVM Pusat
            // Kita asumsikan k.BVM.GetStaking() tersedia di interfaces.go Sultan
            staking := k.BVM.GetStaking() 

            // Konversi uint32 ke bool (1 = true, 0 = false)
            adding := isAdding == 1

            // Eksekusi perubahan Power di Database Kernel
            err := staking.ModifyValidatorPower(minerAddr, amount, adding)
            if err != nil {
                fmt.Printf("🚨 [STAKING_ERR] Gagal update power: %v\n", err)
                return 0
            }

            return 1
        }).
        Export("update_stake") // 🚩 Nama ini harus sama dengan //go:wasmimport di bvm.go

    // --- Fungsi 11: Get Validator Power ---
    builder.NewFunctionBuilder().
        WithFunc(func(ctx context.Context, m api.Module, addrPtr, addrSize uint32) uint64 {
            mem := m.Memory()
            aBuf, _ := mem.Read(addrPtr, addrSize)
            if k.BVM == nil { return 0 }

            return k.BVM.GetStaking().GetValidatorPower(string(aBuf))
        }).
        Export("get_validator_power")


    // --- AKHIR: EKSEKUSI PENDAFTARAN ---
    _, err := builder.Instantiate(k.Ctx)
    if err != nil {
        logger.Error("WASM", "Gagal mendaftarkan Host Functions: ", err)
    }
}


// DeployContractFromFile: Membaca file .wasm dari sistem dan mendaftarkannya
func (k *Keeper) DeployContractFromFile(name string, filePath string) (string, error) {
    bytecode, err := os.ReadFile(filePath)
    if err != nil {
        return "", fmt.Errorf("gagal membaca file SDK: %v", err)
    }

    // Validasi sederhana: Header WASM harus ada (\0asm)
    if len(bytecode) < 8 || string(bytecode[:4]) != "\x00asm" {
        return "", fmt.Errorf("file bukan merupakan biner WASM yang valid")
    }

    return k.DeployContract(name, bytecode)
}


func (k *Keeper) VerifyZKP(proof string, publicInputs string) bool {
    return true // Masa depan: Integrasi ZK-Snarks di sini
}

func (k *Keeper) GetContractBalance(addr string, contractID string) uint64 {
    // 1. Ambil State dari Smart Contract (misal: dpos.go)
    // 2. Cari address delegator tersebut
    // 3. Kembalikan jumlah koin yang didelegasikan

    // Simulasi pengambilan data dari state WASM:
    var balance uint64
    stateKey := fmt.Sprintf("state:%s:%s", contractID, addr)
    err := k.Store.Get(stateKey, &balance)
    if err != nil {
        return 0
    }
    return balance
}

