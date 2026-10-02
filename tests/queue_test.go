package tests

import (
	"sort"
	"testing"

	"github.com/ade-nugraha306/sky-project/internal/library"
	"github.com/ade-nugraha306/sky-project/internal/queue"
)

func makeTracks(paths ...string) []library.Track {
	out := make([]library.Track, len(paths))
	for i, p := range paths {
		out[i] = library.Track{Path: p, Title: p}
	}
	return out
}

func TestShuffle_PreservesElements(t *testing.T) {
	in := makeTracks("a", "b", "c", "d", "e", "f", "g", "h")
	out := queue.Shuffle(in)

	if len(out) != len(in) {
		t.Fatalf("length changed: got %d, want %d", len(out), len(in))
	}

	// Sortir kedua slice by path, bandingkan.
	inCopy := make([]library.Track, len(in))
	copy(inCopy, in)
	sort.Slice(inCopy, func(i, j int) bool { return inCopy[i].Path < inCopy[j].Path })

	outCopy := make([]library.Track, len(out))
	copy(outCopy, out)
	sort.Slice(outCopy, func(i, j int) bool { return outCopy[i].Path < outCopy[j].Path })

	for i := range inCopy {
		if inCopy[i].Path != outCopy[i].Path {
			t.Errorf("element %d: got %s, want %s", i, outCopy[i].Path, inCopy[i].Path)
		}
	}
}

func TestShuffle_DoesNotMutateOriginal(t *testing.T) {
	in := makeTracks("a", "b", "c", "d", "e")
	original := make([]library.Track, len(in))
	copy(original, in)

	_ = queue.Shuffle(in)

	for i := range in {
		if in[i].Path != original[i].Path {
			t.Errorf("original mutated at index %d: got %s, want %s",
				i, in[i].Path, original[i].Path)
		}
	}
}

func TestShuffle_Empty(t *testing.T) {
	out := queue.Shuffle(nil)
	if len(out) != 0 {
		t.Errorf("expected empty, got %d", len(out))
	}
}

func TestFindIndex_Found(t *testing.T) {
	tracks := makeTracks("a", "b", "c")
	idx := queue.FindIndex(tracks, "b")
	if idx != 1 {
		t.Errorf("got %d, want 1", idx)
	}
}

func TestFindIndex_NotFound(t *testing.T) {
	tracks := makeTracks("a", "b", "c")
	idx := queue.FindIndex(tracks, "z")
	if idx != -1 {
		t.Errorf("got %d, want -1", idx)
	}
}

func TestFindIndex_Empty(t *testing.T) {
	idx := queue.FindIndex(nil, "a")
	if idx != -1 {
		t.Errorf("got %d, want -1", idx)
	}
}