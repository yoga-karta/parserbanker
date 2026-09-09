# Parse Bankers — Panduan Vibecoding Bangun dari Nol

---

## Parse Bankers

### Daftar Isi

> **Catatan penting sebelum mulai baca**
Dokumen ini murni soal **mengobrol dengan Claude** untuk minta dibikinkan, dites, dan
dipaketkan jadi aplikasi siap pakai. Semua pekerjaan teknis di baliknya — menulis kode,
mencoba jalankan, membetulkan kalau error, sampai membungkusnya jadi installer —
dikerjakan Claude sendiri. Tidak ada langkah yang mengharuskan mengenal istilah pemrograman
atau perkakas developer apa pun.

---

### 1. Apa Itu Vibecoding?

#### 1.1 Penjelasan singkat

**Vibecoding** artinya bikin aplikasi dengan cara **ngobrol biasa** ke AI (dalam hal ini
Claude), pakai bahasa sehari-hari, tanpa perlu ngerti coding sama sekali. Kamu jelasin apa yang
kamu mau — masalah apa yang mau diselesaikan, siapa yang bakal pakai, gimana cara kerjanya
yang kamu bayangkan — dan Claude yang mikirin serta ngerjain seluruh bagian teknisnya:
nulis programnya, nyoba jalanin sendiri, betulin kalau ketemu error, sampai pas saatnya
membungkusnya jadi aplikasi yang siap dipakai.

Dokumen ini adalah cerita nyata gimana **Parse Bankers** sendiri dibangun dengan cara ini,
tahap demi tahap, dari yang tadinya cuma ide sederhana — "aku mau dua jenis file transaksi
ini otomatis dicocokkan" — sampai jadi aplikasi Windows lengkap dengan branding resmi yang
sekarang dipakai staf cabang setiap hari. Tiap tahap di Bagian 2 disertai **contoh kalimat
prompt asli** yang bisa langsung dicontek dan disesuaikan untuk kebutuhan lain.

#### 1.2 Yang perlu disiapkan

| Punya Claude Code | Cukup terpasang di komputer. Tidak perlu akun developer khusus atau langganan tambahan di luar itu. |
|---|---|
| Folder kosong | Buka folder kosong di komputer, itu tempat Claude bakal nyimpen dan ngerjain seluruh aplikasinya. Tidak perlu disiapkan apa-apa di dalamnya dulu. |
| Contoh data asli (kalau ada) | Kalau aplikasinya bakal ngolah jenis file tertentu (seperti file EJ dan RC di sini), siapin contoh file aslinya untuk dilampirkan — contoh nyata jauh lebih membantu Claude ngerti persis bentuknya, dibanding cuma dideskripsikan lewat kata-kata. |
| Laptop/PC Windows (belakangan) | Untuk ngobrol dan nyoba-nyoba di awal, komputer apa saja bisa dipakai. Tapi begitu masuk ke tahap membungkus aplikasi jadi installer Windows siap-pasang (Tahap 3 di Bagian 2), itu perlu dikerjakan di laptop atau PC Windows — sekadar fakta perangkat keras: hasil akhirnya memang aplikasi Windows, jadi proses terakhir merakitnya juga perlu Windows. |

> **Nggak perlu ngerti coding sama sekali**
Sepanjang delapan tahap di Bagian 2, tidak ada satu pun langkah yang mengharuskan mengetik
perintah teknis, mengerti istilah pemrograman, atau membuka perkakas developer. Semua yang
perlu dilakukan hanyalah **menjelaskan apa yang diinginkan**, dengan kalimat sendiri, seperti
lagi menjelaskan ke rekan kerja — Claude yang menerjemahkannya jadi aplikasi sungguhan.

---

### 2. Tahapan Membangun Parse Bankers dari Nol

Delapan tahap berikut adalah urutan nyata bagaimana Parse Bankers dibangun, dari prompt paling
awal sampai versi final yang sekarang dipakai sehari-hari. Tiap tahap berisi penjelasan singkat
kenapa tahap itu diperlukan, satu contoh kalimat prompt yang bisa langsung dicontek, dan
catatan singkat soal apa yang bakal dilakukan Claude setelah menerima prompt itu.

- 1. Kebutuhan dasar & minta dibikinin aplikasi pencocokan
2Jadi aplikasi mandiri di laptop, tanpa server/login
3Dibungkus jadi installer Windows
4Tampilan dirapikan & didesain ulang
5Branding logo resmi dipasang
6Contoh ngasih revisi requirement yang jelas
7Contoh lapor bug secara efektif
8Cek akhir, testing, & installer final

- Prompt awal — jelasin kebutuhan dasar, minta dibikinkan aplikasi pencocokan file
Semuanya dimulai dari satu masalah sehari-hari: dua jenis catatan transaksi yang harus
dicocokkan manual, dan itu lambat serta rawan salah. Tahap ini adalah prompt pembuka yang
menjelaskan masalahnya dari sudut pandang orang yang mengalaminya langsung, bukan dari sudut
pandang teknis — dilengkapi contoh file asli supaya Claude paham persis bentuk datanya.
Contoh prompt
"Halo Claude, aku kerja di bank, tugasku tiap hari cocokin dua jenis catatan transaksi ATM.
Yang pertama namanya file EJ — itu rekaman otomatis dari mesin ATM, isinya panjang
banget, satu file bisa puluhan ribu baris, dan satu transaksi itu nyebar di banyak baris
karena nyatet detail kayak kapan kartu masuk, PIN dimasukin, duit keluar, kadang malah
duitnya ketarik balik ke mesin (namanya 'rollback'). Yang kedua namanya file RC — ini
catatan dari sistem bank, lebih rapi, satu baris satu transaksi, isinya nomor rekening,
nomor urut transaksi, nominal, sama tanda transaksi itu debit atau kredit. Kerjaanku
sekarang buka dua file itu manual tiap hari, nyari nomor transaksi yang sama di kedua file
satu-satu, terus bandingin nominalnya. Capek, kadang keliru, apalagi transaksi yang duitnya
ketarik balik itu suka keselip nggak kelihatan. Aku mau kamu bikinin aplikasi: aku tinggal
pilih file EJ hari itu, pilih file RC hari itu, pencet satu tombol, terus aplikasinya
otomatis bandingin dan ngasih tabel hasil — mana yang cocok, mana yang beda
nominalnya, mana yang cuma ada di salah satu file, dan yang paling penting: mana transaksi
yang kelihatannya duitnya udah ketarik balik ke mesin tapi bank tetap nyatat nominal itu
keluar, soalnya itu tandanya duit kemungkinan hilang. Aku juga mau bisa donlot hasilnya ke
Excel biar gampang dilampirin ke laporan. Contoh file aslinya aku lampirin ya, biar kamu
ngerti persis bentuknya."
**Apa yang bakal Claude lakuin**
Claude bakal nanya sedikit hal yang belum jelas kalau perlu, lalu langsung menulis
programnya dari nol — bagian yang membaca kedua file, bagian yang membandingkan isinya,
sampai tampilan sederhana untuk memilih file, memencet tombol, dan melihat hasilnya. Claude
mencoba menjalankannya sendiri pakai contoh file yang dilampirkan, dan kalau ada yang belum
jalan, dibetulkan sendiri sampai benar-benar berhasil.

- Minta dijadikan aplikasi desktop mandiri, tanpa server dan tanpa login
Versi paling awal tadi masih berjalan seperti "website" yang dibuka lewat alamat di server
kantor, bisa diakses banyak orang, makanya ada layar login. Begitu jelas bahwa aplikasi ini
cuma akan dipakai satu orang di satu laptop pribadinya, layar login itu jadi langkah tambahan
yang tidak perlu — tahap ini mengubah aplikasi supaya berdiri sendiri penuh di satu
laptop, tanpa bergantung ke server mana pun.
Contoh prompt
"Claude, ternyata aplikasi ini nanti cuma bakal dipakai satu orang di satu laptop kantor,
dan laptopnya juga sering nggak connect internet. Jadi tolong aplikasinya dibikin bisa
jalan sendirian penuh di laptop itu aja, nggak butuh nyambung ke server mana-mana. Terus,
karena yang pakai cuma aku sendirian di laptopku, halaman login yang tadinya ada nggak
usah dipakai lagi, tolong dihapus aja biar buka aplikasinya langsung ke halaman kerja. Oh
iya, pastiin juga aplikasinya beneran nggak bisa diakses dari laptop lain, sekalipun
kebetulan nyambung ke wifi yang sama."
**Apa yang bakal Claude lakuin**
Claude bakal mengubah bagian dalam programnya supaya semua data disimpan di laptop itu
sendiri, mencopot halaman login sepenuhnya (termasuk tombolnya, bukan cuma disembunyikan),
dan memastikan aplikasinya cuma bisa dibuka dari laptop itu sendiri — bukan dari
komputer lain di jaringan yang sama.

- Minta dibungkus jadi installer Windows yang gampang dipasang
Aplikasinya sekarang sudah mandiri, tapi masih kebuka mirip tab browser dengan alamat teknis
di atasnya — berpotensi bikin bingung atau curiga orang yang tidak terbiasa. Tahap ini
minta dua hal sekaligus: jadikan jendela aplikasi biasa (seperti aplikasi Windows pada
umumnya), dan bungkus semuanya jadi satu file installer yang tinggal diklik dua kali untuk
dipasang.
Contoh prompt
"Sekarang soal tampilannya waktu dibuka — sekarang masih kebuka kayak tab browser,
ada alamat teknis di atas, aku khawatir orang yang pakai nanti bingung atau malah curiga.
Tolong bikin aplikasinya kebuka sebagai jendela aplikasi biasa aja, ada judulnya sendiri
'Parse Bankers', ada ikonnya sendiri, persis kayak aplikasi Windows kebanyakan, jangan
kelihatan alamat browser sama sekali. Terus buat cara masangnya aku pengen simpel: aku
kasih satu file ke staf cabang, dia tinggal klik dua kali, ikutin next-next biasa, kelar.
Setelah kepasang harus otomatis muncul ikon di Desktop sama di Start Menu, dan kalau nanti
mau dicopot juga bisa lewat menu uninstall Windows biasa. Semua ini harus tetap jalan tanpa
internet, baik waktu dipasang maupun waktu dipakai."
**Apa yang bakal Claude lakuin**
Claude bakal mengubah cara aplikasinya tampil supaya jadi jendela aplikasi asli dengan judul
dan ikon sendiri, lalu menyiapkan satu file installer lengkap dengan wizard pasang
"Next – Next – Finish", shortcut Desktop dan Start Menu, serta entri uninstall
resmi.
**Satu fakta perangkat keras yang perlu diketahui**
Langkah merakit jendela aplikasi dan installer di tahap ini **harus dikerjakan di laptop
atau PC Windows**, karena hasil akhirnya memang aplikasi Windows. Kalau proses ngobrolnya
dimulai dari komputer lain, cukup pindahkan ke laptop Windows begitu sampai di tahap ini
— Claude akan memberi tahu kalau ada bagian yang memang perlu diselesaikan di sana.

- Minta tampilan dirapikan dan didesain ulang
Di titik ini aplikasinya sudah berfungsi, tapi tampilannya masih berupa dua gaya terpisah
yang membingungkan. Tahap ini menyatukan semuanya jadi satu tampilan modern dan konsisten:
kartu ringkasan angka besar di atas, tabel hasil dengan tombol saring per kategori, dan
panel detail yang muncul saat satu baris transaksi diklik.
Contoh prompt
"Tampilannya sekarang masih berantakan menurutku, ada dua mode tampilan yang bikin bingung,
aku maunya disederhanain jadi satu tampilan aja yang rapi dan modern. Di atas aku pengen ada
kartu-kartu ringkasan angka besar (total transaksi, jumlah tiap kategori temuan), terus di
bawahnya tabel hasil lengkap dengan tombol buat nyaring per kategori biar bisa fokus lihat
satu jenis temuan aja. Kalau aku klik satu baris di tabel, munculin detail lengkap transaksi
itu di panel samping, biar tabelnya sendiri nggak penuh sesak informasi. Buat warnanya,
coba pakai nuansa oranye sama biru tua, soalnya nanti bakal ada logo perusahaan yang aku
pasang juga."
**Apa yang bakal Claude lakuin**
Claude bakal mendesain ulang tampilannya, menggabungkan dua gaya lama jadi satu, menyusun
ulang kartu ringkasan, tabel, tombol saring, dan panel detail, lalu mencoba jalankan lagi
untuk memastikan semua fungsi lama masih berjalan normal setelah tampilannya berubah.

- Minta branding logo resmi dipasang
Aplikasinya sudah rapi dan sudah dipakai sungguhan, jadi saatnya memasang identitas resmi
perusahaan supaya terasa seperti aplikasi resmi, bukan lagi versi coba-coba.
Contoh prompt
"Ini aku kasih file logo resmi perusahaan (BNI) dalam bentuk gambar. Tolong pasang logo ini
di pojok atas aplikasi, terus jadiin juga sebagai ikon aplikasinya — yang muncul di
taskbar dan pas di-Alt-Tab — dan ikon installer-nya juga. Oh iya, ini aplikasinya
sekarang udah dipakai beneran sehari-hari ya, bukan lagi versi coba-coba, jadi tolong cek
juga kalau masih ada tulisan 'Pilot' atau semacamnya di aplikasi, itu dihapus aja."
**Apa yang bakal Claude lakuin**
Claude bakal mengolah file logo itu supaya rapi dipasang (misalnya latar belakangnya
dijadikan transparan), memasangnya di beberapa tempat sekaligus (header aplikasi, ikon
program, ikon installer), dan membersihkan sisa-sisa tulisan lama yang sudah tidak relevan.

- Contoh cara ngasih revisi requirement yang jelas ke Claude
Ada satu momen nyata di perjalanan Parse Bankers: penjelasan soal pola transaksi tertentu
sempat salah dibaca sebagai permintaan bikin kategori baru bernama "Rollback", padahal
maksudnya cuma deskripsi ciri-ciri transaksi yang memang sudah semestinya masuk kategori
Selisih Kurang yang sudah ada. Begitu ketahuan salah,
langsung dikembalikan seperti semula. Ini contoh nyata bagaimana ngasih revisi requirement
yang jelas ke Claude, sekaligus contoh bagaimana meminta Claude untuk mengonfirmasi ulang
kalau ada instruksi yang bisa diartikan dua cara.
Contoh prompt
"Eh Claude, soal kategori temuan tadi — waktu aku bilang 'transaksi yang nominalnya
sama tapi statusnya rollback', maksudku itu BUKAN minta kamu bikin kategori baru bernama
'Rollback' ya. Itu cuma aku jelasin ciri-ciri satu jenis transaksi, yang sebenernya harusnya
tetap masuk ke kategori 'Selisih Kurang' yang udah ada, sama kayak sebelumnya. Tolong
balikin lagi kayak semula — hapus kategori 'Rollback' yang baru itu, di mana pun dia
nongol (tombol saring, ringkasan angka, dan lain-lain), dan pastiin transaksi kayak gitu
balik masuk ke Selisih Kurang seperti awal. Ke depannya, kalau aku jelasin ciri-ciri
transaksi kayak gini lagi, coba tanya dulu ke aku: ini maksudnya kategori baru, atau cuma
penjelasan buat kategori yang udah ada?"
**Apa yang bakal Claude lakuin**
Claude bakal mengembalikan perubahan yang salah itu ke kondisi semula, lalu mencoba lagi
pakai data transaksi asli untuk memastikan hasilnya benar-benar sama seperti sebelum
kesalahpahaman itu terjadi. Ke depannya, kalau ada instruksi yang bisa diartikan dua cara
(seperti yang diminta di contoh prompt di atas), Claude akan menanyakan dulu maksudnya
sebelum langsung mengerjakan — supaya kejadian serupa tidak terulang.

- Contoh cara melaporkan bug secara efektif ke Claude
Setelah aplikasinya jadi installer Windows sungguhan, muncul satu bug nyata: tombol export
Excel/TXT sempat tidak merespons sama sekali begitu dijalankan sebagai aplikasi terpasang di
laptop, walaupun berjalan normal sebelumnya. Ini contoh laporan bug yang efektif —
jelas soal apa yang dilakukan, apa yang seharusnya terjadi, apa yang sebenarnya terjadi, dan
di kondisi apa persis bug itu muncul.
Contoh prompt
"Claude, ada bug nih di aplikasi versi Windows kemarin: pas aku klik tombol 'Export Excel'
atau 'Export TXT' di aplikasi yang udah kepasang di laptop, nggak kejadian apa-apa. Nggak
ada file kesimpan, nggak ada pesan error, tombolnya kayak nggak ngapa-ngapain, padahal
proses rekonsiliasinya sendiri udah selesai dan hasilnya kelihatan normal di tabel. Ini
kejadiannya cuma di aplikasi yang udah jadi installer di laptop Windows, bukan pas masih
dicoba-coba di komputermu. Tolong dicek kenapa, terus dibetulin sampai tombol export-nya
beneran ngedownload filenya."
**Apa yang bakal Claude lakuin**
Claude bakal menelusuri kenapa tombol itu tidak bekerja — biasanya karena ada satu
detail kecil yang perilakunya berbeda di aplikasi Windows dibanding saat masih dicoba di
komputer biasa — membetulkannya, lalu mencoba lagi sampai tombol export-nya benar-benar
berhasil mengunduh file.

- Prompt penutup — pastikan semuanya jalan, ditest, dan installer final siap dipakai
Tahap terakhir sebelum aplikasinya benar-benar dipakai sehari-hari: minta Claude mengecek
ulang semuanya secara menyeluruh dari awal sampai akhir, memakai data transaksi asli, sebelum
menyiapkan file installer versi final yang siap dikirim.
Contoh prompt
"Oke Claude, sebelum ini beneran dipakai sehari-hari, tolong dicek ulang menyeluruh ya
— coba jalanin aplikasinya dari awal sampai akhir pakai data transaksi asli yang udah
aku kasih, pastiin semua tombol jalan (proses, saring kategori, export Excel, export TXT,
riwayat, reset), dan pastiin angka-angka di ringkasan hasilnya masuk akal dan sama kayak
yang aku cek manual. Kalau ketemu yang aneh, tolong dibenerin dulu sebelum lanjut. Kalau
semua udah oke, tolong siapin file installer final yang siap aku kirim ke laptop staf
cabang."
**Apa yang bakal Claude lakuin**
Claude bakal menjalankan aplikasinya ujung ke ujung, mencoba berbagai skenario (termasuk
kasus-kasus data yang agak aneh atau tidak lengkap), membetulkan kalau ketemu masalah, lalu
menyiapkan file installer versi final yang sudah siap dikirim dan dipasang.

---

### 3. Tips Biar Obrolan Sama Claude Makin Oke

Beberapa kebiasaan kecil ini yang bikin proses vibecoding di atas berjalan lancar dari awal
sampai akhir, dan cocok dipakai untuk proyek apa pun, bukan cuma Parse Bankers.

> 1. Kasih contoh nyata, jangan cuma dijelaskan
Lampirkan contoh file, gambar, atau data asli setiap kali relevan
— seperti file EJ dan RC asli di Tahap 1. Contoh nyata jauh lebih meyakinkan buat Claude
dibanding deskripsi panjang lewat kata-kata saja.

> 2. Kalau hasilnya belum pas, bilang apa adanya
Tidak perlu sungkan bilang "ini belum sesuai, maunya begini...".
Semakin jelas bedanya antara yang diharapkan dan yang terjadi, semakin cepat dan tepat
Claude membetulkannya — seperti contoh di Tahap 7.

> 3. Minta Claude jelasin balik kalau ragu
Kalau instruksinya bisa diartikan dua cara, minta Claude
menjelaskan ulang pemahamannya dengan kata-katanya sendiri sebelum mulai mengerjakan —
supaya salah paham ketahuan dari awal, bukan setelah terlanjur jadi, seperti pelajaran dari
Tahap 6.

> 4. Selalu minta dicoba pakai data asli dulu
Sebelum benar-benar dipakai sehari-hari, minta Claude menjalankan
dan menguji aplikasinya memakai data asli, bukan cuma data contoh — seperti prompt
penutup di Tahap 8. Perilaku aplikasi terhadap data nyata sering mengungkap hal yang tidak
kelihatan dari data contoh.

> Kalau hanya satu paragraf yang sempat dibaca
Vibecoding berarti membangun aplikasi lewat obrolan biasa dengan Claude, bukan lewat menulis
kode sendiri. Jelaskan masalahnya seperti menjelaskan ke rekan kerja, lampirkan contoh nyata
kalau ada, kasih tahu kalau hasilnya belum pas, dan selalu minta dicoba dengan data asli
sebelum benar-benar dipakai — delapan tahap di Bagian 2 adalah bukti bahwa alur sesederhana
itu cukup untuk membangun aplikasi sungguhan dari nol sampai siap pakai.

Dokumen ini adalah panduan ketiga dari seri dokumentasi Parse Bankers, disusun dari riwayat
pembuatan aplikasi versi 1.1.0 yang sesungguhnya. Untuk penjelasan kebutuhan dan cara pakai
aplikasinya sehari-hari, lihat Panduan Versi Mudah; untuk spesifikasi teknis lengkap, lihat
Panduan Membangun Parse Bankers.

