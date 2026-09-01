package main

import (
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
	"github.com/webview/webview_go"
)

const addr = "127.0.0.1:8080"

func main() {
	if err := os.MkdirAll(tmpRoot, 0700); err != nil {
		log.Fatalf("gagal membuat tmp root: %v", err)
	}

	store := NewJobStore()

	app := fiber.New(fiber.Config{
		BodyLimit: 1024 * 1024 * 1024,
	})

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,DELETE",
		AllowHeaders: "Origin, Content-Type, Accept, X-API-Token",
	}))

	api := app.Group("/api")

	// Flow 2 langkah: create -> load (ej & rc terpisah) -> process -> (stop opsional) -> results/export
	api.Post("/jobs", handleCreateJob(store))
	api.Post("/jobs/:id/load/:role", handleLoadFile(store))
	api.Post("/jobs/:id/process", handleProcess(store))
	api.Post("/jobs/:id/stop", handleStop(store))
	api.Post("/jobs/:id/reset", handleResetJob(store))
	api.Get("/jobs/:id", handleJobStatus(store))
	api.Get("/jobs/:id/log", handleJobLog(store))
	api.Get("/jobs/:id/results", handleJobResults(store))
	api.Get("/jobs/:id/export", handleExport(store))

	api.Get("/history", handleHistory)
	api.Delete("/history", handleClearHistory)

	// Frontend static (hasil next export, di-embed via embed.go) - didaftar
	// belakangan supaya /api/* di atas selalu menang duluan.
	webRoot, err := fs.Sub(webUI, "webui")
	if err != nil {
		log.Fatalf("gagal membuka embedded webui: %v", err)
	}
	app.Use("/", filesystem.New(filesystem.Config{
		Root:  http.FS(webRoot),
		Index: "index.html",
	}))

	go func() {
		log.Fatal(app.Listen(addr))
	}()
	waitForServer(addr)

	w := webview.New(false)
	defer w.Destroy()
	w.SetTitle("Parse Bankers - Reconciliation Portal")
	w.SetSize(1280, 800, webview.HintNone)
	setWindowIconFromExe(w.Window())
	w.Navigate("http://" + addr + "/")
	w.Run()
}

// waitForServer nunggu backend siap nerima koneksi sebelum window dibuka,
// biar nggak race dengan app.Listen di goroutine atas (max 5 detik).
func waitForServer(addr string) {
	for i := 0; i < 100; i++ {
		conn, err := net.DialTimeout("tcp", addr, 50*time.Millisecond)
		if err == nil {
			conn.Close()
			return
		}
		time.Sleep(50 * time.Millisecond)
	}
	log.Fatal("server tidak siap dalam 5 detik")
}
