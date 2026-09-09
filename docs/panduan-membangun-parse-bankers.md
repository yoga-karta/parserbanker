# Panduan Membangun Parse Bankers

---

## Parse Bankers

### Daftar Isi

> **Catatan istilah**
**EJ** (*Electronic Journal*) = catatan otomatis yang ditulis mesin ATM setiap ada aktivitas, berupa file teks panjang.
**RC** = file settlement kas dari sistem bank, satu baris satu transaksi.
**Rekonsiliasi** = proses mencocokkan kedua sumber itu untuk mencari transaksi yang tidak sinkron.
**Rec Num** (nomor rekord) = nomor urut transaksi di terminal ATM, dipakai sebagai kunci pencocokan.

---

### 1. Kebutuhan / Requirements

#### 1.1 Masalah yang diselesaikan

Setiap hari, mesin ATM menulis catatan otomatis (**EJ — Electronic Journal**) berisi seluruh
aktivitas mesin: kartu masuk, PIN dimasukkan, nominal diminta, uang keluar, uang ditarik balik,
kartu diambil, dan seterusnya. Bentuknya file teks mentah yang sangat panjang — satu file harian
dari satu terminal saja bisa lebih dari **25.000 baris**, dan satu transaksi tunggal tersebar di
puluhan baris berurutan.

Di sisi lain, sistem bank menerbitkan **file RC / settlement kas**: satu baris satu transaksi,
berisi nomor kartu, nomor rekord, nominal, dan tanda debit/kredit. Ini adalah "versi bank" dari
peristiwa yang sama.

Tugas staf cabang adalah memastikan **kedua versi itu cocok**. Kalau tidak cocok, artinya ada uang
yang perlu ditelusuri: mesin mencatat uang keluar tapi bank tidak membukukan, atau sebaliknya.
Sebelum aplikasi ini ada, pencocokan dikerjakan manual — membuka file teks raksasa,
mencari nomor rekord satu per satu, membandingkan nominal dengan mata. Konsekuensinya:

- **Lambat.** Ribuan baris transaksi dicocokkan satu-satu, memakan waktu berjam-jam setiap hari.
- **Rawan salah.** Kasus yang paling gampang terlewat justru yang paling berisiko: transaksi
yang sudah dicatat sukses di mesin, lalu uangnya ditarik balik (*rollback*) — polanya
tersembunyi di baris-baris kecil di tengah log.
- **Sulit ditelusuri ulang.** Hasil pencocokan kemarin tidak tersimpan rapi, jadi kalau ada
pertanyaan susulan harus mengulang dari awal.

**Parse Bankers** mengotomatiskan seluruh proses ini: staf memuat dua file, menekan satu tombol,
dan dalam hitungan detik mendapat tabel hasil yang sudah dikelompokkan menurut jenis masalahnya,
siap diekspor ke Excel sebagai lampiran laporan.

#### 1.2 Siapa penggunanya

| Jumlah pengguna | **1 orang** — staf cabang bank, bukan orang IT, tidak bisa dan tidak perlu bisa memprogram. |
|---|---|
| Perangkat | **1 laptop Windows 11** milik kantor cabang. Bukan server, bukan jaringan, bukan banyak komputer. |
| Koneksi internet | **Tidak ada / tidak dipakai.** Aplikasi harus jalan penuh dalam kondisi offline. |
| Dukungan IT di lokasi | **Tidak ada.** Tidak ada admin yang bisa memasang database, mengatur server, atau memperbaiki konfigurasi. Segalanya harus selesai dengan sekali klik installer. |
| Frekuensi pakai | Harian, mengikuti siklus settlement kas cabang. |

> **Konsekuensi desain dari profil pengguna ini**
Semua keputusan teknis di dokumen ini mengalir dari satu kalimat: satu orang non-IT,
satu laptop, tanpa internet, tanpa bantuan teknis. Setiap fitur yang tidak melayani kalimat itu
(login, multi-user, server, database terpisah, pembaruan otomatis) sengaja dihapus dari lingkup.

#### 1.3 Kebutuhan fungsional (apa yang harus bisa dilakukan aplikasi)

| No | Kebutuhan | Rincian |
|---|---|---|
| F-1 | Memuat file EJ | Pengguna memilih file log mentah ATM dari laptopnya. Aplikasi mengurainya menjadi daftar transaksi yang rapi, lalu menampilkan berapa transaksi yang berhasil dibaca. Kalau formatnya tidak dikenali, aplikasi berhenti dengan pesan yang jelas — bukan diam-diam menghasilkan nol baris. |
| F-2 | Memuat file RC | Pengguna memilih file settlement bank. Aplikasi menghitung berapa baris valid yang terbaca dan menampilkannya, tanpa langsung memproses. |
| F-3 | Pratinjau file | Sebelum diproses, isi awal file bisa diintip di layar supaya pengguna yakin tidak salah pilih file. |
| F-4 | Jalankan rekonsiliasi | Setelah kedua file dimuat, satu tombol menjalankan pencocokan penuh berdasarkan nomor rekord. Prosesnya berjalan di latar dan bisa **dihentikan** di tengah jalan lewat tombol STOP. |
| F-5 | Ringkasan hasil | Panel angka besar: total transaksi EJ, total baris RC, jumlah per kategori, total nominal EJ, total nominal Cash, dan total nominal yang masuk kategori Selisih Kurang. |
| F-6 | Tabel hasil + filter | Seluruh baris hasil ditampilkan dalam tabel bernomor halaman, dengan tombol filter per kategori (Selisih Kurang / Tidak Ditemukan / Match / Data Rusak). Tiap baris menampilkan nomor rekord, tanggal, terminal, no rekening, no transaksi, nominal EJ, nominal Cash, status EJ, status Cash, kategori, kemungkinan penyebab, dan keterangan. |
| F-7 | Ekspor hasil | Hasil bisa diunduh sebagai **Excel (.xlsx)** — dengan header tebal dan baris berwarna sesuai kategori — atau sebagai **teks (.txt)** berpembatas tanda `|`. Bisa mengekspor semua baris sekaligus atau hanya satu kategori. |
| F-8 | Riwayat | Setiap proses yang selesai dicatat permanen di laptop. Hasil lama tetap bisa dibuka kembali walau aplikasi sudah ditutup dan dibuka lagi. Ada tombol untuk membersihkan riwayat. |
| F-9 | Reset | Satu tombol membuang sesi berjalan beserta file sementaranya, siap mulai dari nol. |

#### 1.4 Kebutuhan non-fungsional dan alasan setiap kompromi

Bagian ini penting untuk dibaca oleh siapa pun yang akan memesan pembangunan ulang: beberapa hal
di bawah **sengaja tidak dibuat**. Itu keputusan sadar dengan alasan biaya dan lingkup, bukan
kelalaian. Kalau konteksnya berubah (misalnya nanti dipakai banyak cabang), keputusan-keputusan
ini yang pertama harus ditinjau ulang.

| Keputusan | Status | Alasan |
|---|---|---|
| Sepenuhnya offline | **Wajib** | Laptop cabang tidak diasumsikan punya internet, dan data transaksi nasabah tidak boleh keluar dari mesin. Seluruh pemrosesan — parsing, pencocokan, ekspor — terjadi di dalam laptop. Tidak ada satu pun panggilan ke server luar. |
| Satu pengguna | **Wajib** | Aplikasi berjalan sebagai program lokal yang hanya mendengarkan alamat internal komputer itu sendiri (`127.0.0.1`), sehingga tidak bisa diakses dari komputer lain di jaringan sekalipun. |
| Tanpa layar login | Sengaja dihapus | Versi awal aplikasi ini adalah web app di server yang diakses satu tim lewat jaringan — di sana login masuk akal. Setelah pindah jadi aplikasi desktop di satu laptop pribadi, ancaman yang jadi alasan login sudah tidak ada lagi: yang bisa membuka aplikasi adalah orang yang sudah bisa membuka laptopnya. Login hanya menyisakan satu layar tambahan yang harus dilewati tiap hari tanpa menambah keamanan nyata. Pengamanan sesungguhnya diserahkan ke **kunci layar Windows** laptop tersebut. |
| Tanpa pembaruan otomatis | Di luar lingkup | Mekanisme auto-update perlu server pembaruan, saluran rilis, dan penanganan gagal-update — pekerjaan besar demi **satu** laptop yang bahkan tidak selalu online. Selama masih satu laptop, mengirim installer baru secara manual setiap ada perbaikan sudah lebih murah dan lebih bisa diprediksi. Ini yang pertama perlu dipertimbangkan lagi kalau jumlah laptop bertambah. |
| Tanpa sertifikat
code-signing | Di luar lingkup | Sertifikat penandatanganan kode berbayar dan perlu proses verifikasi badan usaha. Untuk aplikasi internal yang tidak didistribusikan ke publik, biayanya tidak sebanding. **Konsekuensi yang harus diterima:** setiap kali installer baru dijalankan, Windows memunculkan layar biru peringatan "Windows protected your PC". Ini normal dan bisa dilewati (lihat Bagian 5.4). Karena isi file berubah tiap build, peringatan ini akan muncul lagi di setiap versi baru — tidak bisa "dihafalkan" oleh Windows. |
| Windows 11 minimum | **Wajib** | Jendela aplikasi dibangun di atas komponen WebView2 milik Microsoft. Di Windows 11 komponen ini sudah **bawaan sistem operasi**, jadi installer tidak perlu ikut membawa atau mengunduhnya — installer jadi ringan dan instalasi tetap bisa offline. Di Windows 10 komponen ini belum tentu ada, dan menanganinya berarti menambah unduhan online saat instalasi. |
| Kecepatan | Target | File EJ ~25.000 baris dan file RC ~1.000 baris harus selesai diproses dalam **hitungan detik**, bukan menit. |
| Instalasi | Target | Sekali jalankan installer, langsung bisa dipakai. Tidak ada pemasangan database, runtime, atau konfigurasi tambahan apa pun oleh pengguna. |

#### 1.5 Catatan sejarah yang wajib diketahui

Aturan kategorisasi aplikasi ini bukan hasil tebakan di atas kertas — ia berubah beberapa kali
setelah diuji dengan data transaksi asli. Dua pelajaran ini penting agar tidak terulang saat dibangun ulang:

> **Pelajaran 1 — kategori "Rollback" yang sempat dibuat lalu dibatalkan (Agustus 2026)**
Klien menyebut pola *"Nominal EJ dan Cash sama, EJ Status Rollback"*. Itu dibaca sebagai perintah
membuat **kategori baru** bernama "Rollback", dan sempat tayang beberapa jam sebelum ketahuan salah.
Maksud sebenarnya adalah **deskripsi pola baris yang memang seharusnya masuk Selisih Kurang**.
Kalau ada kalimat requirement yang ambigu antara "kriteria kategori baru" dan "penjelasan kategori lama",
tanyakan dulu sebelum dikerjakan.

> **Pelajaran 2 — urutan pemeriksaan kategori itu menentukan hasil**
Pernah terjadi 12 dari 13 baris "Selisih Kurang" ternyata palsu: baris EJ berstatus rollback yang
**sama sekali tidak punya pasangan** di file RC ikut terseret masuk, padahal tanpa nominal pembanding
tidak ada "selisih" apa pun untuk dihitung. Penyebabnya: pemeriksaan status rollback dijalankan
**sebelum** pemeriksaan "pasangan tidak ditemukan". Urutan pemeriksaan yang benar dikunci di
Bagian 5.3 dan wajib diikuti persis.

---

### 2. Diagram Alur Kerja Aplikasi

Diagram berikut menunjukkan perjalanan data yang sesungguhnya, dari dua file mentah di laptop
sampai menjadi tabel hasil dan file Excel. Semua kotak di dalam bingkai putus-putus berjalan
**di dalam satu file aplikasi (.exe)** di laptop pengguna — tidak ada server, tidak ada database
terpisah, tidak ada koneksi keluar.

#### 2.1 Membaca diagram dalam bahasa sehari-hari

1. **Dua file masuk lewat pintu yang berbeda.** File EJ butuh penerjemah khusus karena bentuknya
catatan bebas; file RC sudah rapi berkolom, jadi bisa langsung dibaca.
2. **Penerjemah EJ bekerja seperti pembaca yang teliti.** Ia membaca dari atas ke bawah dan
menandai kapan sebuah transaksi mulai dan berakhir. Selama di dalam satu transaksi, ia memungut
potongan informasi yang relevan (jam, nomor kartu, ID terminal, nominal, nomor urut, status).
Begitu transaksi berakhir, catatan itu ditulis sebagai satu baris rapi, lalu ingatannya dikosongkan
untuk transaksi berikutnya. **Pengosongan ingatan ini krusial** — tanpa itu, kejadian setelah
transaksi berakhir bisa salah tertempel ke transaksi sebelumnya.
3. **Kedua sisi bertemu di satu tabel.** Nomor rekord dipakai sebagai kunci. Jenis penggabungan
yang dipakai (*full outer join*) menjamin **tidak ada baris yang hilang**: transaksi yang hanya
ada di EJ tetap muncul, begitu juga yang hanya ada di RC.
4. **Setiap baris hasil gabungan dinilai bertingkat.** Aplikasi memeriksa syarat satu per satu dari
atas ke bawah dan berhenti di syarat pertama yang terpenuhi — persis seperti alur "kalau ini,
maka itu; kalau bukan, lanjut cek berikutnya". Urutannya tidak boleh diacak.
5. **Hasilnya bisa dilihat, disaring, dan diunduh.** Semua tetap di dalam laptop.

---

### 3. Prompt AI

Teks di dalam kotak oranye di bawah ini adalah **satu instruksi utuh yang bisa disalin apa adanya**
dan diberikan kepada asisten AI pemrogram (atau developer manusia) untuk membangun ulang aplikasi
ini dari nol. Isinya sengaja dibuat sangat spesifik — menyebut nama teknologi, aturan penguraian
file, dan aturan kategorisasi secara persis — supaya hasilnya bukan aplikasi rekonsiliasi
"kira-kira mirip", tetapi aplikasi yang sama.

> **Cara memakainya**
Salin seluruh isi kotak (tiga paragraf), tempelkan sebagai satu pesan. Sertakan juga **satu contoh
file EJ.TXT dan satu contoh file RC-file.txt yang asli** — contoh nyata jauh lebih meyakinkan
daripada deskripsi. Bagian 5.1 dan 5.2 dokumen ini bisa dilampirkan sebagai rujukan format.

> Prompt siap salin
Buatkan saya aplikasi desktop Windows bernama **“Parse Bankers”** untuk merekonsiliasi log
mesin ATM dengan file settlement kas bank. Aplikasi dipakai satu orang staf bank non-teknis di satu
laptop Windows 11 offline, tanpa layar login, tanpa koneksi internet sama sekali. Tumpukan
teknologinya harus persis begini: backend **Go 1.22** dengan framework web **Fiber v2** yang
hanya mendengarkan di `127.0.0.1:8080`; mesin data **DuckDB tertanam** lewat pustaka
`github.com/marcboeker/go-duckdb` (butuh CGO); ekspor Excel dengan
`github.com/xuri/excelize/v2`; frontend **Next.js** satu halaman, murni sisi klien
(tanpa SSR dan tanpa API route), diekspor statis dengan `output: "export"` lalu ditanam ke
dalam biner Go memakai `//go:embed` dan disajikan oleh Fiber; jendela aplikasi native
memakai **WebView2** lewat `github.com/webview/webview_go` (bukan Electron, bukan Tauri)
berjudul “Parse Bankers - Reconciliation Portal” ukuran 1280×800, yang dibuka hanya
setelah server lokal siap menerima koneksi. Seluruhnya harus menjadi **satu file .exe**, dibundel
dengan installer **Inno Setup** yang memasang shortcut Desktop + Start Menu dan entri uninstall,
dan dibangun otomatis oleh **GitHub Actions runner windows-latest** (karena CGO tidak praktis
di-cross-compile dari Linux) dengan mingw-w64 sebagai kompiler C. Tanpa code-signing. File sementara
simpan di `os.TempDir()`, sedangkan riwayat dan hasil yang harus awet simpan di folder
konfigurasi pengguna (`os.UserConfigDir()`) sebagai `history.jsonl` dan
`results/<job-id>/result.duckdb`.
Alur kerjanya: pengguna memuat **file EJ** (log mentah ATM), lalu **file RC** (settlement bank),
lalu menekan tombol proses. **Parser EJ** harus berupa pembaca baris-per-baris dengan mesin status:
blok transaksi dipotong pada baris yang memuat `TRANSACTION START` (dua potongan pertama
baris itu, dipisah spasi, adalah tanggal dan jam — simpan sebagai timestamp) dan ditutup pada baris
yang memuat `TRANSACTION END`; pada **kedua** penanda itu catatan berjalan ditulis ke
keluaran lalu **direset** — reset di TRANSACTION END wajib, karena baris mesin seperti
“Rollback OK” kadang muncul setelah transaksi ditutup tapi sebelum transaksi berikutnya
dimulai, dan tanpa reset baris itu salah tertempel ke transaksi yang baru saja selesai. Di dalam blok,
ambil dengan pola pencocokan teks: nomor kartu dari `CARD NUMBER <spasi> <digit dan
bintang>`, terminal dari `Terminal ID [ ... ]`, nominal dari `Amount :
<angka>`, nomor urut dari `TRAN SEQ NR [ <angka> ]`. Status diisi
`SUCCESS` bila ada baris `TRANSACTION REPLIED`; `FAILED` bila ada
`TRANSACTION FAILED` atau `TRANSACTION DECLINED`; dan tiga status khusus yang
harus disimpan apa adanya bila barisnya menyebut persis teks tersebut:
`ROLLBACK OK`, `ROLLBACK NOTES SUCCESSFULLY`, dan
`SHUTTER OPENED FOR NOTES REMOVAL` — pencocokannya harus tepat frasa itu, jangan hanya
kata “Rollback” atau “Shutter”, karena baris seperti
`----- Rollback Notes ----` adalah housekeeping biasa dan bukan rollback sungguhan. Satu blok
bisa berisi lebih dari satu percobaan transaksi: **tulis dan reset catatan berjalan** setiap kali
muncul `Amount :` kedua sementara nominal sebelumnya belum kosong, dan setiap kali muncul
`PIN ENTERED` sementara nomor urut sebelumnya sudah terisi — tanpa ini, status rollback
milik percobaan kedua akan salah menempel ke percobaan pertama yang sebenarnya sukses. Baris hanya
ditulis kalau nomor urut dan nominalnya terisi, dan blok dengan kombinasi timestamp + nomor urut +
nominal yang identik dianggap duplikat ekspor dan hanya dihitung sekali. Hasilnya CSV enam kolom:
`timestamp, card_masked, terminal_id, amount, seq_nr, status`. **File RC** dibaca langsung
oleh DuckDB `read_csv` dengan `delim=';'`, `all_varchar=true`,
`quote=''`, `ignore_errors=true`, `null_padding=true`; nomor rekening
diambil dari `split_part(column0,'/',1)`, nomor rekord dari
`split_part(column1,'/',2)`, nomor transaksi dari `split_part(column2,'/',5)`,
nominal dari `column3`, tanda debit/kredit dari `column4`, dan tanggal dari
`column6`; baris yang nomor rekordnya tidak bisa dikonversi ke bilangan bulat harus dibuang.
Pencocokan dilakukan di DuckDB dengan `FULL OUTER JOIN` antara tabel EJ dan tabel RC pada
`rec_num`, lalu setiap baris hasil diberi kategori lewat pemeriksaan bertingkat dengan
**urutan yang tidak boleh diubah**: (1) kalau `rec_num` di salah satu sisi kosong →
**tidak_ditemukan**; (2) selain itu, kalau nominal EJ atau nominal Cash gagal dikonversi jadi angka
→ **data_invalid**; (3) selain itu, kalau nominal Cash sama persis dengan nominal EJ **dan**
status EJ termasuk salah satu dari tiga nilai `'ROLLBACK OK'`,
`'ROLLBACK NOTES SUCCESSFULLY'`, `'SHUTTER OPENED FOR NOTES REMOVAL'` →
**selisih_kurang**, dengan keterangan “Transaksi rollback di EJ - dana kemungkinan sudah
keluar” dan kemungkinan penyebab “Nasabah Diuntungkan”; (4) selain itu, kalau nominal
Cash sama dengan nominal EJ → **match**; (5) sisanya (nominal berbeda) →
**tidak_ditemukan** dengan keterangan “Nominal EJ dan Cash tidak sama”. Tidak ada kategori
“Selisih Lebih” dan tidak ada kategori “Rollback” terpisah. Tampilannya satu
halaman: dua kotak pemilih berkas dengan tombol pratinjau dan jumlah record hasil pembacaan, tombol
proses dan tombol berhenti, panel ringkasan berisi total EJ, total Cash, jumlah tiap kategori, total
nominal EJ, total nominal Cash, dan total nominal Selisih Kurang, lalu tabel hasil berhalaman dengan
tombol saring per kategori berwarna (Selisih Kurang merah, Tidak Ditemukan ungu, Match hijau, Data Rusak
kuning) dan kolom: rec num, tanggal, terminal, no rekening, no transaksi, nominal EJ, nominal Cash, EJ
status, cash status, hasil, kemungkinan penyebab, keterangan. Sediakan ekspor **.xlsx** (header tebal,
warna latar baris mengikuti kategori) dan **.txt** berpembatas `|`, untuk semua kategori
atau satu kategori saja, dengan daftar-putih nama kategori yang boleh masuk kueri. Simpan riwayat tiap
proses ke berkas JSON-per-baris supaya hasil lama tetap bisa dibuka setelah aplikasi ditutup, dan buat
status pekerjaan bisa dibangun ulang dari berkas hasil di disk kalau aplikasi baru saja dijalankan
ulang. Semua teks antarmuka dan pesan kesalahan dalam **Bahasa Indonesia** yang mudah dipahami orang
non-IT. Terakhir, tuliskan pengujian otomatis untuk parser EJ dan untuk logika kategorisasi —
minimal kasus rollback setelah TRANSACTION END, percobaan kedua dalam satu sesi kartu, blok duplikat, dan
kelima cabang kategori di atas.

#### 3.1 Kenapa prompt ini panjang dan sangat rinci

Instruksi pendek seperti "buatkan aplikasi rekonsiliasi ATM" akan menghasilkan aplikasi yang tampak benar
tetapi salah di detail yang justru menentukan nilai uang. Tiga hal berikut adalah yang paling sering
salah dan karena itu ditulis eksplisit di prompt:

- **Reset catatan di TRANSACTION END.** Kalau ini terlewat, transaksi yang sebenarnya bersih akan
salah ditandai rollback — artinya laporan menuduh ada uang bermasalah padahal tidak.
- **Urutan pemeriksaan kategori.** Kalau pemeriksaan rollback didahulukan, baris tanpa pasangan
akan membanjiri kategori Selisih Kurang (kejadian nyata: 12 dari 13 baris palsu).
- **Pencocokan frasa status yang harus tepat.** Mencocokkan hanya kata "Rollback" akan menangkap
baris housekeeping `----- Rollback Notes ----` yang muncul di hampir setiap transaksi.

---

### 4. Tech Stack

Semua pilihan di bawah tunduk pada satu batasan: hasil akhirnya harus **satu file yang bisa dipasang
dan langsung jalan di laptop Windows 11 offline, tanpa bantuan IT**.

#### 4.1 Daftar teknologi

| Lapisan | Pilihan | Peran dan alasan |
|---|---|---|
| Bahasa & logika inti | **Go 1.22** | Menghasilkan satu berkas program mandiri tanpa perlu memasang runtime apa pun di laptop pengguna. Ini alasan utamanya — bahasa yang butuh runtime terpisah (Java, .NET, Python, Node) akan menambah langkah instalasi yang tidak bisa dilakukan pengguna non-IT. |
| Server web internal | **Fiber v2** | Menyediakan alamat internal (`/api/...`) yang dipanggil oleh tampilan, sekaligus menyajikan halaman tampilan itu sendiri. Hanya mendengarkan `127.0.0.1` sehingga tidak terlihat dari jaringan. |
| Mesin pencocokan data | **DuckDB** via `go-duckdb` (butuh CGO) | Database analitik yang **ikut tertanam di dalam aplikasi** — tidak perlu memasang server database terpisah. Ia bisa membaca file CSV langsung dan mengerjakan penggabungan ribuan baris dalam hitungan milidetik. Hasil tiap proses disimpan sebagai satu berkas `result.duckdb` yang bisa dibuka lagi kapan saja. |
| Ekspor Excel | **excelize v2** | Menulis berkas `.xlsx` asli lengkap dengan header tebal dan pewarnaan baris per kategori, tanpa perlu Microsoft Excel terpasang di mesin yang membuatnya. |
| Tampilan | **Next.js** (React), ekspor statis | Satu halaman, seluruhnya berjalan di sisi klien. Dengan `output: "export"` ia menjadi kumpulan berkas HTML/CSS/JS biasa — tidak butuh Node.js di laptop pengguna. |
| Penyatuan | `//go:embed` | Hasil ekspor tampilan dimasukkan **ke dalam** berkas program Go, sehingga distribusinya benar-benar satu berkas, bukan satu berkas plus folder aset yang bisa hilang atau salah letak. |
| Jendela aplikasi | **WebView2** via `webview_go` | Membuka tampilan sebagai **jendela aplikasi biasa** dengan judul dan ikon sendiri di taskbar — bukan tab browser. Lihat alasan lengkap di 4.2. |
| Installer | **Inno Setup** | Wizard "Next – Next – Finish" yang familiar, memasang shortcut Desktop dan Start Menu serta entri resmi di "Add or Remove Programs" untuk uninstall. |
| Proses build | **GitHub Actions**, runner `windows-latest` | Kompilasi dilakukan di mesin Windows sungguhan. Lihat 4.3. |
| Penandatanganan kode | **Tidak ada** | Keputusan biaya; konsekuensinya SmartScreen (lihat Bagian 5.4). |
| Sistem operasi | **Windows 11 64-bit** | WebView2 sudah bawaan sistem, jadi installer tidak perlu mengunduh apa pun saat dipasang. |

#### 4.2 Kenapa WebView2, bukan Electron atau Tauri

Tampilan aplikasi ini adalah halaman web yang berjalan lokal. Ada tiga cara lazim menampilkannya
sebagai aplikasi desktop:

| Opsi | Keputusan | Pertimbangan |
|---|---|---|
| Buka di browser
(`localhost:8080`) | Ditolak | Pengguna non-IT akan melihat address bar berisi alamat teknis dan bisa bingung atau curiga aplikasinya "mengirim data ke internet". Juga mudah tertutup tanpa sengaja bersama tab lain. |
| Electron | Ditolak | Membungkus salinan lengkap mesin browser Chromium **plus** runtime Node.js ke dalam paket aplikasi. Ukurannya berlipat ratusan megabita dan menambah satu ekosistem alat build lagi — terlalu berat untuk aplikasi satu halaman seperti ini. |
| Tauri | Ditolak | Secara teknis tepat dan ringan, tetapi menuntut pemasangan rantai alat bahasa Rust secara penuh hanya untuk membuat bingkai jendela — menambah satu bahasa dan satu rantai build lagi ke proyek yang sudah punya Go. |
| WebView2 | Dipakai | Memakai mesin browser yang **sudah ada di dalam Windows 11**. Tidak ada yang perlu dibundel, tidak ada alat build tambahan, dan hasilnya tetap jendela aplikasi biasa dengan judul dan ikon sendiri. Satu-satunya berkas pendamping adalah `WebView2Loader.dll` kecil yang ikut dipasang installer. |

#### 4.3 Kenapa build harus dilakukan di mesin Windows

Pustaka DuckDB (dan juga WebView2) sesungguhnya ditulis dalam bahasa C/C++. Agar bisa dipakai dari Go,
ia dijembatani lewat mekanisme bernama **CGO**. Akibatnya, membuat versi Windows dari komputer Linux
tidak bisa dilakukan dengan perintah sederhana — ia butuh rantai kompilasi C lengkap untuk target
Windows, yang rapuh dan sulit dirawat.

Solusinya: kompilasi dijalankan langsung **di mesin Windows asli** yang disediakan gratis oleh
GitHub Actions (runner `windows-latest`). Alurnya otomatis setiap ada perubahan kode:

1. Pasang Go 1.22, Node.js 22, dan kompiler C **mingw-w64**.
2. Bangun tampilan Next.js jadi berkas statis, salin ke folder yang akan ditanam ke program.
3. Tanamkan ikon aplikasi ke dalam berkas `.exe`.
4. Kompilasi backend dengan CGO aktif menjadi satu `.exe`.
5. Ambil `WebView2Loader.dll` resmi dari paket Microsoft.
6. Bangun installer dengan Inno Setup, lalu unggah dua hasil: **installer siap kirim** dan
**exe mentah** (untuk pengujian cepat tanpa instalasi).

> **Konsekuensi praktis yang perlu diketahui pemesan**
Karena kompilasi bergantung pada layanan build online, **proses pembuatan versi baru butuh internet**
— walaupun **aplikasi hasilnya** tidak butuh internet sama sekali. Yang offline adalah
pemakaiannya, bukan pembuatannya.

---

### 5. Dan Lain-lain

#### 5.1 Spesifikasi format file EJ (log mesin ATM)

File EJ adalah teks biasa. Setiap baris diawali tanggal dan jam, diikuti keterangan aktivitas.
Satu transaksi terbentang dari baris `TRANSACTION START` sampai `TRANSACTION END`,
dan di antaranya bisa terdapat puluhan baris teknis mesin, termasuk salinan struk yang dicetak.
Berikut **potongan asli** dari berkas `EJ.TXT` milik terminal `S1GBMSR055`:

Dan berikut potongan asli sebuah transaksi yang **uangnya ditarik kembali oleh mesin** — pola inilah yang menjadi inti kategori Selisih Kurang:

Yang diambil aplikasi dari setiap blok transaksi:

| Kolom hasil | Diambil dari baris | Contoh nilai |
|---|---|---|
| timestamp | Dua potongan pertama baris `TRANSACTION START` | `08/07/2026 00:40:04` |
| card_masked | `CARD NUMBER <digit/bintang>` | `526422******2731` |
| terminal_id | `Terminal ID [ ... ]` | `S1GBMSR055` |
| amount | `Amount : <angka>` | `30000` |
| seq_nr | `TRAN SEQ NR [ <angka> ]` → jadi kunci **rec_num** | `0228` → `228` |
| status | Baris penanda hasil (lihat di bawah) | `SUCCESS` / `ROLLBACK OK` |

Nilai status yang dikenali, berikut baris pemicunya:

| Nilai status | Baris pemicu di log |
|---|---|
| SUCCESS | baris memuat `TRANSACTION REPLIED` |
| FAILED | baris memuat `TRANSACTION FAILED` atau `TRANSACTION DECLINED` |
| ROLLBACK OK | baris memuat persis frasa `Rollback OK` |
| ROLLBACK NOTES SUCCESSFULLY | baris memuat persis frasa `Rollback Notes Successfully` |
| SHUTTER OPENED FOR NOTES REMOVAL | baris memuat persis frasa `Shutter Opened for notes removal` |

> **Jebakan yang harus dihindari**
Baris `----- Rollback Notes ----` muncul di banyak transaksi sebagai catatan rutin mesin dan
**bukan** penanda rollback. Hanya baris yang menyebut frasa lengkap di tabel di atas yang boleh
mengubah status. Selain itu baris rollback kadang muncul **setelah** `TRANSACTION END` —
karena itu catatan berjalan wajib direset di penanda END, bukan hanya di START berikutnya.

#### 5.2 Spesifikasi format file RC (settlement kas bank)

File RC jauh lebih sederhana: satu baris satu transaksi, kolom dipisah titik koma (`;`).
Di dalam beberapa kolom masih ada sub-bagian yang dipisah garis miring (`/`).
Berikut **dua baris asli** dari `RC-file.txt`:

Pembagian kolomnya (dihitung mulai dari nol seperti di dalam aplikasi):

| Kolom | Isi contoh | Dipakai sebagai | Cara pengambilan |
|---|---|---|---|
| column0 | `5198930157870667/BNI` | No rekening / kartu | potongan ke-1 dipisah `/` |
| column1 | `S1GBMSR055/233/616365` | **rec_num** — kunci pencocokan | potongan ke-2 dipisah `/` → `233` |
| column2 | `0210/S1GBMSR055/233 /…/0487794394/VL` | No transaksi | potongan ke-5 dipisah `/` |
| column3 | `1700000` | Nominal Cash | langsung, dikonversi ke angka |
| column4 | `D` atau `K` | Cash Status (debit/kredit) | langsung |
| column5 | `326600000` | tidak dipakai (saldo berjalan) | — |
| column6 | `07/08/26` | Tanggal transaksi | langsung |

Baris yang nomor rekordnya tidak bisa dibaca sebagai angka (misalnya baris header atau baris rusak)
**dibuang secara diam-diam**, bukan menggagalkan pembacaan seluruh file. Kalau setelah penyaringan
tidak tersisa satu baris pun, aplikasi menolak file itu dengan pesan agar pengguna memeriksa formatnya.

> **Keanehan yang diketahui pada column2**
Untuk baris setoran (kode `VL`), potongan ke-5 berisi nomor transaksi asli seperti
`0487794394`. Tetapi untuk baris penarikan (kode `CW`/`VK`) susunan
potongannya lebih pendek, sehingga potongan ke-5 justru berisi kode jenis transaksi seperti
`CW`. Kolom "No Transaksi" pada baris jenis ini karena itu tampil sebagai kode, bukan nomor.
Ini **tidak memengaruhi hasil pencocokan** (yang memakai rec_num dan nominal), hanya memengaruhi
tampilan satu kolom informasi.

#### 5.3 Daftar lengkap aturan kategorisasi

Setiap baris hasil penggabungan diperiksa **berurutan dari atas ke bawah** dan berhenti pada aturan
pertama yang cocok. Urutan ini adalah bagian dari spesifikasi, bukan detail teknis yang boleh diubah.

| Urut | Kategori | Syarat | Arti bagi staf |
|---|---|---|---|
| **1** | Tidak Ditemukan | Nomor rekord hanya ada di satu sisi — tidak ada pasangannya di file yang lain. | Keterangan: "Tidak ada di EJ" atau "Tidak ada di RC". Transaksi tercatat di satu sistem tapi tidak di sistem lainnya — perlu ditelusuri manual. |
| **2** | Data Rusak | Pasangan ditemukan, tetapi nominal EJ atau nominal Cash **tidak bisa dibaca sebagai angka** (kosong atau rusak). | Keterangan: "Nominal tidak terbaca (data rusak/kosong)". Bukan masalah uang, melainkan masalah kualitas data — biasanya file terpotong atau ada baris cacat. |
| **3** | Selisih Kurang | **Nominal EJ sama persis dengan nominal Cash**, **dan** status EJ adalah salah satu dari tiga nilai: `ROLLBACK OK`, `ROLLBACK NOTES SUCCESSFULLY`, `SHUTTER OPENED FOR NOTES REMOVAL`. | Keterangan: "Transaksi rollback di EJ - dana kemungkinan sudah keluar". Kemungkinan penyebab: **"Nasabah Diuntungkan"**. Ini kategori paling penting: mesin mencatat menarik kembali uang, tetapi bank tetap membukukan nominal yang sama — ada indikasi uang sudah keluar tanpa terkoreksi. |
| **4** | Match / Klop | Nominal EJ sama dengan nominal Cash, dan status EJ **bukan** salah satu dari tiga status di atas. | Keterangan: "-". Aman, tidak perlu tindakan. |
| **5** | Tidak Ditemukan | Sisanya — pasangan ada, nominal terbaca, tetapi **nominal EJ dan Cash berbeda**. | Keterangan: "Nominal EJ dan Cash tidak sama". Sengaja tidak dianggap cocok otomatis: perbedaan nominal selalu perlu ditinjau manusia. |

> **Riwayat perubahan aturan Selisih Kurang — penting bagi yang akan membangun ulang**
Definisi kategori ini berubah tiga kali. **Versi awal:** generik — setiap baris yang nominal
Cash-nya lebih kecil dari nominal EJ, ditambah kategori pendamping "Selisih Lebih" untuk kebalikannya.
**Perubahan pertama (1 Sep 2026):** disempitkan drastis — hanya baris berstatus rollback
**dan** nominal EJ–Cash **sama**; kategori "Selisih Lebih" dihapus total; baris yang nominalnya
berbeda dipindah ke Tidak Ditemukan karena dianggap butuh tinjauan manual. **Perubahan kedua
(3 Sep 2026, berlaku saat ini):** pemicunya diperluas dari satu status menjadi **tiga status EJ**
seperti tertulis di baris 3 tabel di atas — ketiganya sama-sama berarti "uang sempat/sudah keluar
dari mesin". Status lain (misalnya `SUCCESS`) tetap dikecualikan dan jatuh ke Match bila
nominalnya sama.

#### 5.4 Cara instal

1. Salin berkas **`ParseBankers-Setup.exe`** ke laptop (lewat flashdisk atau jaringan internal).
2. Klik dua kali berkas tersebut.
3. **Windows akan menampilkan layar biru bertuliskan "Windows protected your PC".** Ini **normal**
dan bukan tanda aplikasi bermasalah — muncul karena aplikasi internal ini tidak didaftarkan berbayar
ke Microsoft (lihat Bagian 1.4). Cara melewatinya:

klik tulisan kecil “More info” → lalu klik tombol “Run anyway”.
4. Ikuti wizard: **Next → Next → Install → Finish**.
5. Selesai. Aplikasi terpasang di folder Program Files, dan otomatis muncul:
ikon **Parse Bankers** di Desktop,
entri **Parse Bankers** di Start Menu,
entri uninstall resmi di **Settings → Apps → Installed apps**.

> **Peringatan SmartScreen akan muncul lagi di setiap versi baru**
Windows menilai berkas berdasarkan isinya. Karena setiap versi baru punya isi yang berbeda, ia akan
dianggap berkas asing lagi. Jadi setiap kali menerima installer baru, langkah "More info → Run anyway"
perlu diulang. Selama tidak ada sertifikat code-signing, ini tidak bisa dihilangkan.

#### 5.5 Cara pakai sehari-hari

| Langkah | Yang dilakukan | Yang terjadi / yang terlihat |
|---|---|---|
| 1 | Buka **Parse Bankers** dari ikon Desktop | Jendela aplikasi terbuka langsung ke halaman kerja. Tidak ada layar login. |
| 2 | Pada kotak **Electronic Journal**, pilih berkas EJ hari itu | Aplikasi menguraikan log dan menampilkan **jumlah transaksi** yang berhasil dibaca. Tombol pratinjau bisa dipakai untuk memastikan berkasnya benar. |
| 3 | Pada kotak **Cash / Reconciliation**, pilih berkas RC hari itu | Aplikasi menampilkan **jumlah baris valid** yang terbaca. |
| 4 | Tekan tombol **proses** | Pencocokan berjalan (hitungan detik). Tombol **STOP** tersedia bila perlu dibatalkan. |
| 5 | Baca panel **ringkasan** | Total EJ, total Cash, jumlah per kategori, total nominal EJ, total nominal Cash, dan total nominal Selisih Kurang. |
| 6 | Klik tombol saring kategori | Tabel hanya menampilkan baris kategori tersebut. Biasanya yang pertama diperiksa adalah Selisih Kurang, lalu Tidak Ditemukan. |
| 7 | Tekan **Export** | Berkas `hasil_rekonsiliasi_<kategori>.xlsx` atau `.txt` tersimpan — siap dilampirkan ke laporan. Bisa mengekspor semua kategori sekaligus atau satu saja. |
| 8 | Bila perlu mulai lagi | Tombol **Reset** membersihkan sesi. Hasil lama tetap bisa dibuka lewat panel **Riwayat**, bahkan setelah aplikasi ditutup dan dibuka kembali. |

#### 5.6 Batasan yang diketahui

| Batasan | Penjelasan dan jalan keluarnya bila nanti dibutuhkan |
|---|---|
| Hanya untuk satu laptop | Tidak ada mekanisme berbagi hasil antar-komputer. Riwayat dan hasil tersimpan hanya di laptop tempat aplikasi dipasang. Kalau nanti dibutuhkan lintas cabang, aplikasi harus dikembalikan ke bentuk layanan terpusat — dan saat itu layar login perlu dihidupkan lagi. |
| Tanpa pembaruan otomatis | Setiap perbaikan berarti **installer baru dikirim manual** dan dipasang ulang oleh pengguna, lengkap dengan langkah melewati SmartScreen. Untuk satu laptop ini masih wajar; untuk banyak laptop tidak. |
| Peringatan SmartScreen selalu muncul | Konsekuensi langsung dari tidak adanya sertifikat code-signing. Sertakan selalu satu kalimat instruksi setiap kali mengirim installer baru. |
| Parser terikat format log ATM tertentu | Aturan penguraian disusun dari log terminal `S1GBMSR055`. Mesin ATM merek atau versi firmware lain bisa memakai penanda baris yang berbeda — parser perlu disesuaikan dan diuji ulang dengan berkas asli dari mesin tersebut. |
| Folder hasil menumpuk | Hasil setiap proses disimpan permanen dan belum ada pembersihan otomatis. Setelah pemakaian panjang, folder ini akan tumbuh — sementara ini bisa dibersihkan lewat tombol pembersih riwayat. |
| Nomor port tetap | Aplikasi memakai alamat internal tetap `127.0.0.1:8080`. Bila kebetulan ada program lain di laptop yang sudah memakai nomor itu, aplikasi gagal dijalankan. Belum ada pemilihan port otomatis. |
| Nomor versi ditulis di dua tempat | Versi aplikasi tercatat di kode program **dan** di berkas installer, dan harus disamakan manual setiap rilis. Kalau frekuensi rilis meningkat, sebaiknya disatukan ke satu sumber. |
| Kolom "No Transaksi" pada baris penarikan | Menampilkan kode jenis transaksi, bukan nomor — lihat catatan di Bagian 5.2. Tidak memengaruhi hasil pencocokan. |

#### 5.7 Ringkasan satu halaman

> Kalau hanya satu paragraf yang sempat dibaca
Parse Bankers adalah aplikasi Windows satu berkas yang membaca **log mentah mesin ATM** dan
**file settlement kas bank**, mencocokkannya berdasarkan **nomor rekord**, lalu mengelompokkan
setiap transaksi ke dalam empat kategori: Match
Selisih Kurang Tidak Ditemukan
Data Rusak — dengan **Selisih Kurang** sebagai temuan
terpenting, yaitu transaksi yang di mesin ditandai uangnya ditarik kembali (tiga status EJ:
*Rollback OK*, *Rollback Notes Successfully*, *Shutter Opened for notes removal*)
tetapi tetap dibukukan bank dengan nominal yang sama persis. Semuanya berjalan offline di satu
laptop, dan hasilnya bisa diekspor ke Excel untuk dilampirkan ke laporan harian.

Dokumen ini disusun dari kode sumber aplikasi versi 1.1.0, berkas data uji asli
(`EJ.TXT`, `RC-file.txt`), catatan roadmap desktop Windows, catatan keputusan
kategori rollback, dan riwayat pengembangan proyek.

