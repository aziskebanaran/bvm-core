#!/bin/bash

# Pastikan kita di folder bvm-core
cd ~/bvm-core

echo "🚀 Memulai sinkronisasi BVM Core ke Awan..."

# 1. Menambahkan semua perubahan
git add .

# 2. Meminta pesan commit dari Sultan
echo "📝 Apa pesan untuk versi ini, Jenderal?"
read message

# 3. Commit dan Push
git commit -m "$message"
git push origin main

echo "✅ MISI SELESAI! Kode sudah aman di GitHub."
