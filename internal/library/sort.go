package library

// SortMode menentukan urutan tampilan track di Library.
// Playlist tidak terpengaruh — selalu urut berdasarkan position.
type SortMode string

const (
	SortByTitle    SortMode = "title"
	SortByArtist   SortMode = "artist"
	SortByAlbum    SortMode = "album"
	SortByDateDesc SortMode = "date"
)

// Key mengembalikan representasi untuk persistensi (config JSON).
func (s SortMode) Key() string {
	return string(s)
}

// Label mengembalikan teks untuk display, dengan prefix "sort: ".
func (s SortMode) Label() string {
	switch s {
	case SortByArtist:
		return "sort: artist"
	case SortByAlbum:
		return "sort: album"
	case SortByDateDesc:
		return "sort: date"
	default:
		return "sort: title"
	}
}

// Next mengembalikan SortMode berikutnya dalam cycle:
// title → artist → album → date → title → ...
func (s SortMode) Next() SortMode {
	switch s {
	case SortByTitle:
		return SortByArtist
	case SortByArtist:
		return SortByAlbum
	case SortByAlbum:
		return SortByDateDesc
	default:
		return SortByTitle
	}
}

// ParseSortMode mengubah string dari config ke SortMode.
// Fallback ke SortByTitle kalau tidak dikenal.
func ParseSortMode(s string) SortMode {
	switch s {
	case "artist":
		return SortByArtist
	case "album":
		return SortByAlbum
	case "date":
		return SortByDateDesc
	default:
		return SortByTitle
	}
}