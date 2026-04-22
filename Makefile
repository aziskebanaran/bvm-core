# --- 👑 BVM COMMAND CENTER (BVM EDITION V2) ---
BINARY_NAME=bvm
# Jalur ke source utama Sultan
CLI_SOURCES=./cmd/bvm/main.go ./cmd/bvm/node.go
WASM_OUT=contracts/test_contract/contract.wasm
WASM_SRC=contracts/test_contract/main.go

.PHONY: setup build start miner check send stats search mempool clean build_wasm super install build-vps build-android

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

build-vps:
	@echo "🖥️  Membangun BVM untuk VPS (Linux x64)..."
	GOOS=linux GOARCH=amd64 go build -o $(BINARY_NAME)-vps $(CLI_SOURCES)
	@echo "✅ File siap: $(BINARY_NAME)-vps"

build-android:
	@echo "📱 Membangun BVM untuk HP Android lain (ARM64)..."
	GOOS=linux GOARCH=arm64 go build -o $(BINARY_NAME)-android $(CLI_SOURCES)
	@echo "✅ File siap: $(BINARY_NAME)-android"

# --- Variabel Target (Bisa diganti saat menjalankan perintah) ---
TARGET_DIR ?= ~/bvm-nexus

# --- 3. AUTO-INSTALLER UNIVERSAL ---
install: super
	@echo "📦 Menginstall BVM ke target: $(TARGET_DIR)..."
	@mkdir -p $(TARGET_DIR)/data/blockchain_db
	@mkdir -p $(TARGET_DIR)/data/apps_storage
	@cp $(BINARY_NAME) $(TARGET_DIR)/$(BINARY_NAME)
	@if [ ! -f $(TARGET_DIR)/node_wallet.json ]; then \
		cp node_wallet.json $(TARGET_DIR)/node_wallet.json; \
	fi
	@echo "🎯 Selesai! Target di $(TARGET_DIR) sudah diperbarui."

build-wallet:
	@echo "🏗️  Membangun Biner Wallet..."
	@mkdir -p $(BUILD_DIR)
	go build -o $(BUILD_DIR)/bvm-wallet wallet/main.go
	@echo "✅ Biner selesai: $(BUILD_DIR)/bvm-wallet"

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
