package main

import (
	"database/sql"
	"fmt"

	"github.com/xuri/excelize/v2"
)

// validExportCategory adalah allow-list kategori yang boleh dipakai buat filter
// export. category datang dari query param (?category=...), jadi divalidasi
// terhadap daftar ini dulu sebelum ditempel ke SQL - "" atau "all" berarti
// tanpa filter (export semua).
var validExportCategory = map[string]bool{
	"match": true, "selisih_kurang": true,
	"tidak_ditemukan": true, "data_invalid": true,
}

func exportWhereClause(category string) string {
	if validExportCategory[category] {
		return fmt.Sprintf(" WHERE category = '%s'", category)
	}
	return ""
}

// ExportTXT menulis diff_result (semua kategori, atau cuma satu kategori kalau
// category diisi) sebagai pipe-delimited text, langsung lewat DuckDB COPY.
func ExportTXT(resultDBPath, outputPath, category string) error {
	db, err := sql.Open("duckdb", resultDBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	_, err = db.Exec(fmt.Sprintf(`
		COPY (
			SELECT rec_num AS "RecNum", tanggal AS "Tanggal", terminal AS "Terminal",
				no_rekening AS "NoRekening", no_transaksi AS "NoTransaksi",
				nominal_ej AS "NominalEJ", nominal_cash AS "NominalCash",
				ej_status AS "EJStatus", cash_status AS "CashStatus",
				category AS "Result", kemungkinan_penyebab AS "KemungkinanPenyebab", keterangan AS "Keterangan"
			FROM diff_result%s ORDER BY rec_num
		) TO '%s' (DELIMITER '|', HEADER true)`, exportWhereClause(category), outputPath))
	return err
}

var categoryFillColor = map[string]string{
	"match":           "E6F9F0",
	"selisih_kurang":  "FDEAEA",
	"tidak_ditemukan": "F3EAFB",
	"data_invalid":    "FEFBE6",
}

// ExportExcel menulis diff_result (semua kategori, atau cuma satu kategori
// kalau category diisi) sebagai .xlsx dengan header tebal + pewarnaan per baris.
func ExportExcel(resultDBPath, outputPath, category string) error {
	db, err := sql.Open("duckdb", resultDBPath)
	if err != nil {
		return err
	}
	defer db.Close()

	rows, err := db.Query(`SELECT rec_num, tanggal, terminal, no_rekening, no_transaksi, nominal_ej, nominal_cash,
		ej_status, cash_status, category, keterangan, kemungkinan_penyebab
		FROM diff_result` + exportWhereClause(category) + ` ORDER BY rec_num`)
	if err != nil {
		return err
	}
	defer rows.Close()

	f := excelize.NewFile()
	defer f.Close()
	sheet := "Hasil Rekonsiliasi"
	f.SetSheetName("Sheet1", sheet)

	headers := []string{"No", "Rec Num", "Tanggal", "Terminal", "No Rekening", "No Transaksi",
		"Nominal EJ", "Nominal Cash", "EJ Status", "Cash Status", "Result", "Kemungkinan Penyebab", "Keterangan"}
	for i, h := range headers {
		cell, _ := excelize.CoordinatesToCellName(i+1, 1)
		f.SetCellValue(sheet, cell, h)
	}
	headerStyle, _ := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{Bold: true},
		Fill: excelize.Fill{Type: "pattern", Color: []string{"DDDDDD"}, Pattern: 1},
	})
	f.SetCellStyle(sheet, "A1", fmt.Sprintf("%s1", colName(len(headers))), headerStyle)

	styleCache := map[string]int{}
	getCategoryStyle := func(category string) int {
		if id, ok := styleCache[category]; ok {
			return id
		}
		argb := categoryFillColor[category]
		if argb == "" {
			argb = "FFFFFF"
		}
		id, _ := f.NewStyle(&excelize.Style{Fill: excelize.Fill{Type: "pattern", Color: []string{argb}, Pattern: 1}})
		styleCache[category] = id
		return id
	}

	rowIdx := 2
	for rows.Next() {
		var recNum, tanggal, terminal, noRek, noTrx, ejStatus, cashStatus, category, keterangan string
		var nominalEJ, nominalCash *int64
		var penyebab *string
		if err := rows.Scan(&recNum, &tanggal, &terminal, &noRek, &noTrx, &nominalEJ, &nominalCash,
			&ejStatus, &cashStatus, &category, &keterangan, &penyebab); err != nil {
			return err
		}

		penyebabVal := ""
		if penyebab != nil {
			penyebabVal = *penyebab
		}
		vals := []interface{}{rowIdx - 1, recNum, tanggal, terminal, noRek, noTrx,
			nominalEJ, nominalCash, ejStatus, cashStatus, category, penyebabVal, keterangan}
		for i, v := range vals {
			cell, _ := excelize.CoordinatesToCellName(i+1, rowIdx)
			f.SetCellValue(sheet, cell, v)
		}
		styleID := getCategoryStyle(category)
		f.SetCellStyle(sheet, fmt.Sprintf("A%d", rowIdx), fmt.Sprintf("%s%d", colName(len(headers)), rowIdx), styleID)
		rowIdx++
	}

	for i := range headers {
		col := colName(i + 1)
		f.SetColWidth(sheet, col, col, 16)
	}

	return f.SaveAs(outputPath)
}

func colName(n int) string {
	c, _ := excelize.ColumnNumberToName(n)
	return c
}
