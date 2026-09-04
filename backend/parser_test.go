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

// TestParseATMLogToCSV_SeqNrFromNoRekord menguji fix bug: anchor nomor urut
// transaksi dulu cuma "TRAN SEQ NR [...]" yang CUMA ada di EJ mesin DN200V.
// Mesin Hyosung/Hitachi/OKI nol kemunculan field itu, jadi SeqNr selalu kosong
// dan flush() nggak pernah nulis baris - hasilnya count=0 dan LoadEJ nolak
// filenya ("tidak ada transaksi valid"). Struk cetak "NO. REKORD <n>" ada di
// SEMUA merek dan nilainya sama persis dengan rec_num yang dipakai buat join ke
// RC, jadi itu yang dipakai sebagai anchor universal. Format strukmya beda
// tipis antar merek: ada yang polos ("NO. REKORD 8783") ada yang pakai titik
// dua ("NO. REKORD   : 5822").
func TestParseATMLogToCSV_SeqNrFromNoRekord(t *testing.T) {
	// Cuplikan asli EJ Hyosung S1CBPNR030 (rec 8786) dan Hitachi S1GSMDR001
	// (varian pakai titik dua) - dipendekin, format barisnya dipertahankan.
	raw := "17/08/2026 03:10:00 TRANSACTION START\n" +
		"17/08/2026 03:10:01 Terminal ID    [S1CBPNR030]\n" +
		"17/08/2026 03:10:01 Card Inserted\n" +
		"17/08/2026 03:10:01 TRANSACTION START\n" +
		"17/08/2026 03:10:01 Card Number 537176XXXXXX7766\n" +
		"17/08/2026 03:10:20 OP Code : ADBBA  A\n" +
		"17/08/2026 03:10:20 Amount : 250000\n" +
		"17/08/2026 03:10:21 TRANSACTION REPLIED\n" +
		"[Transaction record]\n" +
		"Trans SEQ Number [8786]\n" +
		"    537176******7766 \n" +
		"    NO. REKORD 8786\n" +
		"17/08/2026 03:10:40 TRANSACTION END\n" +
		"08/20/2026 17:30:12 TRANSACTION START\n" +
		"08/20/2026 17:30:12 Terminal ID    [S1GSMDR001]\n" +
		"08/20/2026 17:30:12 Card Number: 519893******7403\n" +
		"08/20/2026 17:30:12 Amount : 000000780012\n" +
		"08/20/2026 17:30:14 TRANSACTION REPLIED\n" +
		"NO. REKORD   : 6388\n" +
		"08/20/2026 17:30:30 TRANSACTION END\n"

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
		t.Fatalf("count = %d, want 2 (EJ tanpa TRAN SEQ NR harus tetap keparse lewat NO. REKORD)", count)
	}

	rows := readCSV(t, outPath)
	if got := rows[1][4]; got != "8786" { // kolom seq_nr
		t.Errorf("seq_nr baris 1 = %q, want 8786 (NO. REKORD tanpa titik dua)", got)
	}
	if got := rows[2][4]; got != "6388" {
		t.Errorf("seq_nr baris 2 = %q, want 6388 (NO. REKORD pakai titik dua)", got)
	}
}

// TestParseATMLogToCSV_CardAndTerminalFormatVariants menguji varian format
// field di 3 merek baru: Hyosung/Hitachi/OKI nulis "Card Number" (bukan
// all-caps) dengan/ tanpa titik dua dan mask-nya bisa pakai 'X' bukan '*',
// sementara OKI nulis "Terminal ID : S1BGBRR006" tanpa kurung siku. Baris
// "Card Number    [5198930898695126]" di blok [Transaction record] OKI berisi
// PAN UTUH tanpa mask - itu TIDAK boleh ikut kepungut ke kolom card_masked.
func TestParseATMLogToCSV_CardAndTerminalFormatVariants(t *testing.T) {
	// Cuplikan asli EJ OKI S1BGBRR006 (rec 3982) + Hyosung (mask 'X').
	raw := "17/08/2026 01:21:22 TRANSACTION START\n" +
		"17/08/2026 01:21:22 Terminal ID : S1BGBRR006\n" +
		"17/08/2026 01:21:32 Card Number : **********695126\n" +
		"17/08/2026 01:21:34 Amount : 450000\n" +
		"17/08/2026 01:21:37 [Transaction record]\n" +
		"17/08/2026 01:21:37 Card Number    [5198930898695126]\n" +
		"    NO. REKORD 3982\n" +
		"17/08/2026 01:22:33 TRANSACTION END\n" +
		"17/08/2026 02:40:01 TRANSACTION START\n" +
		"17/08/2026 02:40:01 Terminal ID    [S1CBPNR030]\n" +
		"17/08/2026 02:40:01 Card Number 537176XXXXXX7766\n" +
		"17/08/2026 02:40:59 Amount : 250000\n" +
		"    NO. REKORD 8786\n" +
		"17/08/2026 02:41:20 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	if _, err := ParseATMLogToCSV(inPath, outPath); err != nil {
		t.Fatal(err)
	}
	rows := readCSV(t, outPath)

	if got := rows[1][1]; got != "**********695126" { // kolom card_masked
		t.Errorf("card OKI = %q, want **********695126 (versi ber-mask, bukan PAN utuh)", got)
	}
	if got := rows[1][2]; got != "S1BGBRR006" { // kolom terminal_id
		t.Errorf("terminal OKI = %q, want S1BGBRR006 (format titik dua tanpa kurung siku)", got)
	}
	if got := rows[2][1]; got != "537176XXXXXX7766" {
		t.Errorf("card Hyosung = %q, want 537176XXXXXX7766 (mask pakai X)", got)
	}
}

// TestParseATMLogToCSV_DepositAmountImpliedDecimals menguji fix nominal setor
// tunai. Di Hitachi & Hyosung, baris "Amount :" pas TRANSACTION REQUESTING buat
// transaksi setor (OP Code BB) dikirim apa adanya dari field ISO-8583 yang
// pakai 2 desimal implisit, jadi nilainya 100x nominal asli (bukti: EJ Hitachi
// rec 5830 "Amount : 30000000" vs RC 300000; Hyosung rec 8783 "Amount :
// 195000000" vs RC 1950000). DN200V & OKI nggak begitu - nominalnya 1:1.
// Daripada nebak per merek, faktor 100-nya dibuktiin sendiri sama mesinnya:
// hasil hitungan uang fisik "Total Amount IDR <n>" di blok yang sama harus
// persis Amount/100. Kalau nggak persis, nominal dibiarkan apa adanya.
func TestParseATMLogToCSV_DepositAmountImpliedDecimals(t *testing.T) {
	// Cuplikan asli EJ Hyosung S1CBPNR030 rec 8783 (setor Rp1.950.000) dan
	// DN200V S1DTRBR013 rec 2727 (setor Rp4.850.000, sudah 1:1).
	raw := "17/08/2026 02:40:00 TRANSACTION START\n" +
		"17/08/2026 02:40:01 Card Number 537176XXXXXX7766\n" +
		"17/08/2026 02:40:47 Notes Counted:\n" +
		"IDR50000x7\n" +
		"IDR100000x16\n" +
		"Rejectx3\n" +
		"Total Amount IDR 1950000\n" +
		"17/08/2026 02:40:59 OP Code : BB     A\n" +
		"17/08/2026 02:40:59 Amount : 195000000\n" +
		"17/08/2026 02:41:01 TRANSACTION REPLIED\n" +
		"    NO. REKORD 8783\n" +
		"    DEPOSIT                 RP1.950.000  \n" +
		"17/08/2026 02:41:20 TRANSACTION END\n" +
		"30/06/2026 08:58:00 TRANSACTION START\n" +
		"30/06/2026 08:58:00 CARD NUMBER 519893******7283\n" +
		"30/06/2026 08:58:30 TOTAL AMOUNT IDR 4850000\n" +
		"30/06/2026 08:58:47 OP Code : BB     A\n" +
		"30/06/2026 08:58:47 Amount : 4850000\n" +
		"30/06/2026 08:58:48 TRANSACTION REPLIED\n" +
		"30/06/2026 08:58:48 TRAN SEQ NR [2727]\n" +
		"    NO. REKORD 2727\n" +
		"30/06/2026 08:59:20 TRANSACTION END\n" +
		// Setor yang nominalnya NGGAK bisa dibuktiin hitungan uang fisik
		// (nggak ada baris Total Amount IDR) - biarkan apa adanya, jangan nebak.
		"19/08/2026 00:03:35 TRANSACTION START\n" +
		"19/08/2026 00:03:35 Card Number: 194634******7344\n" +
		"19/08/2026 00:03:35 OP Code : BB     A\n" +
		"19/08/2026 00:03:35 Amount : 000470000000\n" +
		"19/08/2026 00:03:37 TRANSACTION REPLIED\n" +
		"    NO. REKORD 5819\n" +
		"19/08/2026 00:03:58 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	if _, err := ParseATMLogToCSV(inPath, outPath); err != nil {
		t.Fatal(err)
	}
	rows := readCSV(t, outPath)

	if got := rows[1][3]; got != "1950000" { // kolom amount
		t.Errorf("amount Hyosung = %q, want 1950000 (Amount:/100 dibuktikan Total Amount IDR)", got)
	}
	if got := rows[2][3]; got != "4850000" {
		t.Errorf("amount DN200V = %q, want 4850000 (sudah 1:1, jangan diutak-atik)", got)
	}
	if got := rows[3][3]; got != "000470000000" {
		t.Errorf("amount tanpa bukti hitungan = %q, want 000470000000 (biarkan apa adanya)", got)
	}
}

// readCSV baca hasil parser jadi slice baris (baris 0 = header).
func readCSV(t *testing.T, path string) [][]string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	rows, err := csv.NewReader(f).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	return rows
}

// TestParseATMLogToCSV_SeqNrZeroPaddingNormalized menguji kejadian nyata di EJ
// DN200V S1GBMSR055 rec 660: satu transaksi yang di-retry ke host ke-log dua
// kali dengan seq sama. Copy pertama nyebut seq di DUA tempat dengan format
// beda - "TRAN SEQ NR [0660]" (nol-padding) dan struk " NO. REKORD  660"
// (tanpa padding) - sedangkan copy kedua cuma nyebut yang ber-padding. Kalau
// seq disimpan apa adanya, dua copy itu punya dedup key beda dan transaksi yang
// sama kehitung dobel. Angkanya sama, jadi nol depannya dibuang.
func TestParseATMLogToCSV_SeqNrZeroPaddingNormalized(t *testing.T) {
	raw := "09/07/2026 15:14:03 TRANSACTION START\n" +
		"09/07/2026 15:14:03 CARD NUMBER 524559******7776\n" +
		"09/07/2026 15:15:29 Amount : 600000\n" +
		"09/07/2026 15:15:32 TRANSACTION REPLIED\n" +
		"09/07/2026 15:15:32 TRAN SEQ NR [0660]\n" +
		" NO. REKORD  660\n" +
		"09/07/2026 15:16:46 Amount : 600000\n" + // retry ke host, seq sama
		"09/07/2026 15:16:47 TRAN SEQ NR [0660]\n" +
		"09/07/2026 15:17:00 TRANSACTION END\n"

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
		t.Fatalf("count = %d, want 1 (retry seq sama gak boleh kehitung dobel)", count)
	}
	if got := readCSV(t, outPath)[1][4]; got != "660" {
		t.Errorf("seq_nr = %q, want 660 (nol depan dibuang biar satu format)", got)
	}
}

// TestParseATMLogToCSV_TwoDepositsOneSessionScaledIndependently menguji urutan
// baris yang bikin normalisasi nominal gampang salah: hasil hitungan uang
// ("Total Amount IDR") selalu muncul SEBELUM baris "Amount :" transaksi yang
// sama. Kalau satu sesi kartu isinya dua setoran (kejadian nyata di EJ Hitachi
// S1GSMDR001 rec 5819 & 5820), pas setoran pertama di-flush, hitungan uang yang
// kesimpen udah kegantikan punya setoran KEDUA - jadi normalisasi harus dikunci
// pas baris "Amount :" dibaca, bukan pas flush.
func TestParseATMLogToCSV_TwoDepositsOneSessionScaledIndependently(t *testing.T) {
	// Cuplikan asli EJ Hitachi S1GSMDR001 (rec 5819 Rp4.700.000, rec 5820
	// Rp1.400.000) - dipendekin, urutan barisnya dipertahankan.
	raw := "08/19/2026 00:02:43 TRANSACTION START\n" +
		"08/19/2026 00:02:44 Card Number: 194634******7344\n" +
		"08/19/2026 00:03:21 Notes Counted:\n" +
		"IDR100000x47\n" +
		"Rejectx3\n" +
		"Total Amount IDR 4700000\n" +
		"08/19/2026 00:03:35 OP Code : BB     A\n" +
		"08/19/2026 00:03:35 Amount : 000470000000\n" +
		"08/19/2026 00:03:37 TRANSACTION REPLIED\n" +
		"    NO. REKORD 5819\n" +
		"08/19/2026 00:04:28 Notes Counted:\n" +
		"IDR100000x14\n" +
		"Total Amount IDR 1400000\n" +
		"08/19/2026 00:04:45 OP Code : BB     A\n" +
		"08/19/2026 00:04:45 Amount : 000140000000\n" +
		"08/19/2026 00:04:48 TRANSACTION REPLIED\n" +
		"    NO. REKORD 5820\n" +
		"08/19/2026 00:05:10 TRANSACTION END\n"

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
	rows := readCSV(t, outPath)
	if got := rows[1][3]; got != "4700000" {
		t.Errorf("amount rec 5819 = %q, want 4700000", got)
	}
	if got := rows[2][3]; got != "1400000" {
		t.Errorf("amount rec 5820 = %q, want 1400000", got)
	}
}

// TestParseATMLogToCSV_DispensedTotalNeverRescales mengunci scope normalisasi
// nominal: yang boleh jadi bukti faktor 100 CUMA hitungan uang MASUK ("Notes
// Counted", setor tunai) - hitungan uang KELUAR ("NOTES DISPENSED", penarikan)
// nggak boleh. Kejadian nyata yang bikin ini penting: EJ DN200V S1DTRBR013 rec
// 3285 punya satu sesi kartu berisi permintaan Rp5.000.000 DAN pengeluaran uang
// Rp50.000 - persis 100x. Kalau hitungan uang keluar dipakai sebagai bukti,
// selisih Rp4.950.000 yang beneran bakal ketutup jadi "match".
func TestParseATMLogToCSV_DispensedTotalNeverRescales(t *testing.T) {
	raw := "01/07/2026 07:06:00 TRANSACTION START\n" +
		"01/07/2026 07:06:00 CARD NUMBER 537176******6807\n" +
		"01/07/2026 07:07:12 NOTES DISPENSED:\n" +
		"IDR50000*1\n" +
		"TOTAL AMOUNT IDR 50000\n" +
		"01/07/2026 07:07:30 Amount : 5000000\n" +
		"01/07/2026 07:07:32 TRANSACTION REPLIED\n" +
		"    NO. REKORD 3285\n" +
		"01/07/2026 07:08:00 TRANSACTION END\n"

	dir := t.TempDir()
	inPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(inPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	outPath := filepath.Join(dir, "out.csv")

	if _, err := ParseATMLogToCSV(inPath, outPath); err != nil {
		t.Fatal(err)
	}
	if got := readCSV(t, outPath)[1][3]; got != "5000000" {
		t.Errorf("amount = %q, want 5000000 (hitungan uang KELUAR bukan bukti skala setor)", got)
	}
}
