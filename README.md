# Sky - Yet Another Music Player Project

> Music player CLI/TUI yang ringan, keyboard-first, cross-platform.

Sky adalah music player berbasis terminal yang ditulis dengan Go. Bisa scan
folder musik lokal, simpan metadata ke SQLite, dan memutar audio dengan
TUI yang responsif. Tidak butuh GUI, tidak butuh server — cukup terminal.

## Fitur

- **Multi-format audio** — MP3, FLAC, OGG (Vorbis), WAV
- **Cross-platform** — Windows, Linux, macOS (pure Go, tanpa CGO)
- **Scan folder rekursif** — dengan skip folder tersembunyi otomatis
- **Metadata parsing** — ID3v1/v2, Vorbis comment, MP4 atom
- **Mini file browser** — tambah folder tanpa perlu ketik path manual
- **Kelola folder** — tambah, hapus, dengan preview isi folder
- **Queue playback** — play dari library, next/prev, auto-advance
- **Repeat mode** — off / one / all, persisten antar sesi
- **Volume control** — dengan mapping logaritmik, persisten
- **Seek** — maju/mundur 5 detik
- **Progress bar** — durasi real-time, format `0:38 ━━━━━ 4:48`
- **Resume state** — ingat track dan posisi terakhir, resume dalam keadaan pause
- **Filter** — cari lagu by title/artist/album
- **Help overlay** — panduan hotkey lengkap dengan `?`
- **Toast notifications** — feedback singkat untuk aksi

## Instalasi

### Prasyarat

- Go 1.21 atau lebih baru
- Terminal yang mendukung ANSI colors (Windows Terminal, iTerm2, GNOME Terminal, Alacritty, WezTerm, dsb)

### Build dari source

```bash
git clone https://github.com/ade-nugraha306/sky-project.git
cd sky-project
go build -o sky .
````

Jalankan:

````bash
./sky        # Unix
.\sky.exe    # Windows
````

Jalankan tanpa build:

````bash
go run .
````

### Hotkeys

#### Library (mode utama)

Tombol|Aksi
------|----
`↑` / `↓`   |Navigasi list
`Enter`	|Play 
`Spasi`   | Pause / Resume
`n` / `>`   | Track berikutnya
`p` / `<`   | Track sebelumnya
`,` / `[`   | Mundur 5 detik
`.` / `]`   | Maju 5 detik
`m`       | Ganti repeat mode
`+` / `=`   | Volume naik 5%
`-` / `_`   | Volume turun 5%
`/`       | Filter / Search mode
`x`       | Bersihkan Filter / Search
`a`       | Tambah folder musik (membuka browser file)
`r`       | rescan library
`q` / `Esc` | Keluar aplikasi
`Ctrl+C`  | Keluar paksa dari aplikasi

---

#### Browser Folder

---

Tombol | Aksi
-------|------
`↑` / `↓`   |Navigasi list
`Enter`     | Masuk Folder
`Backspace` / `←` | Kembali / Naik satu level
`s`          | Pilih folder yang sedang dibuka sebagai music folder
`/`           | Cari folder
`?`           | Buka Panduan
`Esc`         | Batal

---

#### Kelola Folder

---


Tombol | Aksi
-------|------
`↑` / `↓`   |Navigasi list
`d` / `Backspace` | Hapus folder dari daftar
`Esc` / `q`  | Kembali ke library

### Batasan

- M4A / AAC / Opus belum didukung
- Tidak ada playlist untuk sekarang.

### Future Update

- [ ]Playlist buat / edit / hapus
- [ ] Mouse Click Support
- [ ] Shuffle mode
- [ ] Sort library
- [ ] Global Search
- [ ] Format M4A