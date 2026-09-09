# Parse Bankers

Aplikasi desktop Windows untuk rekonsiliasi harian mesin ATM: mencocokkan
**EJ** (Electronic Journal — log mentah mesin ATM) dengan **RC** (file
settlement kas dari sistem bank), lalu mengelompokkan hasilnya (Match,
Selisih Kurang, Tidak Ditemukan) siap diekspor ke Excel/TXT.

Dipakai internal oleh staf cabang BNI — 1 laptop, 1 user, offline
sepenuhnya, tanpa perlu paham coding sama sekali untuk memakainya.

## Tech stack

| Bagian | Teknologi |
|---|---|
| Backend | Go + [Fiber v2](https://gofiber.io/), [DuckDB](https://duckdb.org/) embedded (`go-duckdb`, CGO) buat query rekonsiliasi |
| Frontend | Next.js (static export, client-side murni — tanpa SSR/API routes) |
| Jendela desktop | [WebView2](https://github.com/webview/webview_go) native (bukan Electron/Tauri) — jendela app biasa, bukan tab browser |
| Installer | Inno Setup (wizard Next-Next-Finish, tanpa code-signing) |
| CI/build | GitHub Actions, `windows-latest` runner (satu-satunya tempat exe & installer beneran dirakit) |

## Struktur folder

```
backend/     Go backend (API rekonsiliasi + jendela WebView2 + embed frontend)
frontend/    Next.js UI (static export ke frontend/out, di-embed ke backend/webui)
installer/   Script Inno Setup buat bikin installer Windows
docs/        Roadmap, catatan proses, dan panduan non-teknis
.github/     Workflow CI yang build exe + installer di windows-latest
```

## Menjalankan secara lokal

Backend dan frontend jalan sebagai dua proses terpisah saat development:

```bash
# Terminal 1 — backend (API di 127.0.0.1:8080)
cd backend
go run .

# Terminal 2 — frontend (dev server Next.js, hit API backend lewat CORS)
cd frontend
npm install
npm run dev
```

Build native `go run .`/`go build` butuh dependency CGO: `go-duckdb`
(butuh compiler C/C++) dan `webview_go` (butuh GTK3 + WebKitGTK dev
headers di Linux, atau WebView2 Runtime bawaan di Windows 11). Kalau
cuma mau iterasi di logic parser/rekonsiliasi tanpa peduli jendelanya,
cukup jalankan test:

```bash
cd backend
go test ./...
```

## Build production (exe + installer Windows)

Proses build "asli" **selalu** lewat CI, bukan build manual di laptop
developer — ini juga cara satu-satunya release ini pernah dirakit
(lihat `docs/build-from-zero.md`). Setiap push ke `main`, GitHub
Actions (`.github/workflows/build-windows.yml`) di runner
`windows-latest`:

1. `npm run build` frontend (static export) → di-copy ke `backend/webui`
2. Embed icon resmi ke resource exe (`go-winres`)
3. `go build` backend dengan `CGO_ENABLED=1 GOOS=windows` → `pilot-diff.exe`
4. Ambil `WebView2Loader.dll` (di-load runtime, bukan static link)
5. Compile installer (`installer/setup.iss`) via Inno Setup → `ParseBankers-Setup.exe`

Artifact `ParseBankers-Setup` (installer) dan `pilot-diff-raw-exe` (exe +
dll mentah, buat debug) bisa diunduh dari halaman run Actions repo ini,
atau trigger manual lewat `workflow_dispatch`.

## Dokumentasi lengkap

- [`docs/roadmap.md`](docs/roadmap.md) — keputusan arsitektur & roadmap fase sebelum porting ke desktop
- [`docs/build-from-zero.md`](docs/build-from-zero.md) — narasi teknis, dibangun dari git log asli, tahap demi tahap dari commit pertama sampai sekarang
- [`docs/panduan-membangun-parse-bankers.md`](docs/panduan-membangun-parse-bankers.md) — panduan lengkap (requirement, desain, cara pakai)
- [`docs/panduan-versi-mudah.md`](docs/panduan-versi-mudah.md) — versi ringkas dari panduan di atas
- [`docs/panduan-vibecoding-bangun-dari-nol.md`](docs/panduan-vibecoding-bangun-dari-nol.md) — cerita proses "vibecoding" (prompt-demi-prompt) membangun app ini dari nol
