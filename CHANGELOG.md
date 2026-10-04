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

## [v0.3.1b] - 2026-10-03

Rilis patch. Fokus: stabilitas database, sinkronisasi library, dan
konsistensi hotkey.

### Diperbaiki

#### Database
- **`SQLITE_BUSY` saat scan folder besar** — SQLite default journal
  mode (`DELETE`) mengunci seluruh file DB, dan `database/sql`
  connection pool bisa memberi koneksi berbeda ke goroutine berbeda
  (scan, refresh, auto-count) sehingga saling kunci. Fix:
  `journal_mode(WAL)`, `busy_timeout(5000)`, `synchronous(NORMAL)`,
  dan `SetMaxOpenConns(1)`.
- **Deadlock di `DeleteTracksUnderFolder`** setelah single-connection
  pool — `defer rows.Close()` menahan satu-satunya koneksi sampai
  fungsi selesai, lalu `Begin()` menunggu koneksi yang tidak akan
  pernah tersedia. Fix: `rows.Close()` eksplisit sebelum `Begin()`.
- **Batch insert saat scan** — `UpsertTracks` membungkus seluruh
  metadata dalam satu transaction, mengurangi commit per file.

#### Sinkronisasi Library
- **Track yang sudah dihapus dari disk tetap muncul setelah rescan** —
  scan hanya melakukan UPSERT tanpa pernah DELETE track yang hilang.
  Fix: `DeleteTracksNotIn(folder, foundPaths)` dipanggil setelah tiap
  folder selesai di-scan. Aman: hanya memproses folder yang berhasil
  di-scan (folder yang error di-skip konservatif). Konsekuensi:
  FK CASCADE otomatis menghapus track yang sama dari semua playlist.

#### Resume State
- **Resume tidak disave saat `q` dari playlist** — di mode Playlists,
  Playlist Detail, Browser, Folders, dan Playlist Picker, `q`
  sebelumnya hanya kembali ke library, bukan quit. Fix: `q` sekarang
  konsisten = save resume + quit di semua mode (kecuali input text).
  `esc` = kembali/naik level.

#### Konsistensi Hotkey
- **`ctrl+c` sekarang bekerja di semua state** — sebelumnya tidak
  di-handle saat user berada di input text, help overlay, atau popup
  konfirmasi duplikat, sehingga user bisa "terjebak" tanpa cara
  force-quit. Sekarang `ctrl+c` di-intercept di paling atas
  `tea.KeyMsg` handler.
- **`ctrl+c` di help overlay tidak lagi save resume** — konsisten
  dengan semantik "force quit".
- **Semantik final hotkey:**
  - `ctrl+c` → force quit, tanpa save (bisa dari state apapun)
  - `q` → save resume + quit (di semua mode non-input)
  - `esc` → kembali/naik level; quit kalau di library

### Testing

- Semua test lama tetap PASS
- Test manual: scan folder besar dengan auto-refresh berjalan di
  latar belakang — tidak ada `SQLITE_BUSY`
- Test manual: hapus file MP3 dari disk, tekan `r` — track hilang
  dari library dan playlist
- Test manual: quit (`q`) dari setiap mode — resume tersimpan
  dengan source yang benar
- Test manual: `ctrl+c` dari input text dan popup konfirmasi —
  aplikasi langsung keluar tanpa save

### Catatan Teknis

- WAL menghasilkan file sampingan (`sky.db-wal`, `sky.db-shm`) saat
  aplikasi berjalan. Saat ditutup dengan bersih, SQLite otomatis
  checkpoint dan menghapus file sampingan. Ini normal.
- Satu koneksi (`SetMaxOpenConns(1)`) menserialisasi semua query.
  Untuk aplikasi TUI single-user, bottleneck tidak terasa.

  ## [v0.3.2b] - 2026-10-03

Rilis patch. Fokus: konsistensi hotkey global.

### Ditambahkan

- **Hotkey volume global (`+`/`=`, `-`/`_`) di semua mode** —
  sebelumnya hanya bekerja di Library. Sekarang bisa diatur dari
  Browser, Folders, Playlists, Playlist Detail, dan Playlist Picker.
  Prinsip: state global (volume, repeat, shuffle, resume) harus
  bisa diubah dari mode manapun tanpa harus kembali ke Library.

### Diubah

- Helper `adjustVolume(delta int)` di TUI — clamp, flash, dan
  persist ke config dalam satu tempat. Menggantikan dua blok
  duplikat di `updateLibrary`.
- Hotkey `+`/`-` dipindahkan dari dispatch lokal di `updateLibrary`
  ke global hotkey block (sejajar dengan `L`, `P`, `B`, `F`, `x`).
- Guard `isFiltering()` tetap berlaku: `+`/`-` saat user sedang
  mengetik filter akan masuk sebagai karakter ke kotak filter,
  bukan men-trigger volume.

### Catatan Teknis

- Volume tetap live-update tanpa interupsi playback
- Perubahan volume langsung di-`Save()` ke `config.json` (konsisten
  dengan `m` repeat dan `s` shuffle)
- Flash menunjukkan ikon `🔊` atau `🔇` (kalau volume 0%)

### Testing

- Test manual: ubah volume dari 5 mode berbeda — semua bekerja
- Test manual: `/` di library, ketik `+` — karakter masuk ke filter,
  volume tidak berubah
- Test manual: ubah volume, `q`, buka lagi — nilai tersimpan

## [v0.5b] - 2026-10-04

Rilis beta keempat. Fokus: konfirmasi hapus di semua aksi destruktif.

### Ditambahkan

#### Konfirmasi Hapus
Semua aksi destruktif sekarang memunculkan popup konfirmasi sebelum
eksekusi. Mencegah data loss karena salah pencet.

- **Hapus folder** (`d` di mode Folders) — popup menampilkan path
  folder + pesan "File asli di disk TIDAK dihapus" untuk menghindari
  kesalahpahaman user awam.
- **Hapus playlist** (`d` di mode Playlists) — popup menampilkan nama
  playlist + pesan "Track di dalamnya tidak terpengaruh".
- **Hapus track dari playlist** (`d` di mode Playlist Detail) — popup
  menampilkan judul track + pesan "Track tetap ada di library —
  hanya dihapus dari playlist ini". Konsisten dengan permintaan:
  dua tipe user (sengaja & tidak sengaja) harus dilindungi.
- **Tambah track duplikat** ke playlist (existing, dari v0.3b) —
  tetap bekerja, sekarang lewat dispatcher yang sama.

Semua popup punya hotkey yang konsisten:
- `y` → konfirmasi
- `n` / `esc` → batal, kembali ke mode sebelumnya
- `q` → save resume + quit
- `ctrl+c` → force quit

#### Generalisasi State Konfirmasi
- Enum `confirmKind` dengan lima nilai:
  `confirmNone`, `confirmDuplicate`, `confirmDeleteFolder`,
  `confirmDeletePlaylist`, `confirmDeleteTrackFromPlaylist`.
- Field baru di `Model`: `confirmFolderPath`, `confirmTrackIndex`,
  `confirmTrackTitle` (di samping `confirmPlaylistID` dan
  `confirmPlaylistName` yang sudah ada).
- Helper `clearConfirm()` — reset semua state popup + kembalikan user
  ke mode yang sesuai dengan konteks popup (bukan selalu ke Library).
- Dispatcher `handleConfirmYes()` — route ke handler spesifik
  berdasarkan `confirmKind`.
- Handler terpisah:
  - `handleConfirmDeleteFolderYes()`
  - `handleConfirmDeletePlaylistYes()`
  - `handleConfirmDeleteTrackFromPlaylistYes()`

### Diperbaiki

- **Popup konfirmasi tidak muncul** — blok render overlay salah
  tempat di dalam `viewHelp()`, bukan di `View()`. Karena `viewHelp()`
  hanya dipanggil saat `m.showHelp == true`, popup tidak pernah
  dirender meskipun state-nya benar. Fix: pindahkan blok ke awal
  `View()`, sebelum cek `showHelp`.
- **Folder langsung terhapus sebelum popup muncul** — di
  `updateFolders` case `d`, kode hapus lama masih tertinggal setelah
  set `m.confirmKind`. Akibatnya folder terhapus duluan, dan popup
  konfirmasi menunjuk ke folder yang sudah tidak ada. Fix: case `d`
  hanya set state popup, logika hapus dipindah ke handler `y`.
- **Playlist langsung terhapus sebelum popup muncul** — bug yang sama
  di `updatePlaylists`. Fix: sama.
- **Track dari playlist langsung terhapus sebelum popup muncul** —
  bug yang sama lagi di `updatePlaylistDetail`. Fix: sama.
- **`clearConfirm()` selalu kembali ke Library** — akibatnya user
  yang membatalkan hapus folder/playlist terlempar ke Library,
  padahal harusnya tetap di mode asalnya. Fix: kontekstual berdasarkan
  `confirmKind`.

### Diubah

- `d` di mode Folders, Playlists, dan Playlist Detail sekarang
  membuka popup, bukan langsung eksekusi.
- Pesan popup hapus folder eksplisit menyebut file asli tidak
  dihapus — supaya user tidak panik.
- Hapus track dari playlist menggunakan `confirmTrackIndex` (by
  posisi), bukan `track_id` — karena playlist mengizinkan duplikat,
  index memastikan hanya instance yang disorot yang terhapus.

### Catatan Teknis

- Popup overlay menggunakan `helpBoxStyle` yang sama dengan help
  overlay — border rounded abu-abu. Konsisten di semua popup.
- `handleConfirmYes()` selalu fallback ke `clearConfirm()` kalau
  `confirmKind` tidak dikenal — defensive programming.
- Bug pola berulang tiga kali (folder, playlist, track) karena saat
  refactor dari "langsung hapus" ke "konfirmasi dulu", kode hapus
  lama tidak dihapus. Pelajaran untuk refactor berikutnya: selalu
  grep untuk pemanggilan `RemoveFolder`, `DeletePlaylist`,
  `RemoveTrackFromPlaylistAt` di luar handler `handleConfirm*Yes`.

### Testing

- Test manual: hapus folder → popup → `n` batal → folder masih ada,
  tetap di mode Folders.
- Test manual: hapus folder → popup → `y` → folder hilang, tetap
  di mode Folders.
- Test manual: hapus playlist → popup → `esc` batal → tetap di mode
  Playlists.
- Test manual: hapus track dari playlist → popup → `n` batal → track
  masih ada, tetap di mode Playlist Detail.
- Test manual: hapus track dari playlist → popup → `y` → track
  hilang dari playlist, tetap ada di library.
- Test manual: duplikat track → popup → `n` batal → kembali ke
  Library (perilaku asli).
- Test manual: `ctrl+c` dari semua popup → force quit tanpa save.
- Test manual: `q` dari semua popup → save resume + quit.
- Unit test: semua PASS (tidak ada perubahan DB layer).