package main

import (
	"bufio"
	"encoding/csv"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type ATMLogRecord struct {
	Timestamp  string
	CardNum    string
	TerminalID string
	Amount     string
	NotesIn    bool
	NotesTotal string
	SeqNr      string
	Status     string
}

var (
	// Mask kartu dicetak beda-beda antar merek: DN200V all-caps tanpa titik dua
	// ("CARD NUMBER 532659******4565"), Hitachi pakai titik dua, Hyosung tanpa
	// titik dua dan mask-nya 'X' bukan '*'. Yang versi kurung siku di blok
	// [Transaction record] sengaja NGGAK ikut ke-match: di OKI isinya PAN utuh
	// tanpa mask, jangan sampai nomor kartu polos bocor ke hasil rekon.
	reCard = regexp.MustCompile(`(?i)CARD NUMBER\s*:?\s*([0-9*X]+)`)
	// OKI nulis "Terminal ID : S1BGBRR006" (titik dua, tanpa kurung siku),
	// merek lain "Terminal ID    [S1GBMSR055]".
	reTerminal = regexp.MustCompile(`Terminal ID\s*[:\[]\s*([^\]\s]*)`)
	reAmount   = regexp.MustCompile(`Amount\s*:\s*([0-9]+)`)
	// Anchor nomor urut transaksi. "TRAN SEQ NR" CUMA ada di EJ DN200V - di
	// Hyosung/Hitachi/OKI nol kemunculan, jadi dulu SeqNr selalu kosong, flush()
	// nggak pernah nulis baris, dan LoadEJ nolak filenya ("tidak ada transaksi
	// valid"). Struk cetak "NO. REKORD <n>" ada di SEMUA merek dan nilainya sama
	// persis dengan rec_num yang dipakai buat join ke RC, jadi dipakai sebagai
	// anchor universal (DN200V tetap kebaca dari dua-duanya, nilainya sama).
	reSeq = regexp.MustCompile(`(?:TRAN SEQ NR\s*\[|NO\. REKORD\s*:?\s*)([0-9]+)`)
	// Hasil hitungan uang fisik yang masuk ke mesin - dipakai sebagai bukti
	// faktor skala nominal setor tunai, lihat komentar di cabang reAmount.
	reNotesTotal = regexp.MustCompile(`(?i)TOTAL AMOUNT IDR\s*([0-9]+)`)
)

// resetKeepTerminal ngosongin field per-transaksi (Amount, SeqNr, CardNum,
// Status, hitungan uang) tapi MEMPERTAHANKAN Terminal ID. Satu file EJ = satu
// mesin fisik, jadi Terminal ID-nya konstan buat semua transaksi di file itu -
// bukan data per-transaksi yang harus di-reset. Ini yang bikin EJ Hyosung &
// Hitachi dulu terminal_id-nya kosong 100%: kedua merek nulis DUA baris
// "TRANSACTION START" berturut-turut per satu transaksi (yang pertama pembuka
// sesi kartu, yang kedua pembuka transaksinya) dengan baris "Terminal ID" ke-log
// di antara keduanya, jadi reset di START kedua ngehapus Terminal ID yang barusan
// kebaca. DN200V & OKI cuma punya satu START dan Terminal ID-nya ke-log ulang
// tiap blok, jadi mereka nggak keubah sama sekali.
func resetKeepTerminal(prev ATMLogRecord) ATMLogRecord {
	return ATMLogRecord{TerminalID: prev.TerminalID}
}

// ParseATMLogToCSV membaca log mentah EJ ATM (multi-line, per-transaction-block)
// dan mengekstrak jadi CSV ternormalisasi. Blok dipotong berdasarkan TRANSACTION
// START berikutnya (atau EOF), bukan baris "akhir transaksi" tertentu - supaya
// semua jenis transaksi (bukan cuma yang berakhir di state tertentu) ke-parse.
func ParseATMLogToCSV(rawLogPath, outputCSVPath string) (int, error) {
	inFile, err := os.Open(rawLogPath)
	if err != nil {
		return 0, err
	}
	defer inFile.Close()

	outFile, err := os.Create(outputCSVPath)
	if err != nil {
		return 0, err
	}
	defer outFile.Close()

	writer := csv.NewWriter(outFile)
	defer writer.Flush()
	writer.Write([]string{"timestamp", "card_masked", "terminal_id", "amount", "seq_nr", "status"})

	scanner := bufio.NewScanner(inFile)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)

	var current ATMLogRecord
	count := 0

	// File EJ mentah kadang punya section yang ke-duplikat persis (misal gara-gara
	// export yang overlap tanggal) - blok dengan timestamp+seq+amount identik
	// dianggap transaksi fisik yang sama, cuma dihitung sekali.
	seen := map[string]bool{}

	flush := func() {
		if current.SeqNr != "" && current.Amount != "" {
			dedupKey := current.Timestamp + "|" + current.SeqNr + "|" + current.Amount
			if seen[dedupKey] {
				return
			}
			seen[dedupKey] = true
			writer.Write([]string{
				current.Timestamp, current.CardNum, current.TerminalID, current.Amount, current.SeqNr, current.Status,
			})
			count++
		}
	}

	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "TRANSACTION START") {
			flush()
			current = resetKeepTerminal(current)
			parts := strings.SplitN(line, " ", 3)
			if len(parts) >= 2 {
				current.Timestamp = parts[0] + " " + parts[1]
			}
			continue
		}

		if strings.Contains(line, "TRANSACTION END") {
			// Flush & reset di sini, bukan cuma di START berikutnya - baris mesin
			// (mis. "Rollback Notes"/"Rollback OK" pas ATM narik balik uang yang
			// nggak diambil nasabah) kadang muncul SETELAH END tapi SEBELUM START
			// transaksi berikutnya. Kalau current nggak direset di sini, baris itu
			// kebaca sebagai bagian transaksi yang barusan kelar dan salah nge-mark
			// transaksi itu ROLLBACK padahal blok transaksinya sendiri bersih.
			flush()
			current = resetKeepTerminal(current)
			continue
		}

		if strings.Contains(line, "PIN ENTERED") && current.SeqNr != "" {
			// Nasabah masuk PIN lagi setelah satu percobaan sebelumnya udah dapat
			// TRAN SEQ NR sendiri = mulai operasi baru di sesi kartu yang sama
			// (mis. abis deposit pertama sukses, ATM minta PIN lagi buat percobaan
			// berikutnya yang bisa aja gagal/rollback). Flush percobaan sebelumnya
			// dulu, jangan sampai status percobaan baru numpuk ke yang lama
			// (kejadian nyata: rec_num 407 - deposit sukses ke-tag ROLLBACK gara-gara
			// percobaan kedua di sesi kartu yang sama gagal PING lalu di-rollback).
			flush()
			current.Amount = ""
			current.SeqNr = ""
			current.Status = ""
		}

		if match := reCard.FindStringSubmatch(line); len(match) > 1 {
			current.CardNum = match[1]
		} else if match := reTerminal.FindStringSubmatch(line); len(match) > 1 {
			current.TerminalID = strings.TrimSpace(match[1])
		} else if match := reAmount.FindStringSubmatch(line); len(match) > 1 {
			if current.Amount != "" {
				// Satu blok TRANSACTION START..END bisa berisi LEBIH DARI SATU
				// percobaan (mis. deposit pertama sukses dapat TRAN SEQ NR, nasabah
				// lanjut nyoba deposit lagi di sesi kartu yang sama dan itu yang
				// PING ERROR/rollback). Flush percobaan sebelumnya sbg baris sendiri
				// dulu sebelum mulai nampung percobaan baru - kalau nggak, status
				// ROLLBACK punya percobaan KEDUA bakal numpuk salah ke percobaan
				// PERTAMA yang udah sukses & punya TRAN SEQ NR sendiri (kejadian
				// nyata: rec_num 407 - deposit sukses ke-tag ROLLBACK gara-gara
				// percobaan berikutnya di sesi kartu yang sama gagal PING).
				flush()
				current.Amount = ""
				current.SeqNr = ""
				current.Status = ""
			}
			current.Amount = match[1]
			// Hitachi & Hyosung ngirim nominal setor tunai apa adanya dari field
			// ISO-8583 yang punya 2 desimal implisit, jadi nilai di baris "Amount :"
			// 100x nominal asli (mis. EJ Hitachi rec 5830 "Amount : 30000000" vs RC
			// 300000; Hyosung rec 8783 "Amount : 195000000" vs RC 1950000). DN200V &
			// OKI nggak begitu, nominalnya 1:1. Faktor 100-nya NGGAK ditebak per
			// merek: mesinnya sendiri yang harus buktiin lewat hasil hitungan uang
			// fisik yang MASUK ("Notes Counted" -> "Total Amount IDR <n>", selalu
			// ke-log SEBELUM baris Amount). Hitungan uang KELUAR ("NOTES DISPENSED")
			// nggak dipakai: EJ DN200V S1DTRBR013 rec 3285 punya satu sesi berisi
			// permintaan Rp5.000.000 dan pengeluaran Rp50.000 - persis 100x tapi itu
			// selisih beneran, bukan beda skala. Kalau nggak persis Amount/100,
			// nominal dibiarkan apa adanya - ini duit nasabah. Dikunci di sini,
			// BUKAN pas flush: satu sesi kartu bisa isi dua setoran, dan pas setoran
			// pertama di-flush hitungan uangnya udah kegantikan punya setoran kedua.
			if t, err := strconv.ParseInt(current.NotesTotal, 10, 64); err == nil && t > 0 {
				if a, err := strconv.ParseInt(current.Amount, 10, 64); err == nil && a == t*100 {
					current.Amount = current.NotesTotal
				}
			}
		} else if upper := strings.ToUpper(line); strings.Contains(upper, "NOTES COUNTED") {
			current.NotesIn = true
		} else if strings.Contains(upper, "NOTES DISPENSED") {
			// Hitungan uang KELUAR (penarikan) nggak boleh jadi bukti skala nominal
			// setor - lihat komentar di baris Amount.
			current.NotesIn = false
		} else if match := reNotesTotal.FindStringSubmatch(line); len(match) > 1 && current.NotesIn {
			// Sengaja NGGAK ikut direset bareng Amount/SeqNr/Status pas percobaan
			// baru di sesi kartu yang sama: mesin cuma ngitung uangnya sekali, terus
			// nasabah ulang PIN-nya (mis. Hyosung rec 8833), jadi bukti nominalnya
			// harus tetap kepakai buat percobaan berikutnya di blok yang sama.
			current.NotesTotal = match[1]
		} else if match := reSeq.FindStringSubmatch(line); len(match) > 1 {
			// Dua anchor-nya nulis angka yang sama dengan format beda: "TRAN SEQ NR
			// [0660]" pakai nol-padding, struk " NO. REKORD  660" nggak. Nol depannya
			// dibuang biar satu transaksi cuma punya satu bentuk seq - kalau nggak,
			// blok yang ke-log dua kali (retry ke host) lolos dedup dan kehitung
			// dobel. Nilai "0" tetap disimpan "0", bukan string kosong (string kosong
			// bikin barisnya nggak ditulis sama sekali).
			current.SeqNr = strings.TrimLeft(match[1], "0")
			if current.SeqNr == "" {
				current.SeqNr = "0"
			}
		} else if strings.Contains(line, "TRANSACTION REPLIED") {
			current.Status = "SUCCESS"
		} else if strings.Contains(line, "TRANSACTION FAILED") || strings.Contains(line, "TRANSACTION DECLINED") {
			current.Status = "FAILED"
		} else if strings.Contains(strings.ToUpper(line), "ROLLBACK OK") {
			// Indikator harus persis "Rollback OK" (bukan cuma kata "Rollback" di
			// baris manapun, mis. "----- Rollback Notes ----") - sebuah transaksi
			// kadang punya 2 baris "Rollback Notes" tapi cuma yang beneran
			// ke-rollback yang diikuti "Rollback OK"; baris kedua cuma housekeeping
			// penutupan transaksi dan bukan rollback sungguhan.
			current.Status = "ROLLBACK OK"
		} else if strings.Contains(strings.ToUpper(line), "ROLLBACK NOTES SUCCESSFULLY") {
			// Status EJ kedua (per keputusan 3 Sep 2026) yang berarti dana sudah
			// keluar mesin: sama perlakuannya dengan "Rollback OK" buat kategorisasi
			// Selisih Kurang (lihat diff.go), tapi disimpan apa adanya (bukan
			// dinormalisasi jadi satu label generik) supaya EJ Status yang
			// ditampilkan ke user tetap persis sama seperti di log mentah.
			current.Status = "ROLLBACK NOTES SUCCESSFULLY"
		} else if strings.Contains(strings.ToUpper(line), "SHUTTER OPENED FOR NOTES REMOVAL") {
			// Status EJ ketiga (per keputusan 3 Sep 2026), perlakuan sama seperti
			// dua status rollback di atas.
			current.Status = "SHUTTER OPENED FOR NOTES REMOVAL"
		}
	}
	flush()

	return count, scanner.Err()
}
