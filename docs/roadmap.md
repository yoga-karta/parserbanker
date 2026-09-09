# Roadmap: Build Aplikasi Desktop Windows (Rekonsiliasi Bank)

**Dicatat:** 29 Agustus 2026, hasil diskusi planning sebelum mulai implementasi.
**Status:** Roadmap disetujui arah besarnya, 1 keputusan masih pending konfirmasi ke Julian (lihat bagian akhir).

## Konteks

App yang sekarang: web app internal, deployed di server via Docker Compose.
- **Backend:** Go (Fiber v2) + DuckDB embedded (`go-duckdb`, pakai CGO) — `auth.go`,
  `main.go`, `jobstore.go`, `diff.go`, `handlers.go`, `export.go`.
- **Frontend:** Next.js, satu halaman (`pages/index.js`), murni client-side
  (`useState`/`useEffect`/`fetch`), TIDAK ada `getServerSideProps`/API routes/SSR.
- Source code asli cuma hidup di server (`/home/yoga`), tidak ada git — semua
  perubahan dikirim lewat shell script yang nge-patch file di server pakai
  exact-string-match + assert count (fail keras kalau state server beda dari
  asumsi, bukan diam-diam nimpa salah).

Tujuan: bikin versi desktop Windows (.exe/.msi/apapun) yang bisa diinstall
user non-IT di laptopnya sendiri, tanpa dia perlu paham coding sama sekali.

## Review sebelum mulai (per lensa)

### Ponytail (arsitektur/simplicity)
- Stack backend sudah lean (DuckDB embedded, no external DB/queue) — pas
  banget buat jadi desktop app, tidak perlu re-arsitektur besar.
- Deploy scripts disiplin (assert-before-patch, fail loud).
- **Risiko terbesar: tidak ada git.** Source of truth cuma state live di
  server + tumpukan shell script patch berurutan. **Prioritas: `git init` +
  commit source dari server SEBELUM/BARENG kerjaan Windows build**, supaya
  ada titik cabang yang jelas.

### Code Reviewer (backend Go)
- `tmpRoot = "/tmp/pilot-diff-jobs"` di `handlers.go` — hardcoded path Unix,
  **wajib diganti** (`os.TempDir()` / folder app-data lokal) sebelum build
  Windows.
- `app.Listen(":8080")` — listen semua interface. Di server "aman" karena
  Docker port-mapping ke `127.0.0.1`. Untuk desktop exe wajib eksplisit jadi
  `app.Listen("127.0.0.1:8080")`.
- SQL dibangun via `fmt.Sprintf("...'%s'...")` di `diff.go`/`export.go` —
  bukan lubang aktif sekarang (path server-generated dari UUID), tapi gaya
  rapuh, worth dirapikan kalau file-nya dibuka lagi.
- Auth (`CLIENT_USER`/`CLIENT_PASS` shared satu tim) — kandidat kuat buat
  **dihapus** di versi desktop single-user (lihat keputusan di bawah).

### Senior Frontend
- `pages/index.js` murni client-side, cocok banget buat `next export`
  (static export) — kemungkinan besar jalan tanpa perubahan berarti.

### Risiko teknis #1: DuckDB + CGO
`go-duckdb` pakai CGO (binding ke library C++). Tidak bisa cuma
`GOOS=windows go build` dari Linux dengan mudah — perlu toolchain
cross-compile ATAU build asli di Windows (GitHub Actions `windows-latest`
runner, gratis — direkomendasikan). **Ini harus divalidasi paling awal**
(Fase 0) sebelum estimasi lain dipercaya.

## Keputusan tech stack (sudah difinalkan)

| Pertanyaan | Keputusan | Alasan |
|---|---|---|
| Berapa laptop/user? | **1 laptop, 1 user, internal-only, offline** | Menghapus kebutuhan auto-update-check, multi-user threat model |
| Code-signing certificate? | **Tidak pakai untuk sekarang** | App internal, tidak didistribusikan luas. Konsekuensi: SmartScreen "Windows protected your PC" akan muncul tiap kali installer versi baru dikirim (hash beda tiap build) — mitigasi: 1 baris instruksi manual "klik More info → Run anyway" tiap kirim update |
| Minimal versi Windows? | **Windows 11** | WebView2 Runtime sudah bawaan OS (inbox component) — tidak perlu bundling/download terpisah, installer lebih ringan |
| Cara install | **Installer wizard (Inno Setup)**, bukan exe polos | Pengalaman "Next-Next-Finish" yang familiar buat non-IT user; dapat shortcut Desktop+Start Menu dan entry uninstall resmi di "Add or Remove Programs". WiX (.msi asli) ditolak — cocok buat deploy enterprise massal (SCCM/Intune) yang tidak relevan di sini |
| Tampilan aplikasi | **Jendela native via WebView2**, bukan buka di tab browser | User non-IT akan bingung/curiga lihat address bar `localhost:8080`. WebView2 (bawaan Win11) kasih jendela app biasa dengan icon taskbar sendiri, tanpa nambah toolchain (bukan Electron — terlalu berat/bundling Chromium+Node sendiri; bukan Tauri — butuh toolchain Rust penuh cuma buat app sekecil ini) |
| Login/auth screen | **Rekomendasi: dihapus** (masih perlu 1x konfirmasi ke Julian) | Threat model yang tadinya jadi alasan ada login (banyak orang akses server yang sama lewat network) sudah tidak ada — ini 1 laptop, 1 orang, app jalan lokal |
| Auto-update mechanism | **Tidak masuk scope sekarang** | App offline, 1 laptop — kirim installer baru manual tiap ada fix sudah cukup. Bisa dibuka lagi jadi fase terpisah kalau nanti nambah laptop/tim |

## Roadmap final

| Fase | Isi | Estimasi |
|---|---|---|
| 0 | Spike: pastikan `go-duckdb` (CGO) + WebView2 beneran bisa dibuild jadi 1 exe Windows 11 | 0.5 hari |
| 1 | Porting backend: fix path `/tmp/...` → `os.TempDir()`, bind `127.0.0.1`, hapus login/auth (setelah dikonfirmasi Julian), simpan data di folder lokal | 0.5 hari |
| 2 | Bungkus jendela native pakai WebView2 (ganti "buka browser" jadi jendela app beneran, icon+title sendiri) | 0.5 hari |
| 3 | Static export frontend (`next export`) + embed ke binary Go (`//go:embed`), serve dari Fiber | 0.5 hari |
| 4 | Installer Inno Setup: icon, shortcut Desktop+Start Menu, entry uninstall | 0.5 hari |
| 5 | Testing di laptop Windows 11 asli client, jalanin end-to-end pakai data real (`EJ.TXT`+`RC-file.txt`), cek export Excel/TXT | 0.5 hari |

**Total: ~3 hari kerja.**

## Yang masih pending / perlu dikonfirmasi ke Julian

1. **Hapus login screen — setuju, atau tetap dipertahankan** karena alasan
   lain (misal audit trail internal), bukan soal keamanan jaringan?

## Catatan operasional

- SmartScreen akan muncul tiap kirim installer versi baru (karena tanpa
  code-signing) — siapkan 1 baris instruksi tiap kirim update: *"kalau
  muncul kotak biru, klik 'More info' → 'Run anyway'."*
- Sebelum/bareng kerjaan ini, `git init` source dari server jadi prioritas
  supaya build desktop punya titik cabang yang jelas dari source of truth,
  bukan garpu selamanya dari server.
