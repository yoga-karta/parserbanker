package main

import (
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
