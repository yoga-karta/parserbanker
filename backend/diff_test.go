package main

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

// TestRunDiff_SelisihKurangRule menguji aturan kategorisasi per 1 Sep 2026:
// "Selisih Kurang" cuma buat baris ROLLBACK yang nominal EJ=Cash sama, tidak
// ada lagi kategori "Selisih Lebih", dan baris nominal beda (rollback atau
// bukan) jatuh ke "Tidak Ditemukan".
func TestRunDiff_SelisihKurangRule(t *testing.T) {
	dir := t.TempDir()

	ejPath := filepath.Join(dir, "ej_normalized.csv")
	ejCSV := "timestamp,card_masked,terminal_id,amount,seq_nr,status\n" +
		"01/09/2026 10:00:00,111,T1,100000,1,SUCCESS\n" + // match (nominal sama, bukan rollback)
		"01/09/2026 10:01:00,222,T1,200000,2,ROLLBACK OK\n" + // selisih_kurang (rollback, nominal sama)
		"01/09/2026 10:02:00,333,T1,300000,3,ROLLBACK OK\n" + // dulu otomatis selisih_kurang, sekarang tidak_ditemukan (rollback tapi nominal beda)
		"01/09/2026 10:03:00,444,T1,400000,4,SUCCESS\n" + // dulu selisih_lebih/selisih_kurang, sekarang tidak_ditemukan (nominal beda, bukan rollback)
		"01/09/2026 10:04:00,555,T1,500000,5,SUCCESS\n" // tidak_ditemukan (tidak ada pasangan RC)
	if err := os.WriteFile(ejPath, []byte(ejCSV), 0644); err != nil {
		t.Fatal(err)
	}

	rcPath := filepath.Join(dir, "rc.txt")
	rcTXT := "111/BNI;S1/1/1;0210/S1/111/1/VB;100000;D;1000000;01/09/26\n" +
		"222/BNI;S1/2/1;0210/S1/222/1/VB;200000;D;1000000;01/09/26\n" +
		"333/BNI;S1/3/1;0210/S1/333/1/VB;350000;D;1000000;01/09/26\n" +
		"444/BNI;S1/4/1;0210/S1/444/1/VB;450000;D;1000000;01/09/26\n"
	if err := os.WriteFile(rcPath, []byte(rcTXT), 0644); err != nil {
		t.Fatal(err)
	}

	resultDBPath := filepath.Join(dir, "result.duckdb")
	s, err := RunDiff(context.Background(), ejPath, rcPath, resultDBPath, LoadOptions{SkipHeader: false})
	if err != nil {
		t.Fatalf("RunDiff error: %v", err)
	}

	if s.Match != 1 {
		t.Errorf("Match = %d, want 1", s.Match)
	}
	if s.SelisihKurang != 1 {
		t.Errorf("SelisihKurang = %d, want 1 (cuma rollback+nominal sama)", s.SelisihKurang)
	}
	if s.TidakDitemukan != 3 {
		t.Errorf("TidakDitemukan = %d, want 3 (rollback-nominal-beda + normal-nominal-beda + tanpa pasangan)", s.TidakDitemukan)
	}
	if s.SelisihNominalKurang != 200000 {
		t.Errorf("SelisihNominalKurang = %d, want 200000 (nominal asli baris selisih_kurang seq 2, bukan selisih EJ-Cash yang selalu 0)", s.SelisihNominalKurang)
	}
}

// TestRunDiff_SelisihKurangThreeStatuses menguji perluasan aturan (3 Sep 2026):
// "Selisih Kurang" berlaku buat 3 EJ Status - "Rollback OK", "Rollback Notes
// Successfully", "Shutter Opened for notes removal" - selama nominal EJ dan
// Cash-nya sama. EJ Status lain (mis. SUCCESS) tetap di-skip dari Selisih
// Kurang walau nominalnya sama - jatuh ke "match" seperti biasa.
func TestRunDiff_SelisihKurangThreeStatuses(t *testing.T) {
	dir := t.TempDir()

	ejPath := filepath.Join(dir, "ej_normalized.csv")
	ejCSV := "timestamp,card_masked,terminal_id,amount,seq_nr,status\n" +
		"01/09/2026 10:00:00,111,T1,100000,1,ROLLBACK OK\n" + // selisih_kurang
		"01/09/2026 10:01:00,222,T1,200000,2,ROLLBACK NOTES SUCCESSFULLY\n" + // selisih_kurang
		"01/09/2026 10:02:00,333,T1,300000,3,SHUTTER OPENED FOR NOTES REMOVAL\n" + // selisih_kurang
		"01/09/2026 10:03:00,444,T1,400000,4,SUCCESS\n" // match (status di luar 3 daftar, tetap di-skip dari selisih_kurang)
	if err := os.WriteFile(ejPath, []byte(ejCSV), 0644); err != nil {
		t.Fatal(err)
	}

	rcPath := filepath.Join(dir, "rc.txt")
	rcTXT := "111/BNI;S1/1/1;0210/S1/111/1/VB;100000;D;1000000;01/09/26\n" +
		"222/BNI;S1/2/1;0210/S1/222/1/VB;200000;D;1000000;01/09/26\n" +
		"333/BNI;S1/3/1;0210/S1/333/1/VB;300000;D;1000000;01/09/26\n" +
		"444/BNI;S1/4/1;0210/S1/444/1/VB;400000;D;1000000;01/09/26\n"
	if err := os.WriteFile(rcPath, []byte(rcTXT), 0644); err != nil {
		t.Fatal(err)
	}

	resultDBPath := filepath.Join(dir, "result.duckdb")
	s, err := RunDiff(context.Background(), ejPath, rcPath, resultDBPath, LoadOptions{SkipHeader: false})
	if err != nil {
		t.Fatalf("RunDiff error: %v", err)
	}

	if s.SelisihKurang != 3 {
		t.Errorf("SelisihKurang = %d, want 3 (ketiga status rollback/shutter, nominal sama)", s.SelisihKurang)
	}
	if s.Match != 1 {
		t.Errorf("Match = %d, want 1 (SUCCESS dengan nominal sama tetap match, bukan selisih_kurang)", s.Match)
	}
}

// TestRunDiff_OKIRejectedDepositEndToEnd nge-tes jalur penuh (parse EJ mentah ->
// join ke RC -> kategorisasi) pakai 2 transaksi OKI S1BGBRR006 yang bedanya
// cuma arah uang: rec 4842 setoran Rp2.500.000 yang DITOLAK host lalu duitnya
// dibalikin (client tegasin ini SATU-SATUNYA Selisih Kurang di mesin itu), dan
// rec 3982 penarikan Rp450.000 yang normal. Dua-duanya nutup dengan kalimat
// "Shutter Opened for notes removal" yang sama persis, jadi cuma arah uangnya
// yang boleh misahin.
//
// Fixture-nya sintetis karena file EJ OKI sample yang kita punya NGGAK punya
// tanggal 19/08/2026 sama sekali (loncat dari 18/08 ke 20/08) - rec 4842 nggak
// ada di sisi EJ, makanya di app kelihatan "Tidak Ditemukan". Logikanya tetap
// harus kebukti bener terlepas dari kelengkapan file sample.
func TestRunDiff_OKIRejectedDepositEndToEnd(t *testing.T) {
	dir := t.TempDir()

	// Cuplikan asli REVISI.docx contoh 2 (rec 4842) + EJ OKI rec 3982.
	raw := "19/08/2026 15:59:37 TRANSACTION START\n" +
		"19/08/2026 15:59:37 Terminal ID : S1BGBRR006\n" +
		"19/08/2026 15:59:53 PIN ENTERED\n" +
		"19/08/2026 15:59:59 Shutter Open -> Insert Cash\n" +
		"19/08/2026 16:00:18 Counted banknote :\n" +
		"IDR100000x25   \n" +
		"Total Amount IDR2500000\n" +
		"19/08/2026 16:00:26 Amount : 2500000\n" +
		"    NO. REKORD 4842\n" +
		"    DEPOSIT       \n" +
		"    TRANSAKSI ANDA DITOLAK\n" +
		"19/08/2026 16:00:36 Shutter Opened for notes removal\n" +
		"19/08/2026 16:00:47 Banknote returned : Succeeded\n" +
		"19/08/2026 16:00:51 TRANSACTION END\n" +
		"17/08/2026 01:21:22 TRANSACTION START\n" +
		"17/08/2026 01:21:22 Terminal ID : S1BGBRR006\n" +
		"17/08/2026 01:21:22 PIN ENTERED\n" +
		"17/08/2026 01:21:34 Amount : 450000\n" +
		"17/08/2026 01:22:19 Banknote separation in cassette : Succeeded\n" +
		"    NO. REKORD 3982\n" +
		"    PENARIKAN  TABUNGAN   \n" +
		"17/08/2026 01:22:19 Shutter Opened for notes removal\n" +
		"17/08/2026 01:22:33 TRANSACTION END\n"

	ejRawPath := filepath.Join(dir, "ej_raw.txt")
	if err := os.WriteFile(ejRawPath, []byte(raw), 0644); err != nil {
		t.Fatal(err)
	}
	ejPath := filepath.Join(dir, "ej_normalized.csv")
	if _, err := ParseATMLogToCSV(ejRawPath, ejPath); err != nil {
		t.Fatal(err)
	}

	// Baris RC asli buat kedua rec (nominalnya cocok sama EJ).
	rcPath := filepath.Join(dir, "rc.txt")
	rcTXT := "5198930240530799/BNI;S1BGBRR006/4842/496876;0210/S1BGBRR006/4842/5198930240530799/0503321332/VL;2500000;D;408600000;08/19/26\n" +
		"5198930898695126/B333;S1BGBRR006/3982/749494;0210/S1BGBRR006/3982/0005198930898695126/VK;450000;K;168650000;08/17/26\n"
	if err := os.WriteFile(rcPath, []byte(rcTXT), 0644); err != nil {
		t.Fatal(err)
	}

	s, err := RunDiff(context.Background(), ejPath, rcPath, filepath.Join(dir, "result.duckdb"),
		LoadOptions{SkipHeader: false})
	if err != nil {
		t.Fatalf("RunDiff error: %v", err)
	}

	if s.SelisihKurang != 1 {
		t.Errorf("SelisihKurang = %d, want 1 (cuma rec 4842 setoran yang ditolak)", s.SelisihKurang)
	}
	if s.SelisihNominalKurang != 2500000 {
		t.Errorf("SelisihNominalKurang = %d, want 2500000 (nominal rec 4842)", s.SelisihNominalKurang)
	}
	if s.Match != 1 {
		t.Errorf("Match = %d, want 1 (rec 3982 penarikan normal, nominal sama)", s.Match)
	}
}
