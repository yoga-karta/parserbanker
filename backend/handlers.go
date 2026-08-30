package main

import (
	"context"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
)

var tmpRoot = filepath.Join(os.TempDir(), "pilot-diff-jobs")

func parseOptions(c *fiber.Ctx) LoadOptions {
	b := func(key string) bool {
		v := c.FormValue(key)
		if v == "" {
			v = c.Query(key)
		}
		return v == "true" || v == "1" || v == "on"
	}
	return LoadOptions{
		IgnoreCase:   b("ignore_case"),
		TrimData:     b("trim_data"),
		SkipHeader:   b("skip_header"),
		ValidateData: b("validate_data"),
	}
}

// handleCreateJob: langkah 0, bikin job kosong buat nampung file yang bakal di-load.
func handleCreateJob(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		jobID := uuid.NewString()
		jobDir := filepath.Join(tmpRoot, jobID)
		if err := os.MkdirAll(jobDir, 0700); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "gagal membuat direktori kerja"})
		}
		store.Create(jobID)
		return c.JSON(fiber.Map{"job_id": jobID})
	}
}

// handleLoadFile: LOAD DATA 1 (role=ej) / LOAD DATA 2 (role=rc). Upload 1 file,
// langsung dihitung total record-nya tanpa menjalankan join.
func handleLoadFile(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		role := c.Params("role")
		if role != "ej" && role != "rc" {
			return c.Status(400).JSON(fiber.Map{"error": "role harus 'ej' atau 'rc'"})
		}
		if _, ok := store.Get(id); !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}

		file, err := c.FormFile("file")
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "file wajib diisi"})
		}

		opts := parseOptions(c)
		store.SetOptions(id, opts)

		jobDir := filepath.Join(tmpRoot, id)
		savePath := filepath.Join(jobDir, role+"_raw.txt")
		if err := c.SaveFile(file, savePath); err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "gagal menyimpan file"})
		}

		if role == "ej" {
			normPath, count, err := LoadEJ(savePath, jobDir)
			if err != nil {
				return c.Status(400).JSON(fiber.Map{"error": err.Error()})
			}
			store.SetEJLoaded(id, normPath, count)
			job, _ := store.Snapshot(id)
			return c.JSON(fiber.Map{"total_record": count, "status": job.Status})
		}

		count, err := LoadRC(savePath, opts)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": err.Error()})
		}
		store.SetRCLoaded(id, savePath, count)
		job, _ := store.Snapshot(id)
		return c.JSON(fiber.Map{"total_record": count, "status": job.Status})
	}
}

// handleProcess: PROCESS/RECON. Kedua file harus sudah di-load.
func handleProcess(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Snapshot(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}
		if !job.EJLoaded || !job.RCLoaded {
			return c.Status(400).JSON(fiber.Map{"error": "kedua file (EJ & RC) harus di-load dulu sebelum proses"})
		}

		ctx, cancel := context.WithCancel(context.Background())
		store.SetProcessing(id, cancel)

		resultDBPath := filepath.Join(tmpRoot, id, "result.duckdb")
		username, _ := c.Locals("username").(string)

		// job adalah SALINAN (dari Snapshot), jadi aman ditutup di goroutine ini -
		// EJPath/RCPath/Options gak akan berubah di bawah kaki kita walau job asli
		// di store terus di-mutasi (Status, Log, dst) oleh request lain yang jalan
		// bersamaan.
		go func() {
			summary, err := RunDiff(ctx, job.EJPath, job.RCPath, resultDBPath, job.Options)
			os.Remove(job.EJPath)
			os.Remove(job.RCPath)
			if err != nil {
				if ctx.Err() != nil {
					return // sudah ditangani oleh Stop(), jangan overwrite status "stopped" jadi "error"
				}
				store.SetError(id, err.Error())
				_ = AppendHistory(HistoryEntry{JobID: id, Username: username, Timestamp: time.Now(), Status: "error", Error: err.Error()})
				return
			}
			store.SetDone(id, summary, resultDBPath)
			_ = AppendHistory(HistoryEntry{
				JobID: id, Username: username, Timestamp: time.Now(), Status: "done",
				Counts: map[string]int{
					"match": summary.Match, "selisih_lebih": summary.SelisihLebih,
					"selisih_kurang": summary.SelisihKurang, "tidak_ditemukan": summary.TidakDitemukan,
				},
			})
		}()

		return c.JSON(fiber.Map{"status": "processing"})
	}
}

func handleStop(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		if !store.Stop(id) {
			return c.Status(400).JSON(fiber.Map{"error": "tidak ada proses yang sedang berjalan untuk job ini"})
		}
		return c.JSON(fiber.Map{"status": "stopped"})
	}
}

// handleReset: RESET. Buang job lama beserta file-nya, siap mulai dari nol.
func handleResetJob(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		store.Stop(id) // batalkan dulu kalau masih jalan
		os.RemoveAll(filepath.Join(tmpRoot, id))
		return c.JSON(fiber.Map{"status": "reset"})
	}
}

func handleJobStatus(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Snapshot(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}
		return c.JSON(job)
	}
}

func handleJobLog(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Snapshot(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}
		return c.JSON(fiber.Map{"log": job.Log})
	}
}

func handleJobResults(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Snapshot(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}
		if job.Status != StatusDone {
			return c.Status(400).JSON(fiber.Map{"error": "job belum selesai diproses"})
		}

		category := c.Query("category", "selisih_kurang")
		page, _ := strconv.Atoi(c.Query("page", "1"))
		pageSize, _ := strconv.Atoi(c.Query("page_size", "50"))
		if page < 1 {
			page = 1
		}
		if pageSize < 1 || pageSize > 500 {
			pageSize = 50
		}

		results, total, err := FetchResults(job.DBPath, category, page, pageSize)
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}

		return c.JSON(fiber.Map{
			"results": results, "total": total, "page": page, "page_size": pageSize,
		})
	}
}

func handleHistory(c *fiber.Ctx) error {
	entries, err := ReadHistory(50)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal membaca riwayat"})
	}
	return c.JSON(fiber.Map{"history": entries})
}

func handleClearHistory(c *fiber.Ctx) error {
	if err := ClearHistory(); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": "gagal menghapus riwayat log"})
	}
	return c.JSON(fiber.Map{"message": "riwayat berhasil dibersihkan"})
}

// handleExport menghasilkan file Excel/TXT dari hasil diff_result dan langsung
// mengirimkannya sebagai download. Tanpa ?category= (atau category=all) berarti
// SELURUH baris (ditandai per baris lewat kolom Result) - dengan ?category=xxx
// cuma baris kategori itu yang di-export.
func handleExport(store *JobStore) fiber.Handler {
	return func(c *fiber.Ctx) error {
		id := c.Params("id")
		job, ok := store.Snapshot(id)
		if !ok {
			return c.Status(404).JSON(fiber.Map{"error": "job tidak ditemukan"})
		}
		if job.Status != StatusDone {
			return c.Status(400).JSON(fiber.Map{"error": "job belum selesai diproses"})
		}

		format := c.Query("format", "xlsx")
		category := c.Query("category", "all")
		suffix := "semua"
		if validExportCategory[category] {
			suffix = category
		}
		outPath := filepath.Join(filepath.Dir(job.DBPath), "export_"+suffix+"."+format)

		var err error
		var filename string
		switch format {
		case "txt":
			err = ExportTXT(job.DBPath, outPath, category)
			filename = "hasil_rekonsiliasi_" + suffix + ".txt"
		case "xlsx":
			err = ExportExcel(job.DBPath, outPath, category)
			filename = "hasil_rekonsiliasi_" + suffix + ".xlsx"
		default:
			return c.Status(400).JSON(fiber.Map{"error": "format tidak didukung, gunakan xlsx atau txt"})
		}
		if err != nil {
			return c.Status(500).JSON(fiber.Map{"error": "gagal export: " + err.Error()})
		}

		c.Set("Content-Disposition", `attachment; filename="`+filename+`"`)
		return c.SendFile(outPath)
	}
}
