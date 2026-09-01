package main

import (
	"context"
	"database/sql"
	"fmt"
	"path/filepath"

	_ "github.com/marcboeker/go-duckdb"
)

// LoadOptions merepresentasikan 4 checkbox di panel OPTION.
type LoadOptions struct {
	IgnoreCase   bool `json:"ignore_case"`
	TrimData     bool `json:"trim_data"`
	SkipHeader   bool `json:"skip_header"`
	ValidateData bool `json:"validate_data"`
}

type Summary struct {
	TotalEJ              int   `json:"total_ej"`
	TotalCash            int   `json:"total_cash"`
	Match                int   `json:"match"`
	SelisihKurang        int   `json:"selisih_kurang"`
	TidakDitemukan       int   `json:"tidak_ditemukan"`
	DataInvalid          int   `json:"data_invalid"`
	NominalEJ            int64 `json:"nominal_ej"`
	NominalCash          int64 `json:"nominal_cash"`
	SelisihNominalKurang int64 `json:"selisih_nominal_kurang"`
}

// LoadEJ memparsing raw log EJ jadi CSV ternormalisasi dan mengembalikan jumlah
// transaksi valid yang berhasil diparsing (dipakai buat tampilan "Total Data" pas LOAD DATA 1).
func LoadEJ(ejLogPath, workDir string) (normalizedPath string, count int, err error) {
	normalizedPath = filepath.Join(workDir, "ej_normalized.csv")
	count, err = ParseATMLogToCSV(ejLogPath, normalizedPath)
	if err != nil {
		return "", 0, fmt.Errorf("gagal parsing log EJ: %w", err)
	}
	if count == 0 {
		return "", 0, fmt.Errorf("tidak ada transaksi valid yang berhasil diparsing dari log EJ, cek format file")
	}
	return normalizedPath, count, nil
}

// LoadRC menghitung jumlah baris valid di file RC (dengan opsi yang sama seperti
// yang bakal dipakai pas RunDiff), tanpa join - dipakai buat tampilan "Total Data" pas LOAD DATA 2.
func LoadRC(rcPath string, opts LoadOptions) (count int, err error) {
	db, err := sql.Open("duckdb", "")
	if err != nil {
		return 0, err
	}
	defer db.Close()

	q := fmt.Sprintf(`SELECT COUNT(*) FROM read_csv('%s', delim=';', header=%v, all_varchar=true,
		quote='', ignore_errors=true, null_padding=true)
		WHERE TRY_CAST(split_part(column1, '/', 2) AS INTEGER) IS NOT NULL`, rcPath, opts.SkipHeader)
	if err := db.QueryRow(q).Scan(&count); err != nil {
		return 0, fmt.Errorf("gagal membaca file RC: %w", err)
	}
	if count == 0 {
		return 0, fmt.Errorf("tidak ada baris valid yang terbaca dari file RC, cek format file")
	}
	return count, nil
}

// RunDiff menjalankan join+kategorisasi. Menerima context supaya bisa dibatalkan (tombol STOP).
func RunDiff(ctx context.Context, normalizedEJPath, rcPath, resultDBPath string, opts LoadOptions) (Summary, error) {
	db, err := sql.Open("duckdb", resultDBPath)
	if err != nil {
		return Summary{}, fmt.Errorf("gagal membuka duckdb: %w", err)
	}
	defer db.Close()

	trimText := func(expr string) string {
		if opts.TrimData {
			return "TRIM(" + expr + ")"
		}
		return expr
	}
	upperText := func(expr string) string {
		if opts.IgnoreCase {
			return "UPPER(" + expr + ")"
		}
		return expr
	}

	stmts := []string{
		fmt.Sprintf(`CREATE TABLE ej AS
			SELECT
				%s AS tanggal,
				%s AS terminal,
				%s AS card_masked,
				TRY_CAST(amount AS BIGINT) AS nominal,
				CAST(TRY_CAST(seq_nr AS INTEGER) AS VARCHAR) AS rec_num,
				%s AS ej_status
			FROM read_csv('%s', header=true, delim=',', all_varchar=true)`,
			trimText("timestamp"), trimText("terminal_id"), trimText("card_masked"),
			upperText(trimText("status")), normalizedEJPath),

		fmt.Sprintf(`CREATE TABLE rc AS
			SELECT
				%s AS card_number,
				CAST(TRY_CAST(split_part(column1, '/', 2) AS INTEGER) AS VARCHAR) AS rec_num,
				%s AS no_transaksi,
				TRY_CAST(column3 AS BIGINT) AS nominal,
				%s AS dc_flag,
				column6 AS trx_date
			FROM read_csv('%s', delim=';', header=%v, all_varchar=true, quote='', ignore_errors=true, null_padding=true)
			WHERE TRY_CAST(split_part(column1, '/', 2) AS INTEGER) IS NOT NULL`,
			trimText("split_part(column0, '/', 1)"), trimText("split_part(column2, '/', 5)"),
			upperText(trimText("column4")), rcPath, opts.SkipHeader),

		// Kategori "data_invalid" ditaruh SEBELUM perbandingan nominal, jadi baris
		// yang nominal-nya gagal di-cast (NULL) berhenti di sini dan tidak pernah
		// jatuh ke kondisi lain.
		//
		// "Selisih Kurang" (per keputusan 1 Sep 2026) HANYA berlaku buat baris
		// ROLLBACK yang nominal EJ dan Cash-nya SAMA - bukan lagi generic "cash
		// lebih kecil dari EJ". Kategori "Selisih Lebih" dihapus total. Baris
		// mana pun yang nominalnya beda (rollback atau bukan) jatuh ke
		// 'tidak_ditemukan' - dianggap butuh review manual, bukan match otomatis.
		`CREATE TABLE diff_result AS
			SELECT
				COALESCE(e.rec_num, c.rec_num, '') AS rec_num,
				COALESCE(e.tanggal, c.trx_date, '') AS tanggal,
				COALESCE(e.terminal, '') AS terminal,
				COALESCE(e.card_masked, c.card_number, '') AS no_rekening,
				COALESCE(c.no_transaksi, '') AS no_transaksi,
				e.nominal AS nominal_ej,
				c.nominal AS nominal_cash,
				COALESCE(e.ej_status, 'NOT FOUND') AS ej_status,
				COALESCE(c.dc_flag, 'NOT FOUND') AS cash_status,
				CASE
					WHEN e.rec_num IS NULL OR c.rec_num IS NULL THEN 'tidak_ditemukan'
					WHEN e.nominal IS NULL OR c.nominal IS NULL THEN 'data_invalid'
					WHEN c.nominal = e.nominal AND e.ej_status = 'ROLLBACK' THEN 'selisih_kurang'
					WHEN c.nominal = e.nominal THEN 'match'
					ELSE 'tidak_ditemukan'
				END AS category,
				CASE
					WHEN e.rec_num IS NULL THEN 'Tidak ada di EJ'
					WHEN c.rec_num IS NULL THEN 'Tidak ada di RC'
					WHEN e.nominal IS NULL OR c.nominal IS NULL THEN 'Nominal tidak terbaca (data rusak/kosong)'
					WHEN c.nominal = e.nominal AND e.ej_status = 'ROLLBACK' THEN 'Transaksi rollback di EJ - dana kemungkinan sudah keluar'
					WHEN c.nominal <> e.nominal THEN 'Nominal EJ dan Cash tidak sama'
					ELSE '-'
				END AS keterangan,
				CASE
					WHEN e.rec_num IS NULL OR c.rec_num IS NULL THEN NULL
					WHEN e.nominal IS NULL OR c.nominal IS NULL THEN NULL
					WHEN c.nominal = e.nominal AND e.ej_status = 'ROLLBACK' THEN 'Nasabah Diuntungkan'
					ELSE NULL
				END AS kemungkinan_penyebab
			FROM ej e FULL OUTER JOIN rc c ON e.rec_num = c.rec_num`,
	}

	for _, q := range stmts {
		if _, err := db.ExecContext(ctx, q); err != nil {
			if ctx.Err() != nil {
				return Summary{}, fmt.Errorf("proses dibatalkan (STOP)")
			}
			return Summary{}, fmt.Errorf("query gagal: %w", err)
		}
	}

	var s Summary
	row := db.QueryRowContext(ctx, `
		SELECT
			(SELECT COUNT(*) FROM ej), (SELECT COUNT(*) FROM rc),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'match'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'selisih_kurang'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'tidak_ditemukan'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'data_invalid'),
			(SELECT COALESCE(SUM(nominal), 0) FROM ej),
			(SELECT COALESCE(SUM(nominal), 0) FROM rc),
			(SELECT COALESCE(SUM(nominal_ej - nominal_cash), 0) FROM diff_result WHERE category = 'selisih_kurang')
	`)
	if err := row.Scan(&s.TotalEJ, &s.TotalCash, &s.Match, &s.SelisihKurang, &s.TidakDitemukan, &s.DataInvalid,
		&s.NominalEJ, &s.NominalCash, &s.SelisihNominalKurang); err != nil {
		if ctx.Err() != nil {
			return Summary{}, fmt.Errorf("proses dibatalkan (STOP)")
		}
		return Summary{}, fmt.Errorf("gagal hitung ringkasan: %w", err)
	}

	return s, nil
}

// FetchSummary menghitung ulang Summary dari result DB yang udah jadi (tabel
// ej/rc/diff_result udah ada) - dipakai buat job lama yang datanya udah nggak
// ada di memori (app baru di-restart) tapi file hasilnya masih ada di disk.
func FetchSummary(resultDBPath string) (Summary, error) {
	db, err := sql.Open("duckdb", resultDBPath)
	if err != nil {
		return Summary{}, err
	}
	defer db.Close()

	var s Summary
	row := db.QueryRow(`
		SELECT
			(SELECT COUNT(*) FROM ej), (SELECT COUNT(*) FROM rc),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'match'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'selisih_kurang'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'tidak_ditemukan'),
			(SELECT COUNT(*) FROM diff_result WHERE category = 'data_invalid'),
			(SELECT COALESCE(SUM(nominal), 0) FROM ej),
			(SELECT COALESCE(SUM(nominal), 0) FROM rc),
			(SELECT COALESCE(SUM(nominal_ej - nominal_cash), 0) FROM diff_result WHERE category = 'selisih_kurang')
	`)
	err = row.Scan(&s.TotalEJ, &s.TotalCash, &s.Match, &s.SelisihKurang, &s.TidakDitemukan, &s.DataInvalid,
		&s.NominalEJ, &s.NominalCash, &s.SelisihNominalKurang)
	return s, err
}

type ResultRow struct {
	RecNum              string  `json:"rec_num"`
	Tanggal             string  `json:"tanggal"`
	Terminal            string  `json:"terminal"`
	NoRekening          string  `json:"no_rekening"`
	NoTransaksi         string  `json:"no_transaksi"`
	NominalEJ           *int64  `json:"nominal_ej"`
	NominalCash         *int64  `json:"nominal_cash"`
	EJStatus            string  `json:"ej_status"`
	CashStatus          string  `json:"cash_status"`
	Category            string  `json:"category"`
	Keterangan          string  `json:"keterangan"`
	KemungkinanPenyebab *string `json:"kemungkinan_penyebab"`
}

func FetchResults(resultDBPath, category string, page, pageSize int) ([]ResultRow, int, error) {
	db, err := sql.Open("duckdb", resultDBPath)
	if err != nil {
		return nil, 0, err
	}
	defer db.Close()

	whereClause := "WHERE category = ?"
	args := []interface{}{category}
	if category == "all" {
		whereClause = "WHERE 1=1"
		args = []interface{}{}
	}

	var total int
	countQuery := "SELECT COUNT(*) FROM diff_result " + whereClause
	if err := db.QueryRow(countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	queryArgs := append(append([]interface{}{}, args...), pageSize, offset)
	rows, err := db.Query(`SELECT rec_num, tanggal, terminal, no_rekening, no_transaksi, nominal_ej, nominal_cash,
		ej_status, cash_status, category, keterangan, kemungkinan_penyebab
		FROM diff_result `+whereClause+` ORDER BY rec_num LIMIT ? OFFSET ?`, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var results []ResultRow
	for rows.Next() {
		var r ResultRow
		if err := rows.Scan(&r.RecNum, &r.Tanggal, &r.Terminal, &r.NoRekening, &r.NoTransaksi,
			&r.NominalEJ, &r.NominalCash, &r.EJStatus, &r.CashStatus, &r.Category, &r.Keterangan, &r.KemungkinanPenyebab); err != nil {
			return nil, 0, err
		}
		results = append(results, r)
	}
	return results, total, nil
}
