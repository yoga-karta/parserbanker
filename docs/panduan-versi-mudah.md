# Parse Bankers — Panduan Versi Mudah Dipahami

---

## Parse Bankers

### Daftar Isi

> **Catatan istilah, biar tidak bingung**
**EJ** (*Electronic Journal*) = catatan otomatis yang ditulis mesin ATM setiap ada aktivitas.
**RC** = catatan settlement kas dari sistem bank. **Rekonsiliasi** = proses mencocokkan kedua
catatan itu untuk mencari transaksi yang tidak sinkron. **Rec Num** (nomor rekord) = nomor urut
transaksi yang dipakai aplikasi sebagai kunci untuk mencocokkan kedua catatan.

---

### 1. Kebutuhan

Bagian ini adalah satu-satunya bagian yang sengaja ditulis lebih rinci, karena isinya adalah
keputusan resmi yang mendasari seluruh aplikasi. Bagian 2 dan seterusnya jauh lebih ringan.

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
yang perlu ditelusuri. Sebelum aplikasi ini ada, pencocokan dikerjakan manual — membuka file teks
raksasa, mencari nomor rekord satu per satu, membandingkan nominal dengan mata. Konsekuensinya:

- **Lambat.** Ribuan baris transaksi dicocokkan satu-satu, memakan waktu berjam-jam setiap hari.
- **Rawan salah.** Kasus yang paling gampang terlewat justru yang paling berisiko: transaksi
yang sudah dicatat sukses di mesin, lalu uangnya ditarik balik (*rollback*) — polanya
tersembunyi di baris-baris kecil di tengah log.
- **Sulit ditelusuri ulang.** Hasil pencocokan kemarin tidak tersimpan rapi, jadi kalau ada
pertanyaan susulan harus mengulang dari awal.

**Parse Bankers** mengotomatiskan seluruh proses ini: staf memuat dua file, menekan satu tombol,
dan dalam hitungan detik mendapat tabel hasil yang sudah dikelompokkan menurut jenis masalahnya,
siap diekspor ke Excel sebagai lampiran laporan. Aplikasi ini berjalan **sepenuhnya di dalam laptop**
— tidak butuh koneksi internet sama sekali, baik saat dipasang maupun saat dipakai sehari-hari.

#### 1.2 Siapa penggunanya

| Jumlah pengguna | **1 orang** — staf cabang bank, bukan orang IT, tidak bisa dan tidak perlu bisa memprogram. |
|---|---|
| Perangkat | **1 laptop Windows 11** milik kantor cabang. Bukan server, bukan jaringan, bukan banyak komputer. |
| Koneksi internet | **Tidak ada / tidak dipakai.** Aplikasi harus jalan penuh dalam kondisi offline. |
| Dukungan IT di lokasi | **Tidak ada.** Tidak ada admin yang bisa memasang database, mengatur server, atau memperbaiki konfigurasi. Segalanya harus selesai dengan sekali klik installer. |
| Frekuensi pakai | Harian, mengikuti siklus settlement kas cabang. |

> **Konsekuensi desain dari profil pengguna ini**
Semua keputusan di dokumen ini mengalir dari satu kalimat: satu orang non-IT,
satu laptop, tanpa internet, tanpa bantuan teknis. Setiap fitur yang tidak melayani kalimat itu
(login, multi-user, server, database terpisah, pembaruan otomatis) sengaja dihapus dari lingkup.

#### 1.3 Yang bisa dilakukan aplikasi

| No | Kebutuhan | Rincian |
|---|---|---|
| 1 | Memuat file EJ | Pengguna memilih file log mentah ATM dari laptopnya. Aplikasi mengurainya menjadi daftar transaksi yang rapi, lalu menampilkan berapa transaksi yang berhasil dibaca. Kalau formatnya tidak dikenali, aplikasi berhenti dengan pesan yang jelas — bukan diam-diam menghasilkan nol baris. |
| 2 | Memuat file RC | Pengguna memilih file settlement bank. Aplikasi menghitung berapa baris valid yang terbaca dan menampilkannya, tanpa langsung memproses. |
| 3 | Pratinjau file | Sebelum diproses, isi awal file bisa diintip di layar supaya pengguna yakin tidak salah pilih file. |
| 4 | Jalankan rekonsiliasi | Setelah kedua file dimuat, satu tombol menjalankan pencocokan penuh. Prosesnya berjalan di latar dan bisa **dihentikan** di tengah jalan lewat tombol STOP. |
| 5 | Ringkasan hasil | Panel angka besar: total transaksi EJ, total baris RC, jumlah per kategori, dan total nominal tiap kategori. |
| 6 | Tabel hasil + filter | Seluruh baris hasil ditampilkan dalam tabel bernomor halaman, dengan tombol filter per kategori. |
| 7 | Ekspor hasil | Hasil bisa diunduh sebagai **Excel (.xlsx)** atau **teks (.txt)**, semua kategori sekaligus atau satu kategori saja. |
| 8 | Riwayat | Setiap proses yang selesai dicatat permanen di laptop. Hasil lama tetap bisa dibuka kembali walau aplikasi sudah ditutup dan dibuka lagi. |
| 9 | Reset | Satu tombol membuang sesi berjalan, siap mulai dari nol. |

#### 1.4 Hal yang sengaja tidak dibuat, dan alasannya

Bagian ini penting: beberapa hal di bawah **sengaja tidak dibuat**. Itu keputusan sadar dengan
alasan biaya dan lingkup, bukan kelalaian.

| Keputusan | Status | Alasan |
|---|---|---|
| Sepenuhnya offline | **Wajib** | Laptop cabang tidak diasumsikan punya internet, dan data transaksi nasabah tidak boleh keluar dari mesin. Seluruh pemrosesan terjadi di dalam laptop, tanpa satu pun kiriman data ke server luar. |
| Satu pengguna | **Wajib** | Aplikasi hanya bisa diakses dari laptop itu sendiri, tidak bisa dibuka dari komputer lain sekalipun berada di jaringan yang sama. |
| Tanpa layar login | Sengaja dihapus | Versi awal aplikasi ini adalah aplikasi web di server kantor yang diakses satu tim lewat jaringan — di sana login masuk akal. Setelah pindah jadi aplikasi di satu laptop pribadi, login hanya menyisakan satu layar tambahan tanpa menambah keamanan nyata. Pengamanan sesungguhnya diserahkan ke **kunci layar Windows** laptop tersebut. |
| Tanpa pembaruan otomatis | Di luar lingkup | Untuk satu laptop, mengirim installer baru secara manual setiap ada perbaikan sudah lebih murah dan lebih bisa diprediksi daripada membangun sistem pembaruan otomatis. |
| Windows 11 minimum | **Wajib** | Jendela aplikasi memakai komponen bawaan Windows 11, sehingga installer tidak perlu ikut membawa apa pun tambahan — instalasi tetap ringan dan bisa offline. |
| Kecepatan | Target | File EJ sekitar 25.000 baris dan file RC sekitar 1.000 baris harus selesai diproses dalam **hitungan detik**, bukan menit. |

#### 1.5 Catatan sejarah yang wajib diketahui

Aturan pengelompokan aplikasi ini bukan hasil tebakan di atas kertas — ia berubah beberapa kali
setelah diuji dengan data transaksi asli. Dua pelajaran ini penting agar tidak terulang:

> **Pelajaran 1 — kategori "Rollback" yang sempat dibuat lalu dibatalkan (Agustus 2026)**
Klien menyebut pola *"Nominal EJ dan Cash sama, EJ Status Rollback"*. Itu sempat dibaca sebagai
perintah membuat **kategori baru** bernama "Rollback", dan sempat tayang beberapa jam sebelum
ketahuan salah. Maksud sebenarnya adalah **deskripsi pola baris yang memang seharusnya masuk
Selisih Kurang**. Cerita lengkapnya ada di Bagian 3.

> **Pelajaran 2 — urutan pemeriksaan kategori itu menentukan hasil**
Pernah terjadi 12 dari 13 baris "Selisih Kurang" ternyata palsu: baris EJ berstatus rollback yang
**sama sekali tidak punya pasangan** di file RC ikut terseret masuk, padahal tanpa nominal
pembanding tidak ada "selisih" apa pun untuk dihitung. Urutan pemeriksaan sudah diperbaiki dan
dikunci sejak itu.

---

### 2. Alur Kerja Aplikasi

Begini cara staf cabang memakai Parse Bankers sehari-hari, dari membuka aplikasi sampai laporan
siap dilampirkan. Tidak ada langkah yang butuh pengetahuan komputer khusus — semuanya klik biasa.

- 1. Buka aplikasi

- 2. Pilih file EJ & RC

- 3. Klik Proses

- 4. Cek hasil per kategori

- 5. Unduh laporan

- Buka aplikasi Parse Bankers
Klik dua kali ikon Parse Bankers di Desktop. Jendela aplikasi langsung terbuka ke halaman
kerja — tidak ada layar login yang harus dilewati.
Masukkan file EJ dan file RC hari itu
Di kotak sebelah kiri, pilih **file EJ** — catatan mentah dari mesin ATM. Di kotak
sebelah kanan, pilih **file RC** — catatan settlement dari bank. Aplikasi langsung
menghitung dan menampilkan berapa baris data yang berhasil dibaca dari masing-masing file,
supaya bisa dipastikan filenya benar sebelum lanjut.
Klik tombol Proses
Tekan tombol hijau **PROCESS / RECON**. Dalam hitungan detik aplikasi selesai
membandingkan kedua file dan bilah status berubah menjadi "Proses selesai!". Ada tombol
**STOP** kalau prosesnya perlu dibatalkan di tengah jalan.
Tampilan asli aplikasi — kotak pemilih file EJ & RC, dan tombol Proses/Recon.
Di balik layar, aplikasi mencocokkan otomatis
Aplikasi membandingkan catatan mesin ATM dengan catatan bank, baris demi baris, mencari
transaksi dengan nomor yang sama di kedua sisi. Setiap pasangan transaksi lalu diberi label
sesuai kondisinya: cocok, ada selisih, atau tidak ditemukan pasangannya. Semua ini terjadi
di dalam laptop, tidak ada data yang dikirim ke mana pun.
Lihat hasil, dikelompokkan per kategori
Tabel hasil muncul lengkap dengan tombol saring di atasnya untuk tiap kategori:
Selisih Kurang (transaksi yang patut dicurigai, biasanya karena
mesin sempat menarik kembali uangnya),
Tidak Ditemukan (tercatat di satu sisi saja),
Match / Klop (aman, kedua catatan sesuai), dan
Data Rusak (nominal tidak terbaca). Klik salah satu tombol untuk
hanya menampilkan baris kategori itu.
Tampilan asli aplikasi — hasil rekonsiliasi dengan tombol saring per kategori.
Unduh laporan
Tekan tombol **Export Excel** atau **Export TXT** untuk menyimpan hasilnya sebagai
file — bisa semua kategori sekaligus, atau hanya kategori yang sedang disaring. File Excel
sudah rapi dengan warna baris sesuai kategorinya, tinggal dilampirkan ke laporan harian.

> **Kalau perlu mulai ulang atau membuka hasil lama**
Tombol **Reset** membersihkan sesi yang sedang berjalan untuk mulai dari nol. Sementara itu,
panel **Riwayat** menyimpan seluruh proses yang pernah dijalankan — hasil kemarin atau
minggu lalu tetap bisa dibuka lagi kapan saja, bahkan setelah aplikasinya ditutup dan dibuka ulang.

---

### 3. Perjalanan Pembuatan

Parse Bankers tidak langsung jadi seperti sekarang. Aplikasi ini melalui beberapa versi, beberapa
kali revisi, dan satu kesalahpahaman kecil yang untungnya cepat dibetulkan. Berikut ceritanya,
dari titik paling awal sampai versi yang sekarang dipakai.

- 1
Awal mula
Aplikasi web yang jalan di server kantor
Versi pertama Parse Bankers adalah aplikasi web internal yang berjalan di
server kantor dan diakses lewat jaringan, lengkap dengan layar login untuk tim yang memakainya bersama.

- 2
Agustus 2026
Kategori "Rollback" yang sempat salah dipahami
Klien menjelaskan sebuah pola transaksi memakai kata "Rollback". Kalimat itu
sempat dibaca sebagai permintaan membuat kategori baru bernama "Rollback", dan sempat tayang
beberapa jam di aplikasi. Setelah dicek ulang bersama klien, ternyata maksudnya bukan kategori
baru, melainkan penjelasan tentang pola transaksi yang memang sudah semestinya masuk kategori
"Selisih Kurang" yang sudah ada. Begitu ketahuan, perubahan itu langsung dikembalikan seperti
semula pada hari yang sama.

- 3
30 Agustus 2026
Kode aplikasi mulai dicatat rapi dan tersimpan aman
Seluruh bagian aplikasi (bagian pengolah data maupun tampilannya) dikumpulkan
dan mulai dicatat riwayat perubahannya secara rapi, sebagai titik awal yang jelas sebelum
aplikasi ini diubah menjadi versi yang bisa dipasang di laptop.

- 4
30 Agustus 2026
Dibikin bisa berdiri sendiri di laptop, tanpa perlu login
Aplikasi diubah agar bisa berjalan sendirian di satu laptop tanpa bergantung
pada server kantor. Karena sekarang hanya dipakai satu orang di laptopnya sendiri, layar login
yang tadinya diperlukan untuk penggunaan bersama juga dihapus — sudah tidak relevan lagi.

- 5
30 Agustus 2026
Dibungkus jadi jendela aplikasi sungguhan
Tampilannya diubah supaya terbuka sebagai **jendela aplikasi Windows biasa**
dengan judul dan ikon sendiri di taskbar, bukan lagi lewat tab browser yang bisa membingungkan
pengguna yang tidak terbiasa dengan istilah teknis.

- 6
30 Agustus 2026
Semuanya dibundel jadi satu file installer Windows
Aplikasi dibungkus menjadi satu file installer yang tinggal diklik dua kali
untuk dipasang — lengkap dengan pembuatan shortcut di Desktop dan Start Menu, mengikuti alur
pasang aplikasi Windows yang sudah dikenal umum.

- 7
1 September 2026
Tampilan dirapikan jadi satu desain modern, dipasangi logo resmi BNI
Dua gaya tampilan yang tadinya terpisah digabung jadi satu tampilan modern
yang lebih rapi dan konsisten. Logo resmi BNI juga dipasang di header aplikasi, ikon program,
dan tampilan installer-nya.

- 8
1–3 September 2026
Serangkaian perbaikan akurasi pencocokan data
Setelah diuji dengan data transaksi asli, ditemukan beberapa celah kecil pada
logika pencocokan — misalnya transaksi kembar yang seharusnya dihitung satu kali, dan urutan
pemeriksaan kategori yang sempat membuat beberapa transaksi salah masuk kelompok (lihat Bagian
1.5). Semuanya diperbaiki bertahap sambil terus diuji ulang dengan data nyata.

- 9
3 September 2026 — sekarang
Versi stabil yang sekarang dipakai
Perbaikan terakhir menuntaskan beberapa bug kecil pada tombol ekspor dan
penyimpanan riwayat. Inilah versi yang sekarang digunakan sehari-hari, dengan tampilan seperti
pada gambar di bawah.

---

### 4. Cara Pakai & Batasan

#### 4.1 Cara memasang (cukup dilakukan sekali)

1. Salin file installer **Parse Bankers** ke laptop (lewat flashdisk atau jaringan internal).
2. Klik dua kali file tersebut.
3. Windows biasanya akan menampilkan layar biru bertuliskan "Windows protected your PC". Ini
**normal** dan bukan tanda ada masalah — muncul karena aplikasi internal ini belum
didaftarkan resmi ke Microsoft (perusahaan besar sekalipun sering mengalami ini untuk aplikasi
internal). Caranya melewati:

klik tulisan kecil "More info" → lalu klik tombol "Run anyway".
4. Ikuti wizard pemasangan: **Next → Next → Install → Finish.**
5. Selesai. Ikon **Parse Bankers** otomatis muncul di Desktop dan Start Menu.

> **Layar biru ini bisa muncul lagi setiap kali ada installer versi baru**
Setiap versi baru punya isi berkas yang berbeda, jadi Windows menganggapnya berkas asing lagi.
Cukup ulangi langkah "More info → Run anyway" setiap kali menerima installer baru.

#### 4.2 Cara pakai sehari-hari

| Langkah | Yang dilakukan | Yang terjadi / yang terlihat |
|---|---|---|
| 1 | Buka **Parse Bankers** dari ikon Desktop | Jendela aplikasi terbuka langsung ke halaman kerja. Tidak ada layar login. |
| 2 | Pada kotak **Electronic Journal**, pilih file EJ hari itu | Aplikasi membaca log dan menampilkan **jumlah transaksi** yang berhasil dibaca. |
| 3 | Pada kotak **Cash / Reconciliation**, pilih file RC hari itu | Aplikasi menampilkan **jumlah baris valid** yang terbaca. |
| 4 | Tekan tombol **Proses** | Pencocokan berjalan (hitungan detik). Tombol **STOP** tersedia bila perlu dibatalkan. |
| 5 | Baca panel **ringkasan** | Total EJ, total Cash, jumlah per kategori, dan total nominal tiap kategori. |
| 6 | Klik tombol saring kategori | Tabel hanya menampilkan baris kategori tersebut. Biasanya yang pertama diperiksa adalah Selisih Kurang, lalu Tidak Ditemukan. |
| 7 | Tekan **Export** | File hasil (Excel atau teks) tersimpan, siap dilampirkan ke laporan. Bisa mengekspor semua kategori sekaligus atau satu saja. |
| 8 | Bila perlu mulai lagi | Tombol **Reset** membersihkan sesi. Hasil lama tetap bisa dibuka lewat panel **Riwayat**, bahkan setelah aplikasi ditutup dan dibuka kembali. |

#### 4.3 Batasan yang diketahui

| Batasan | Penjelasan |
|---|---|
| Hanya untuk satu laptop | Tidak ada mekanisme berbagi hasil antar-komputer. Riwayat dan hasil tersimpan hanya di laptop tempat aplikasi dipasang. |
| Tanpa pembaruan otomatis | Setiap perbaikan berarti installer baru dikirim dan dipasang ulang secara manual, termasuk mengulang langkah melewati layar biru "Windows protected your PC". |
| Cara baca log terikat pada satu jenis mesin ATM | Cara aplikasi membaca file EJ disusun berdasarkan format log dari mesin ATM yang dipakai sekarang. Kalau nanti ada mesin dengan merek atau format catatan berbeda, cara bacanya perlu disesuaikan dan diuji ulang dengan data asli dari mesin itu. |
| Folder hasil menumpuk seiring waktu | Hasil setiap proses disimpan permanen dan belum ada pembersihan otomatis. Bisa dibersihkan sewaktu-waktu lewat tombol pembersih riwayat. |
| Kolom "No Transaksi" pada baris penarikan | Untuk beberapa jenis transaksi penarikan tunai, kolom ini menampilkan kode jenis transaksi, bukan nomor transaksi. Ini hanya soal tampilan satu kolom informasi dan tidak memengaruhi hasil pencocokan. |

> Kalau hanya satu paragraf yang sempat dibaca
Parse Bankers adalah aplikasi Windows yang membaca **catatan mesin ATM** dan
**catatan settlement kas bank**, mencocokkan keduanya secara otomatis, lalu mengelompokkan
setiap transaksi ke dalam empat kategori: Match
Selisih Kurang Tidak Ditemukan
Data Rusak — dengan **Selisih Kurang** sebagai temuan
terpenting yang paling perlu ditindaklanjuti. Semuanya berjalan offline di satu laptop, dan
hasilnya bisa diunduh sebagai Excel untuk dilampirkan ke laporan harian.

Dokumen ini adalah edisi bahasa mudah dari panduan lengkap aplikasi Parse Bankers versi 1.1.0.
Versi teknis lengkap (spesifikasi format file, aturan pencocokan rinci, dan tumpukan teknologi
yang dipakai) tersedia terpisah bagi pembaca yang membutuhkan detail tersebut.

