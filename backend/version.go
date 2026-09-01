package main

// AppVersion - bump tiap rilis, dan samakan juga dengan AppVersion di
// installer/setup.iss. Dua tempat manual karena Inno Setup gak bisa baca
// konstanta Go langsung.
// ponytail: sinkronisasi manual 2 file, ganti ke satu file VERSION + step CI
// yang inject ke keduanya kalau rilisnya udah sering banget.
const AppVersion = "1.1.0"
