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

### Navgiasi Global

Tombol|Aksi
------|----
`L`   | Ke Library
`P`  	| Ke Playlists
`B` 	| Ke File Browser
`F`  	| Ke Kelola Folder
`x`	  | Bersihkan filter di list aktif
`?`	  | Buka panduan hotkey lengkap

#### Library 

Tombol|Aksi
------|----
`↑` / `↓`   |Navigasi list
`Enter`	|Play 
`Spasi`	| Pause / resume
`n` / `>`	| Track berikutnya
`p` / `<`	| Track sebelumnya (restart kalau > 3 detik)
`,` / `[`	| Mundur 5 detik
`.` / `]`	| Maju 5 detik
`m`	     | Ganti repeat: off → one → all
`s`	     | Toggle shuffle
`+` / `=`	 | Volume naik 5%
`-` / `_`	 | Volume turun 5%
`/`      | Filter / search
`x`	     | Bersihkan filter
`t`	     | Tambah track ke playlist
`a`	     | Tambah folder musik (buka browser)
`d`	     | Kelola folder musik
`r`	     | Rescan library
`q` / `Esc`	| Keluar (resume disimpan)
`Ctrl+C` | Keluar paksa

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

---

#### Playlists

---

Tombol | Aksi
-------|------
`↑` / `↓`   | Navigasi list
`Enter`	| Buka detail playlist
`n`	    | Buat playlist baru
`r`   	| Rename playlist
`d`	    | Hapus playlist
`/`	    | Filter playlist
`Esc` / `q`	| Kembali ke library

---

#### Detail Playlist

---
Tombol | Aksi
-------|------
`↑` / `↓`   | Navigasi list
`Enter`	| Play track 
`n` / `>`	| Track berikutnya
`p` / `<`	| Track sebelumnya (restart kalau > 3 detik)
`,` / `[`	| Mundur 5 detik
`.` / `]`	| Maju 5 detik
`Spasi`	| Pause / resume
`m`	    | Ganti repeat mode
`d`	    | Hapus track dari playlist
`/`	    | Filter track
`Esc` / `Backspace` | Kembali ke daftar Playlists

---

#### Playlist Picker

---

Tombol | Aksi
-------|------
`↑` / `↓`   | Navigasi list
`Enter`	| Play playlist
`/`	    | Filter playlist
`Esc`	  | Batal

---

### Tech Stack

Komponen | Library
---------|--------
TUI framework  | [Bubble Tea](https://github.com/charmbracelet/bubbletea)
TUI components | [Bubbles](https://github.com/charmbracelet/bubbles)
TUI styling	   | [Lip Gloss](https://github.com/charmbracelet/lipgloss)
Audio playback | [gopxl/beep v2](https://github.com/gopxl/beep)
Database       | [modernc.org/sqlite](https://gitlab.com/cznic/sqlite)
Metadata       | [dhowden/tag](https://github.com/dhowden/tag)

---

### Batasan

- M4A / AAC / Opus belum didukung — beep v2 tidak punya decoder-nya.
File dengan ekstensi ini akan di-skip saat scanning.

- Hanya Windows 10+ — dependency terbaru (Bubble Tea, SQLite) butuh
Go 1.24+, yang tidak jalan di Windows 8.1. Branch legacy-win8 untuk
dukungan Windows 8.1 belum dibuat.

- Tidak ada sort library — urutan selalu artist, album, title.

- Tidak ada mouse support — hanya keyboard.

- Tidak ada konfirmasi hapus folder — hati-hati saat menekan d di
mode Kelola Folder.

### Future Update

- [X] Playlist buat / edit / hapus
- [ ] Mouse Click Support
- [X] Shuffle mode
- [ ] Sort library
- [ ] Global Search
- [ ] Format M4A
- [ ] Konfirmasi hapus folder dan playlist
- [ ] Reorder track dalam playlist
- [ ] Info track (bitrate, sample rate, tahun, genre)
- [ ] Bulk add ke playlist

#### Belum dijadwalkan (bukan prioritas utama)
- [ ] Prevent sleep during playback
- [ ] File watcher (fsnotify) untuk auto-refresh library
- [ ] Cover art di terminal (chafa/sixel/kitty)
- [ ] Export / import M3U
- [ ] Format M4A (via ffmpeg subprocess)
- [ ] Refactor audio ke actor model
- [ ] Media session integration (MPRIS / SMTC / macOS Now Playing)

Lihat [CHANGELOG.md]() untuk riwayat perubahan per versi
