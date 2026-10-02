# Changelog

Semua perubahan penting pada Sky didokumentasikan di file ini.

Format mengikuti [Keep a Changelog](https://keepachangelog.com/),
versi mengikuti [Semantic Versioning](https://semver.org/).

## [v0.1b] - 2026-09-29

Rilis beta pertama. Fondasi music player TUI.

### Ditambahkan

#### Fondasi Project
- Go module `github.com/ade-nugraha306/sky-project`
- Struktur modular: `internal/{config,db,library,player,tui}`
- Pure Go, tanpa CGO — cross-compile tanpa setup compiler tambahan

#### Library & Filesystem
- Scan folder musik rekursif, cross-platform Windows/Unix
- Skip folder tersembunyi otomatis (prefix `.`)
- Support ekstensi `.mp3`, `.flac`, `.ogg`, `.wav` di scanner
- Parse metadata via `dhowden/tag` (ID3v1/v2, Vorbis comment, MP4 atom)
- Fallback ke nama file kalau metadata kosong

#### Database
- SQLite pure Go (`modernc.org/sqlite`)
- Tabel `tracks` dengan `path UNIQUE`
- `UpsertTrack` idempoten — rescan tidak duplikat
- `AllTracks`, `CountTracks`, `DeleteTracksUnderFolder` (prefiks-match, transaction)

#### Audio Playback
- Decoder dinamis per ekstensi: MP3, FLAC, OGG, WAV
- Resampling ke 44.1 kHz — pitch konsisten antar sample rate
- Pause/resume race-safe dengan `speaker.Lock()`
- Seek ±5 detik dengan recreate resampler (buang buffer lama)
- Volume logaritmik (log2), live update tanpa reload
- Auto-close file handle lama saat ganti lagu

#### TUI
- Mode Library: list track, filter `/`, progress bar, now playing bar
- Mode Browser: navigasi folder + preview isi (split view 40/60)
- Mode Folders: kelola folder musik terdaftar
- Help overlay (`?`) dengan semua hotkey per kategori
- Toast notification 3 detik untuk aksi sementara

#### Playback Features
- Play dari library (queue = visible list, termasuk hasil filter)
- Next/prev dengan restart threshold 3 detik
- Auto-next saat lagu selesai
- Repeat: off / one / all — persist di config
- Seek ±5 detik
- Volume ±5% — persist di config

#### Resume State
- Simpan path + posisi (ms) saat keluar
- Resume track + seek ke posisi
- Resume selalu dalam keadaan pause
- Track hilang dari disk → resume di-skip, config di-clear

#### Config
- Auto-generate config default di `~/.config/sky/config.json`
- Normalisasi path ke forward slash (JSON-safe, portable)
- Pesan error informatif kalau config rusak

### Diperbaiki

- Pitch naik/turun antar lagu → resampling ke target tetap
- File descriptor bocor saat ganti lagu → `Close()` streamer lama
- JSON escape error di Windows (`\U` di path) → normalisasi path
- Play dari hasil filter salah lagu → `VisibleItems()` + lookup by path
- "queue selesai" nyangkut di status bar → pakai toast flash

### Catatan Teknis

- Target sample rate 44.1 kHz untuk semua playback
- Backward compat: config lama dengan path backslash tetap bisa dibaca,
  dinormalisasi ke forward slash saat `Save()` berikutnya

---

## [v0.3b] - 2026-10-02

Rilis beta kedua. Fokus: playlist lengkap, shuffle, navigasi antar mode.

### Ditambahkan

#### Playlist
- CRUD playlist: buat, rename, hapus
- Mode TUI baru: Playlists, Playlist Picker, Playlist Detail
- Tambah track ke playlist dari library (`t`)
- Popup konfirmasi kalau track sudah ada di playlist (duplikat diizinkan)
- Hapus track dari playlist (dari detail)
- Play dari playlist — tetap di detail, tidak kembali ke library
- Progress bar + now playing bar di detail playlist
- Schema DB baru:
  - Tabel `playlists` (id, name unique, created_at)
  - Tabel `playlist_tracks` (id, playlist_id, track_id, position, added_at)
  - FK CASCADE: delete playlist → track di dalamnya ikut terhapus;
    delete track → hilang dari semua playlist
- Rename playlist aman by design — `UPDATE` hanya menyentuh kolom `name`,
  `playlist_tracks.playlist_id` tidak berubah

#### Shuffle
- Mode shuffle dengan algoritma Fisher-Yates (unbiased)
- Badge `[🔀]` di progress bar
- Persist di config
- Reshuffle otomatis saat mencapai akhir queue
- Toggle on/off di tengah playback — track yang sedang diputar tidak terputus

#### Navigasi & Sumber Pemutaran
- Label sumber di now playing bar: `[Library]` atau `[NamaPlaylist]`
- Prefix `▶` di list — kontekstual, hanya muncul di list yang menjadi sumber
- Global hotkey: `L` (Library), `P` (Playlists), `B` (Browser),
  `F` (Folders), `x` (clear filter)
- Guard filter: hotkey global tidak intercept saat user sedang mengetik filter
- Resume ingat sumber pemutaran (Library vs Playlist) antar sesi
- Fallback ke Library kalau playlist sudah dihapus

#### Package Baru
- Package `queue`: `Shuffle` (Fisher-Yates), `FindIndex` —
  reusable untuk playlist dan komponen lain

#### Config
- Field `LastPlaylistID` (0 = library, > 0 = playlist)
- Field `Shuffle` (bool, persist antar sesi)

### Diperbaiki

- **Repeat one/all tidak jalan di playlist** — guard `if m.mode != modeLibrary`
  di `tickMsg` dihapus; `handleTrackEnd` sekarang dipanggil di semua mode
  yang memutar audio
- **Race condition di `handleTrackEnd`** — `currentTrack` di-reset sebelum
  `playAt` untuk mencegah tick berikutnya memanggil handler lagi sebelum
  `playResultMsg` masuk
- **Blok kode `tracksLoadedMsg` ter-copy dua kali** — pisahkan blok isi `items`
  dan blok resume; sebelumnya loop resume keburu dipasang sebelum `tracks`
  dideklarasikan
- **Status "memuat playlist..." nyangkut** — set status ke ringkasan jumlah
  playlist setelah load selesai
- **`playlistPicker` kosong** — `SetSize` di `WindowSizeMsg`, sebelumnya
  width/height = 0 sehingga list render kosong meski data ada
- **`modePlaylists` tidak ada di dispatch switch** — akibatnya `n` di mode
  playlist malah trigger play-next; tambah case di dispatch
- **`playlistTracksMsg` masih dipakai** — digantikan `playlistDetailMsg`
  yang membawa playlistID + nama, bukan hanya nama + tracks

### Diubah

- `config.SetResume` sekarang terima 3 parameter (path, pos, playlistID)
- `config.ClearResume` reset `LastPlaylistID` juga
- Hapus `LastPaused` dari config — legacy, resume selalu pause
- Hapus tabel `folders` dari schema — tidak pernah dipakai,
  sumber folder ada di config JSON
- Hapus `itoa` manual, ganti `fmt.Sprintf`
- Di playlist detail, `p` = previous (sebelumnya play-all)
- `playAt` di playlist detail tidak lagi keluar ke library

### Testing

- Folder `tests/` dengan 4 file test terpisah dari `internal/`
- ~58 test PASS: config, db, library, queue, playlist
- Test baru untuk:
  - `RemoveTrackFromPlaylistAt` — hapus by position, duplikat di-handle benar
  - FK CASCADE: delete playlist → track hilang; delete track → hilang dari playlist
  - `TrackInPlaylist` — exists/not-exists/wrong-playlist
  - `LastPlaylistID` persist round-trip + `omitempty` behavior
- Test `SetResume` di-update untuk signature baru (3 argumen)

### Catatan Teknis

- FK CASCADE diaktifkan via `?_pragma=foreign_keys(1)` di DSN SQLite —
  default SQLite tidak aktif
- `isUniqueConstraintError` deteksi error via string matching pesan driver —
  fragile, tapi berfungsi untuk `modernc.org/sqlite`
- Rename playlist aman by design (FK refer ke `id`, bukan `name`)
- Resume selalu pause — konsisten dengan v0.1b
