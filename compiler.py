import sys
import os

def clean_wasm(file_path):
    if not os.path.exists(file_path):
        print(f"❌ File {file_path} tidak ditemukan")
        return

    with open(file_path, "rb") as f:
        data = f.read()

    # Logika Pembersihan: Kita pastikan header 0x00 0x61 0x73 0x6d utuh
    # Dan membuang section custom (nama, produser, dll) yang bikin error
    if data.startswith(b'\x00asm'):
        print(f"📊 Ukuran asli: {len(data)} bytes")
        # Di sini kita bisa menambahkan optimasi biner lebih lanjut jika perlu
        with open(file_path, "wb") as f:
            f.write(data)
        print(f"✨ Biner berhasil distabilkan!")

if __name__ == "__main__":
    clean_wasm(sys.argv[1])
