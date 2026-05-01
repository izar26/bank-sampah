-- Active: 1777610373544@@127.0.0.1@5432@bank_sampah
-- =======================================================
-- RESET DATA BANK SAMPAH BACKEND (PostgreSQL)
-- =======================================================

-- Menggunakan CASCADE untuk mengabaikan batasan Foreign Key selama penghapusan

-- Hapus seluruh detail nasabah yang dicairkan
TRUNCATE TABLE si_items CASCADE;

-- Hapus seluruh dokumen Surat Instruksi
TRUNCATE TABLE si_documents CASCADE;

-- Hapus seluruh antrean callback ke SIMAK
TRUNCATE TABLE callback_queue CASCADE;

-- Hapus seluruh riwayat aktivitas (Audit Log)
TRUNCATE TABLE audit_logs CASCADE;

-- Hapus data sekolah jika Anda ingin mendaftar ulang di Dashboard Bank
TRUNCATE TABLE schools CASCADE;

-- Catatan:
-- Tabel admins TIDAK di-truncate agar Anda tetap bisa login.
