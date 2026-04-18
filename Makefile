# --- 👑 BVM COMMAND CENTER (BVM EDITION V2) ---
BINARY_NAME=bvm
# Jalur ke source utama Sultan
CLI_SOURCES=./cmd/bvm/main.go ./cmd/bvm/node.go
WASM_OUT=contracts/test_contract/contract.wasm
WASM_SRC=contracts/test_contract/main.go

.PHONY: setup build start miner check send stats search mempool clean build_wasm super

# --- 1. SETUP & BUILD ---
tidy:
	@echo "🧹 Merapikan Folder Tools..."
	@mkdir -p tools
	@if [ -f compiler.py ]; then mv compiler.py tools/; fi
	@go mod tidy
	@echo "✅ Folder bersih dan rapi!"

build_wasm:
	@echo "🏗️  Membangun Smart Contract WASM..."
	@mkdir -p build
	GOOS=wasip1 GOARCH=wasm go build -o build/dpos.wasm ./x/wasm/contracts/dpos.go
	GOOS=wasip1 GOARCH=wasm go build -o build/test_contract.wasm ./contracts/test_contract/main.go
	@echo "🪄  Menstabilkan biner dengan Compiler Sultan..."
	@python3 tools/compiler.py build/dpos.wasm
	@python3 tools/compiler.py build/test_contract.wasm
	@echo "✅ Semua Kontrak Teroptimasi!"

build:
	@echo "🏗️  Membangun Otak Pusat (CLI)..."
	@go build -o $(BINARY_NAME) $(CLI_SOURCES)
	@echo "✅ File './bvm' berhasil dikompilasi!"

# --- 2. RUNNING NODE ---
start: build
	@echo "🚀 Memulai Kernel BVM Sultan..."
	@./$(BINARY_NAME) node start

# --- 3. MINING & MEMPOOL ---
miner:
	@echo "⛏️  Memulai Penambangan BVM..."
	@go run ./bvm-external-miner/main.go

mempool:
	@echo "📦 Memeriksa Antrean Transaksi..."
	@./$(BINARY_NAME) mempool

# --- 4. WALLET & USER SEARCH ---
check:
	@echo "🔍 Mengecek Saldo Wallet..."
	@./$(BINARY_NAME) wallet balance

search:
	@echo "🔍 Mencari User: $(addr)"
	@./$(BINARY_NAME) wallet search $(addr)

send:
	@echo "💸 Mengirim BVM..."
	@./$(BINARY_NAME) wallet send --to $(to) --amount $(amount)

# --- 5. CLEANING ---
clean:
	@echo "🧹 Membersihkan Sampah..."
	@rm -f $(BINARY_NAME)
	@rm -rf ./data/test_db
	@echo "✨ Workspace Bersih!"

super: clean setup build build_wasm
