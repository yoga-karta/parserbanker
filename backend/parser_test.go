package main

import (
	"encoding/csv"
	"os"
	"path/filepath"
	"testing"
)

// TestParseATMLogToCSV_DedupExactDuplicateBlock menguji fix buat file EJ mentah
// yang punya section ke-duplikat persis (kejadian nyata di EJ (1).TXT, section
// ~5000 baris ke-copy dua kali gara-gara overlap export) - blok dengan
// timestamp+seq+amount identik cuma boleh dihitung sekali.
func TestParseATMLogToCSV_DedupExactDuplicateBlock(t *testing.T) {
	block := "17/06/2026 10:51:26 TRANSACTION START\n" +
		"17/06/2026 10:51:26 CARD NUMBER 532659******4565\n" +
		"17/06/2026 10:51:47 Amount : 1000000\n" +
		"17/06/2026 10:51:50 TRANSACTION REPLIED\n" +
		"17/06/2026 10:51:50 TRAN SEQ NR [0000]\n"
	raw := block + block // section ke-duplikat persis

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	count, err := ParseATMLogToCSV(inPath, outPath)
	if err != nil {
		t.Fatal(err)
	}
	if count != 1 {
		t.Errorf("count = %d, want 1 (blok duplikat harusnya cuma dihitung sekali)", count)
	}

	// Transaksi beda (seq beda) tetap harus kehitung normal, bukan ikut kefilter.
	rawTwoDifferent := block + "17/06/2026 10:52:00 TRANSACTION START\n" +
		"17/06/2026 10:52:00 CARD NUMBER 111111\n" +
		"17/06/2026 10:52:10 Amount : 500000\n" +
		"17/06/2026 10:52:15 TRANSACTION REPLIED\n" +
		"17/06/2026 10:52:15 TRAN SEQ NR [0001]\n"
	inPath2 := filepath.Join(dir, "ej_raw2.txt")
	os.WriteFile(inPath2, []byte(rawTwoDifferent), 0644)
	outPath2 := filepath.Join(dir, "out2.csv")
	count2, err := ParseATMLogToCSV(inPath2, outPath2)
	if err != nil {
		t.Fatal(err)
	}
	if count2 != 2 {
		t.Errorf("count2 = %d, want 2 (2 transaksi beneran beda gak boleh ke-dedup)", count2)
	}
}

// TestParseATMLogToCSV_RollbackAfterEndNotAttributedToPrevTransaction menguji
// fix bug nyata: "Rollback Notes"/"Rollback OK" yang muncul di antara
// TRANSACTION END satu transaksi dan TRANSACTION START transaksi berikutnya
// (mis. ATM narik balik uang yang nggak diambil nasabah) dulu kebaca sebagai
// bagian transaksi SEBELUMNYA yang udah kelar bersih, jadi salah ditandai
// ROLLBACK padahal blok transaksi itu sendiri nggak nyebut "Rollback" sama
// sekali.
func TestParseATMLogToCSV_RollbackAfterEndNotAttributedToPrevTransaction(t *testing.T) {
	raw := "17/06/2026 10:51:26 TRANSACTION START\n" +
		"17/06/2026 10:51:26 CARD NUMBER 532659******4565\n" +
		"17/06/2026 10:51:47 Amount : 1000000\n" +
		"17/06/2026 10:51:50 TRANSACTION REPLIED\n" +
		"17/06/2026 10:51:50 TRAN SEQ NR [0000]\n" +
		"17/06/2026 10:51:55 TRANSACTION END\n" +
		"17/06/2026 10:51:58 ----- Rollback Notes ----\n" + // milik proses mesin, bukan transaksi 0000
		"17/06/2026 10:52:03   Rollback OK\n" +
		"17/06/2026 10:52:10 TRANSACTION START\n" +
		"17/06/2026 10:52:10 CARD NUMBER 111111\n" +
		"17/06/2026 10:52:15 Amount : 500000\n" +
		"17/06/2026 10:52:20 TRANSACTION REPLIED\n" +
		"17/06/2026 10:52:20 TRAN SEQ NR [0001]\n" +
		"17/06/2026 10:52:25 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	count, err := ParseATMLogToCSV(inPath, outPath)
	if err != nil {
		t.Fatal(err)
	}
	if count != 2 {
		t.Fatalf("count = %d, want 2", count)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	// rows[0] header, rows[1] seq 0000, rows[2] seq 0001. Kolom terakhir = status.
	if got := rows[1][len(rows[1])-1]; got != "SUCCESS" {
		t.Errorf("seq 0000 status = %q, want SUCCESS (Rollback Notes sesudah END-nya bukan miliknya)", got)
	}
}

// TestParseATMLogToCSV_RollbackNotesWithoutOKIsNotRollback menguji keputusan
// bisnis (revisi 1 Sep 2026): indikator ROLLBACK harus persis "Rollback OK".
// Kejadian nyata di EJ.TXT: satu transaksi kadang punya 2 baris "----- Rollback
// Notes ----", tapi cuma rollback yang PERTAMA yang beneran diikuti "Rollback
// OK" - baris "Rollback Notes" kedua cuma housekeeping penutupan transaksi
// (mis. abis notes yang ditolak diambil dari output tray) dan TIDAK diikuti
// "OK". Transaksi kayak gini tetap harus ROLLBACK (karena OK yang pertama ada),
// tapi transaksi yang "Rollback Notes"-nya nggak PERNAH diikuti "OK" sama
// sekali harus tetap status aslinya (SUCCESS), bukan ROLLBACK.
func TestParseATMLogToCSV_RollbackNotesWithoutOKIsNotRollback(t *testing.T) {
	raw := "09/07/2026 16:46:19 TRANSACTION START\n" +
		"09/07/2026 16:46:19 CARD NUMBER 194634******6969\n" +
		"09/07/2026 16:47:03 Amount : 2400000\n" +
		"09/07/2026 16:47:05 TRANSACTION REPLIED\n" +
		"09/07/2026 16:47:05 TRAN SEQ NR [0694]\n" +
		"09/07/2026 16:47:06 ----- Rollback Notes ----\n" + // rollback #1, beneran
		"09/07/2026 16:47:16   Rollback OK\n" +
		"09/07/2026 16:47:24   Refused Notes in Output Tray TAKEN\n" +
		"09/07/2026 16:47:25 ----- Rollback Notes ----\n" + // housekeeping, TANPA OK
		"09/07/2026 16:47:28 TRANSACTION END\n" +
		"09/07/2026 16:48:01 TRANSACTION START\n" +
		"09/07/2026 16:48:01 CARD NUMBER 111111\n" +
		"09/07/2026 16:48:10 Amount : 100000\n" +
		"09/07/2026 16:48:15 TRANSACTION REPLIED\n" +
		"09/07/2026 16:48:15 TRAN SEQ NR [0695]\n" +
		"09/07/2026 16:48:16 ----- Rollback Notes ----\n" + // rollback disebut tapi NGGAK PERNAH ada OK
		"09/07/2026 16:48:20 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	if _, err := ParseATMLogToCSV(inPath, outPath); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if got := rows[1][len(rows[1])-1]; got != "ROLLBACK OK" {
		t.Errorf("seq 0694 status = %q, want ROLLBACK OK (ada Rollback OK beneran di blok ini)", got)
	}
	if got := rows[2][len(rows[2])-1]; got != "SUCCESS" {
		t.Errorf("seq 0695 status = %q, want SUCCESS (Rollback Notes tanpa OK bukan rollback beneran)", got)
	}
}

// TestParseATMLogToCSV_RollbackNotesSuccessfullyAndShutterOpened menguji 2 EJ
// Status tambahan (keputusan 3 Sep 2026) yang diperlakukan sama seperti
// "Rollback OK" buat kategorisasi Selisih Kurang (lihat diff.go), tapi
// disimpan apa adanya di kolom status (bukan dinormalisasi ke satu label).
func TestParseATMLogToCSV_RollbackNotesSuccessfullyAndShutterOpened(t *testing.T) {
	raw := "17/06/2026 10:00:00 TRANSACTION START\n" +
		"17/06/2026 10:00:00 CARD NUMBER 111111\n" +
		"17/06/2026 10:00:05 Amount : 100000\n" +
		"17/06/2026 10:00:10 TRANSACTION REPLIED\n" +
		"17/06/2026 10:00:10 TRAN SEQ NR [0001]\n" +
		"17/06/2026 10:00:15   Rollback Notes Successfully\n" +
		"17/06/2026 10:00:20 TRANSACTION END\n" +
		"17/06/2026 10:01:00 TRANSACTION START\n" +
		"17/06/2026 10:01:00 CARD NUMBER 222222\n" +
		"17/06/2026 10:01:05 Amount : 200000\n" +
		"17/06/2026 10:01:10 TRANSACTION REPLIED\n" +
		"17/06/2026 10:01:10 TRAN SEQ NR [0002]\n" +
		"17/06/2026 10:01:15   Shutter Opened for notes removal\n" +
		"17/06/2026 10:01:20 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	if _, err := ParseATMLogToCSV(inPath, outPath); err != nil {
		t.Fatal(err)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	if got := rows[1][len(rows[1])-1]; got != "ROLLBACK NOTES SUCCESSFULLY" {
		t.Errorf("seq 0001 status = %q, want ROLLBACK NOTES SUCCESSFULLY", got)
	}
	if got := rows[2][len(rows[2])-1]; got != "SHUTTER OPENED FOR NOTES REMOVAL" {
		t.Errorf("seq 0002 status = %q, want SHUTTER OPENED FOR NOTES REMOVAL", got)
	}
}

// TestParseATMLogToCSV_SecondAttemptRollbackNotAttributedToFirst menguji bug
// nyata rec_num 407: satu blok TRANSACTION START..END bisa punya LEBIH DARI
// SATU percobaan. Deposit pertama (seq 0407, Rp3.800.000) sukses dapat TRAN
// SEQ NR sendiri. Nasabah lanjut coba deposit lagi di sesi kartu yang sama -
// percobaan KEDUA ini gagal di PING ERROR (nggak pernah dapat TRAN SEQ NR
// sendiri) terus di-rollback. Rollback punya percobaan kedua ini dulu salah
// numpuk ke record seq 0407 (yang beneran sukses & nggak ada sangkut paut).
func TestParseATMLogToCSV_SecondAttemptRollbackNotAttributedToFirst(t *testing.T) {
	raw := "25/07/2026 23:12:14 TRANSACTION START\n" +
		"25/07/2026 23:12:14 CARD NUMBER 519893******0668\n" +
		"25/07/2026 23:12:19 PIN ENTERED\n" +
		"25/07/2026 23:13:05 Amount : 3800000\n" +
		"25/07/2026 23:13:08 TRANSACTION REPLIED\n" +
		"25/07/2026 23:13:08 TRAN SEQ NR [0407]\n" +
		"25/07/2026 23:13:27 ----- Deposit Cash & Print ----\n" +
		"25/07/2026 23:13:31 PIN ENTERED\n" + // percobaan deposit KEDUA, sesi kartu sama - masuk PIN lagi
		"25/07/2026 23:14:27 EMV AID A0000006021010 / 519893******0668 STARTED\n" +
		"25/07/2026 23:14:27 ***** Tran Request State *****\n" +
		"25/07/2026 23:14:39 PING ERROR\n" + // gagal SEBELUM sempat ada baris Amount/TRAN SEQ NR sendiri
		"25/07/2026 23:14:39 ----- Rollback Notes ----\n" +
		"25/07/2026 23:14:59   Rollback OK\n" +
		"25/07/2026 23:15:18 CARD(519893******0668) TAKEN\n" +
		"25/07/2026 23:15:18 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	count, err := ParseATMLogToCSV(inPath, outPath)
	if err != nil {
		t.Fatal(err)
	}
	// Percobaan kedua nggak pernah dapat TRAN SEQ NR (gagal PING sebelum
	// TRANSACTION REQUESTING) - nggak ada yang bisa direkonsiliasi buat itu,
	// jadi cuma 1 baris (seq 0407) yang harusnya kehasil.
	if count != 1 {
		t.Fatalf("count = %d, want 1 (percobaan kedua nggak punya TRAN SEQ NR, nggak boleh ke-flush)", count)
	}

	f, err := os.Open(outPath)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}

	got := rows[1]
	if got[len(got)-1] != "SUCCESS" {
		t.Errorf("seq 0407 status = %q, want SUCCESS (rollback punya percobaan kedua, bukan punya 0407)", got[len(got)-1])
	}
	if got[3] != "3800000" { // kolom amount
		t.Errorf("seq 0407 amount = %q, want 3800000 (bukan ketiban amount percobaan kedua)", got[3])
	}
}
