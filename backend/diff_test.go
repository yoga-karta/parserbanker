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
		"01/09/2026 10:01:00,222,T1,200000,2,ROLLBACK\n" + // selisih_kurang (rollback, nominal sama)
		"01/09/2026 10:02:00,333,T1,300000,3,ROLLBACK\n" + // dulu otomatis selisih_kurang, sekarang tidak_ditemukan (rollback tapi nominal beda)
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
}
