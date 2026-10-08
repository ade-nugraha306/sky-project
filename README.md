# Sky - Yet Another Music Player Project

> Music player CLI/TUI yang ringan, keyboard-first, cross-platform.

Sky adalah music player berbasis terminal yang ditulis dengan Go. Bisa scan
folder musik lokal, simpan metadata ke SQLite, dan memutar audio dengan
TUI yang responsif. Tidak butuh GUI, tidak butuh server — cukup terminal.

## Fitur

### Audio
- **Multi-format** — MP3, FLAC, OGG (Vorbis), WAV
- **Cross-platform** — Windows, Linux, macOS (pure Go, tanpa CGO)
- **Pitch konsisten** — semua track di-resample ke 44.1 kHz
- **Volume logaritmik** — penurunan terasa natural di telinga
- **Seek cepat** — mundur/maju 5 detik tanpa glitch

### Library
- **Scan folder rekursif** — cross-platform, skip folder tersembunyi
- **Metadata parsing** — ID3v1/v2, Vorbis comment, MP4 atom
- **Folder aktif** — tampilkan hanya track dari satu folder, ganti dengan `f`
- **Sort library** — title / artist / album / date added via hotkey `o`
- **Filter** — cari lagu by title/artist/album
- **Rescan** — perbarui library saat ada file baru
- **Sinkronisasi otomatis** — track yang dihapus dari disk ikut terhapus saat rescan

### Playlist
- **CRUD lengkap** — buat, rename, hapus
- **Tambah track** dari library dengan satu hotkey
- **Deteksi duplikat** — popup konfirmasi kalau track sudah ada
- **Hapus track** dari playlist tanpa menyentuh library
- **Reorder track** — `Shift+K` naik, `Shift+J` turun
- **Play dari playlist** — queue = isi playlist

### Playback
- **Queue** — play dari library atau playlist
- **Shuffle** — Fisher-Yates, reshuffle otomatis di akhir queue
- **Repeat** — off / one / all
- **Auto-next** — lanjut otomatis saat lagu selesai
- **Resume state** — ingat track, posisi, sumber, dan repeat mode
- **Reorder saat shuffle** — queue dan DB disinkronkan by path

### UI
- **Mini file browser** — tambah folder tanpa ketik path manual
- **Preview folder** — lihat isi folder sebelum ditambahkan (split view)
- **Kelola folder** — tambah, hapus, dengan track cleanup otomatis
- **Konfirmasi hapus** — folder, playlist, dan track dari playlist
- **Help overlay** — panduan hotkey lengkap dengan `?`
- **Toast notifications** — feedback singkat untuk setiap aksi
- **Kontekstual `▶`** — prefix hanya muncul di list yang jadi sumber pemutaran
- **Source label** — `[Library]` atau `[NamaPlaylist]` di now playing bar

## Instalasi

### Prasyarat

- Go 1.21 atau lebih baru
- Terminal yang mendukung ANSI colors (Windows Terminal, iTerm2,
  GNOME Terminal, Alacritty, WezTerm, dsb)

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
`+` / `=` | Volume naik 5%
`-` / `_` | Volume turun 5%
`?`	  | Buka panduan hotkey lengkap
`Ctrl+C` | Keluar paksa

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
`o`      | Ganti sort order: title → artist → album → date
`f`      | Ganti folder aktif
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
`Enter`     | Set sebagai folder aktif
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
`s`     | Toggle Shuffle
`Shift + K` | Pindah track ke atas
`Shift + J  | Pindah track ke bawah
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

### Konfigurasi
Contoh file config JSON di `~/.config/sky/config.json`
```json
{
  "music_folders": [
    "~/Music",
    "~/Music/Collections"
  ],
  "volume": 75,
  "last_track_path": "~/Music/Artist - Title.mp3",
  "last_position_ms": 76200,
  "repeat_mode": "all",
  "shuffle": false,
  "last_playlist_id": 5,
  "last_active_folder": "~/Music",
  "sort_mode": "artist"
}
```

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

- Sort dan filter di-disable saat shuffle on. Matikan shuffle dulu untuk mengaktifkan kembali.

- Tidak ada mouse support — hanya keyboard.

- Belum ada konfirmasi quit saat ada lagu diputar — menekan q
langsung keluar (resume disimpan).

### Future Update (Roadmap)
- [ ] Konfirmasi quit saat ada lagu yang diputar
- [ ] Global search (lintas library dan playlist)
- [ ] Mouse click support (bubblezone)
- [ ] Info track overlay (bitrate, sample rate, tahun)


#### Belum dijadwalkan (bukan prioritas utama)
- [ ] Prevent sleep during playback
- [ ] Media session integration (MPRIS / SMTC / macOS Now Playing)
- [ ] File watcher (fsnotify) untuk auto-refresh library
- [ ] Bulk add ke playlist (semua hasil filter sekaligus)
- [ ] Progress indicator saat scan folder besar
- [ ] Cover art di terminal (chafa/sixel/kitty)
- [ ] Export / import M3U
- [ ] Format M4A (via ffmpeg subprocess)
- [ ] Refactor audio ke actor model (prasyarat media session)

Lihat [CHANGELOG.md](./CHANGELOG.md) untuk riwayat perubahan per versi
