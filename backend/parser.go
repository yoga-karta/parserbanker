package main

import (
	"bufio"
	"encoding/csv"
	"os"
	"regexp"
	"strings"
)

type ATMLogRecord struct {
	Timestamp  string
	CardNum    string
	TerminalID string
	Amount     string
	SeqNr      string
	Status     string
}

var (
	reCard     = regexp.MustCompile(`CARD NUMBER\s+([0-9*]+)`)
	reTerminal = regexp.MustCompile(`Terminal ID\s*\[([^\]]*)\]`)
	reAmount   = regexp.MustCompile(`Amount\s*:\s*([0-9]+)`)
	reSeq      = regexp.MustCompile(`TRAN SEQ NR\s*\[([0-9]+)\]`)
)

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
			current = ATMLogRecord{}
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
			current = ATMLogRecord{}
			continue
		}

		if match := reCard.FindStringSubmatch(line); len(match) > 1 {
			current.CardNum = match[1]
		} else if match := reTerminal.FindStringSubmatch(line); len(match) > 1 {
			current.TerminalID = strings.TrimSpace(match[1])
		} else if match := reAmount.FindStringSubmatch(line); len(match) > 1 {
			current.Amount = match[1]
		} else if match := reSeq.FindStringSubmatch(line); len(match) > 1 {
			current.SeqNr = match[1]
		} else if strings.Contains(line, "TRANSACTION REPLIED") {
			current.Status = "SUCCESS"
		} else if strings.Contains(line, "TRANSACTION FAILED") || strings.Contains(line, "TRANSACTION DECLINED") {
			current.Status = "FAILED"
		} else if strings.Contains(strings.ToUpper(line), "ROLLBACK") {
			current.Status = "ROLLBACK"
		}
	}
	flush()

	return count, scanner.Err()
}
