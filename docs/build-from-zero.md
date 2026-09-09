# Parse Bankers — Dibangun dari Nol, Step by Step (dari Git Log Asli)

Dokumen ini bukan cerita yang ditulis ulang — ini adalah urutan commit asli
di `git log`, dijelaskan satu-satu. Tujuannya membuktikan proses coding-nya
nyata: apa yang berubah, kenapa berubah, dan bagaimana tiap perubahan
divalidasi, dari commit pertama (`ddb9050`) sampai commit terbaru
(`af0e485`) di branch `main`.

Jalankan `git log --oneline` di root repo ini untuk melihat urutan penuh —
setiap judul tahap di bawah adalah commit sungguhan yang bisa dibuka isinya
dengan `git show <hash>`.

## Tahap 0 — Titik awal: kode dari server, belum ada git

**Commit:** `ddb9050`, `d89648c`, `e5979aa`

Aplikasi ini awalnya adalah web app internal yang hidup di satu server
(`/home/yoga`), dijalankan lewat Docker Compose. **Tidak ada git sama
sekali** — setiap perubahan sebelumnya dikirim lewat shell script yang
nge-patch file langsung di server pakai exact-string-match + assert count
(sengaja fail keras kalau state server beda dari asumsi, bukan diam-diam
menimpa yang salah).

Commit `ddb9050` adalah snapshot pertama: seluruh source backend (Go/Fiber,
autentikasi, parser, diff engine) dan frontend (satu halaman Next.js
1.464 baris) di-commit apa adanya dari server — 2.715 baris kode, titik
cabang pertama yang jelas. Ini prioritas nomor satu sebelum kerjaan lain
dimulai (lihat `docs/roadmap.md`, bagian "Ponytail"): tanpa git, tidak ada
cara aman untuk bereksperimen porting ke desktop tanpa risiko merusak
versi produksi yang sedang jalan.

## Tahap 1 — Fase 0: validasi risiko teknis terbesar duluan

**Commit:** `68629ee`

Sebelum janji "3 hari kerja" di roadmap dipercaya, risiko teknis terbesar
harus divalidasi dulu: backend pakai `go-duckdb`, yang mengikat ke library
C++ lewat CGO — tidak bisa sekadar `GOOS=windows go build` dari Linux.
Commit ini menambahkan workflow GitHub Actions minimal yang cuma satu
tugas: pastikan backend Go bisa di-cross-build jadi exe Windows di runner
`windows-latest` milik GitHub (gratis). Baru setelah spike ini hijau,
fase-fase berikutnya dikerjakan dengan percaya diri.

## Tahap 2 — Fase 1: porting backend jadi desktop-safe

**Commit:** `da3575f`, `8ffaa59`

Dua masalah dibereskan sebelum kode ini pantas disebut "aplikasi desktop":

- **Path & binding.** `tmpRoot` yang tadinya hardcoded `/tmp/pilot-diff-jobs`
  (asumsi server Linux) diganti `os.TempDir()`; folder riwayat pindah ke
  `os.UserConfigDir()`; dan `app.Listen(":8080")` yang tadinya "aman" karena
  Docker memetakan port ke `127.0.0.1`, diganti eksplisit
  `app.Listen("127.0.0.1:8080")` — sekarang tidak ada Docker yang jadi
  jaring pengaman, jadi backend sendiri yang harus menolak koneksi dari
  luar loopback.
- **Hapus login/auth.** Sesuai keputusan roadmap (1 laptop, 1 user, offline
  — ancaman "banyak orang akses server yang sama" sudah tidak relevan),
  seluruh `auth.go` dihapus: endpoint login/logout, middleware auth, layar
  login di frontend, token di localStorage. Tombol "Keluar" diganti
  "Reset" yang benar-benar mengembalikan state, bukan pura-pura logout.
  Divalidasi dengan `npm run build` lolos dan halaman tetap full static —
  syarat wajib supaya Fase 3 (static export) nggak mendadak gagal nanti.

## Tahap 3 — Fase 2+3: dari tab browser jadi jendela aplikasi beneran

**Commit:** `5e9c091`

Ini titik paling berisiko secara teknis. Tiga perubahan sekaligus:

1. `next.config.js` diset `output: 'export'` — frontend sekarang
   nge-compile jadi file statis (`frontend/out/`), bukan butuh Node server
   jalan terus.
2. `backend/embed.go` pakai `//go:embed all:webui` — hasil static export
   di-embed langsung ke dalam binary Go tunggal (diisi CI dari
   `frontend/out` sebelum `go build`, tidak dicommit ke git — lihat
   `.gitignore`).
3. `main.go` menyalakan Fiber di goroutine terpisah, menunggu port benar-benar
   terbuka (`waitForServer`, timeout 5 detik), baru membuka jendela
   `webview_go` (1280×800) yang navigate ke `localhost` — user melihat
   jendela aplikasi biasa dengan icon taskbar sendiri, bukan address bar
   browser yang bisa bikin curiga staf non-IT.

Catatan jujur dari commit message aslinya: perubahan ini **belum bisa
divalidasi penuh secara lokal** saat itu (`webview_go` butuh header
development GTK3/WebKit2GTK yang belum terpasang) — API call ke
`webview_go` dicocokkan manual ke source-nya di Go module cache, dan
build sungguhan menunggu hasil run CI di `windows-latest`. Ini contoh
nyata kenapa Tahap 1 (spike CI) dikerjakan lebih dulu: begitu CI sudah
terbukti bisa cross-build, fase berikutnya bisa dikerjakan dengan aman
meski belum tervalidasi 100% di mesin dev.

## Tahap 4 — Fase 4: installer, bukan exe polos

**Commit:** `01af741`, `7fe1866`

`installer/setup.iss` (Inno Setup) menghasilkan wizard Next-Next-Finish
dengan shortcut Desktop + Start Menu dan entry uninstall resmi di "Add or
Remove Programs" — sesuai keputusan roadmap (WiX/.msi ditolak karena
cocok buat deploy massal enterprise yang tidak relevan di sini). CI
menambahkan langkah `choco install innosetup` + compile `setup.iss` jadi
`ParseBankers-Setup.exe`, di-upload sebagai artifact utama.

Commit berikutnya (`7fe1866`) adalah contoh kecil disiplin scope: paket
`choco install innosetup` ternyata cuma bundle `English.isl` bawaan, bukan
`Indonesian.isl` yang tadinya diminta di `setup.iss` — bikin compile
installer gagal (`Couldn't open include file`). Daripada vendor-in file
translation Inno Setup segala, wizard installer dibiarkan default bahasa
Inggris (isi aplikasinya sendiri tetap Indonesia) — perbaikan paling
sederhana yang tersedia, bukan yang paling "lengkap".

## Tahap 5 — Setelah jalan: business rule berubah, UI dirombak, branding dipasang

**Commit:** `39ce129`, `6619edf`, `e6931fb`, `a03b6c4`, `8f3660d`, `1aa5eb7`

Begitu app bisa dibuild dan dipasang, iterasi berikutnya didorong feedback
pemakaian nyata:

- **Aturan bisnis berubah** (`39ce129`): definisi "Selisih Kurang"
  dipersempit jadi khusus baris ROLLBACK dengan nominal EJ=Cash sama
  persis; kategori "Selisih Lebih" dihapus total dari backend & frontend.
  Divalidasi bukan cuma `go test`, tapi disimulasikan ulang manual
  terhadap data real (`new-revisi/EJ.TXT` + file RC pasangannya) sampai
  hasil kategori match persis dengan implementasi asli.
- **Bug data nyata** (`e6931fb`): laporan "baris salah masuk tab Selisih
  Kurang" ditelusuri sampai akar masalahnya — file EJ mentah punya ~5.000
  baris yang terduplikasi persis byte-per-byte (kemungkinan dari export
  yang tanggalnya overlap). Dua fix sekaligus: parser dedup blok transaksi
  identik, dan React key tabel hasil diperbaiki supaya tidak collide kalau
  suatu saat memang ada `rec_num` kembar beneran.
- **Redesign UI** (`a03b6c4`): dua tampilan terpisah (Simple/Pro) digabung
  jadi satu layar modern-minimalist dengan warna korporat BNI. Fix teknis
  yang ketemu di jalan: CSS harus `<style jsx global>` bukan `<style jsx>`
  biasa, karena styled-jsx cuma scope ke komponen yang literal menulis
  tag `<style>`-nya — ketahuan lewat screenshot headless Chrome yang
  menunjukkan komponen anak tidak ter-style.
- **Branding resmi** (`8f3660d`): logo BNI (JPG background hitam solid)
  diproses lewat chroma-key jadi 3 aset transparan (header, favicon, icon
  exe), di-embed ke resource exe pakai `go-winres` tanpa mengubah kode Go
  sama sekali.
- **Icon jendela yang jalan masih generik** (`1aa5eb7`): `go-winres`
  ternyata cuma memperbaiki icon yang dibaca File Explorer/shortcut/
  installer — jendela yang sedang berjalan tetap pakai icon Windows
  generik, karena `webview_go` selalu membuat window class dengan
  `LoadImage(hInstance, IDI_APPLICATION, ...)`, tidak pernah membaca
  resource custom. Fix-nya kode Win32 API langsung
  (`ExtractIconW` + `WM_SETICON`) di file baru `window_icon_windows.go`,
  dengan stub kosong `window_icon_other.go` untuk non-Windows supaya
  tetap bisa dibuild dan dites di Linux.

## Tahap 6 — v1.1.0: stabilitas status & riwayat

**Commit:** `f2fdb73`, `1989a16`, `1343ceb`

Tiga bug atribusi status ROLLBACK dibereskan berurutan, masing-masing
dengan kasus nyata yang jadi bukti:

1. `f2fdb73` — parser sekarang flush & reset record tepat di
   `TRANSACTION END`, bukan menunggu `START` berikutnya, supaya baris
   sistem sesudah END tidak nyasar menandai transaksi sebelumnya. Sekalian
   riwayat dipindah ke penyimpanan permanen (`%AppData%`) supaya bisa
   dibuka lagi walau app sudah di-restart.
2. `1989a16` — status ROLLBACK dipersempit jadi cuma set kalau baris EJ
   eksplisit bilang "Rollback OK" (bukan sekadar mengandung kata
   "Rollback"), karena satu transaksi bisa punya dua "Rollback Notes" tapi
   cuma yang pertama diikuti "Rollback OK".
3. `1343ceb` — bug rec_num 407: percobaan deposit kedua di sesi kartu yang
   sama (gagal sebelum dapat nomor urut sendiri) numpuk rollback-nya ke
   transaksi pertama yang sukses. Parser sekarang flush record begitu
   ketemu sinyal percobaan baru (Amount kedua, atau PIN masuk lagi
   setelah percobaan sebelumnya sudah punya nomor urut).

## Tahap 7 — Saga lintas merek ATM: dari "cuma DN200V" jadi 4 merek

**Commit:** `15b94e1`, `5fee72b`, `4d5f7b6`, `fa8f187`, `8d7939a`

Ini rangkaian commit paling berat secara investigasi, karena parser yang
tadinya cuma dites terhadap satu merek mesin (DN200V) ternyata gagal total
begitu diuji ke tiga merek baru (Hyosung, Hitachi, OKI):

- `4d5f7b6` — anchor nomor urut transaksi (`TRAN SEQ NR [n]`) ternyata
  cuma ada di format DN200V; tiga merek lain nol kemunculan, jadi parser
  menolak file mereka mentah-mentah. Anchor dipindah ke `NO. REKORD n`
  yang ada di semua merek. Sekalian ketemu nominal setor tunai Hitachi &
  Hyosung yang terkirim 100× nilai asli (dibuktikan lewat hitungan uang
  masuk di log yang sama, bukan ditebak per merek).
- `fa8f187` — 100% baris Hyosung dan Hitachi keluar dengan `terminal_id`
  kosong. Akarnya: kedua merek menulis dua baris `TRANSACTION START`
  berturut-turut per transaksi, dan reset state di `START` kedua ikut
  menghapus Terminal ID yang baru saja terbaca di antara keduanya — padahal
  Terminal ID itu konstan per file (satu file = satu mesin fisik), bukan
  data per-transaksi yang seharusnya di-reset.
- `8d7939a` — yang paling kritis: 1.567 baris "Selisih Kurang" palsu
  senilai Rp1.780.900.000 di 7 mesin sample, sementara data client
  (`REVISI.docx`) bilang yang benar cuma 1 kasus per mesin. Dua akar
  masalah berbeda ditemukan (batas percobaan baru case-sensitive yang
  tidak pernah kena tulisan Hitachi/Hyosung; dan status "Shutter Opened"
  dipakai untuk dua kejadian berlawanan arah uang di OKI). Fix-nya
  universal — bukan percabangan per merek — dan diverifikasi ulang ke
  ketujuh mesin: OKI turun dari 1.356 kasus palsu ke 0.
- `15b94e1` dan `5fee72b` — perluasan cakupan status (3 EJ Status
  dianggap setara buat kategorisasi) dan fix bug UI turunannya (tombol
  Export mati total di WebView2 karena `<a target="_blank">` diblokir
  oleh WebView2, dan polling status yang bocor terus setelah tombol Stop
  ditekan).

Pola yang konsisten di seluruh tahap ini: setiap fix disertai bukti
kuantitatif sebelum/sesudah terhadap data mesin real (bukan cuma "test
lolos"), dan commit message mencatat berapa baris yang berubah kategori
serta kenapa.

## Tahap 8 — Polish terakhir

**Commit:** `af0e485`

Permintaan client: hilangkan jam dari kolom TANGGAL, sisakan tanggalnya
saja. Satu titik perubahan (`diff.go`, `split_part` di SQL DuckDB) karena
kolom tanggal untuk semua tampilan dan export sudah mengalir dari satu
sumber yang sama — bukti bahwa desain "satu sumber kebenaran" dari
tahap-tahap sebelumnya terbayar di sini: perubahan yang terdengar
kosmetik cuma butuh diubah di satu tempat, bukan dicari-cari di banyak
file.

---

**Cara verifikasi klaim di atas:** jalankan `git log --stat` di root repo
untuk melihat file yang berubah persis di tiap commit, atau `git show
<hash>` untuk lihat diff lengkapnya. Semua commit dari `15b94e1` dan
seterusnya juga mencantumkan `Co-Authored-By`/`Claude-Session` — sesi
kerja sungguhan yang bisa ditelusuri.
