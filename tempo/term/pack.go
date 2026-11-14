package term

import (
	"encoding/binary"
	"math/big"

	"github.com/go-muse/musex/common/fraction"
)

// A list is stored as one immutable string, its pack, which encodes the whole
// content canonically: lists with equal content have equal packs, so == on
// List and Term compares content without copying. The layout uses fixed
// 8-byte little-endian integers (u64) and strings prefixed with their length
// as a u64 (str):
//
//	header    u64 words, u64 places, u64 synonyms, u64 offset of the synonyms
//	offsets   u64 offset of each word record, in list order
//	records   per word: u64 place, u8 group, u8 guide flags,
//	          [str lower numerator, str lower denominator] if flag 1,
//	          [str upper numerator, str upper denominator] if flag 2,
//	          str name, str English meaning, str Russian meaning
//	synonyms  per synonym: str alias, u64 index of the word
//
// Guide bounds are stored reduced, as big-endian magnitudes of a positive
// numerator and denominator.

const (
	u64Size = 8

	headerWords    = 0
	headerPlaces   = u64Size
	headerSynonyms = 2 * u64Size
	headerSynStart = 3 * u64Size
	headerSize     = 4 * u64Size

	flagLower = 1
	flagUpper = 2
)

// synonym is an additional name that List.Parse accepts for the word with the
// given index.
type synonym struct {
	alias string
	index int
}

// packer builds a pack.
type packer struct {
	b []byte
}

func (p *packer) u64(v int) {
	p.b = binary.LittleEndian.AppendUint64(p.b, uint64(v))
}

func (p *packer) str(s string) {
	p.u64(len(s))
	p.b = append(p.b, s...)
}

// bound appends a positive reduced fraction as its numerator and denominator.
func (p *packer) bound(f fraction.Fraction) {
	num, den, _ := f.ReducedBigPair()
	p.str(string(num.Bytes()))
	p.str(string(den.Bytes()))
}

// putAt overwrites the u64 at offset off.
func (p *packer) putAt(off, v int) {
	binary.LittleEndian.PutUint64(p.b[off:], uint64(v))
}

// packWord is the content of one word before packing.
type packWord struct {
	place   int
	group   Group
	guide   *Guide
	name    string
	english string
	russian string
}

// pack encodes words in list order and the synonyms of the list.
func pack(words []packWord, places int, synonyms []synonym) string {
	var p packer
	p.u64(len(words))
	p.u64(places)
	p.u64(len(synonyms))
	p.u64(0) // offset of the synonyms, set below
	offsets := len(p.b)
	for range words {
		p.u64(0)
	}
	for i, w := range words {
		p.putAt(offsets+i*u64Size, len(p.b))
		p.u64(w.place)
		var flags byte
		if w.guide != nil {
			if w.guide.lower.IsValid() {
				flags |= flagLower
			}
			if w.guide.upper.IsValid() {
				flags |= flagUpper
			}
		}
		p.b = append(p.b, byte(w.group), flags)
		if flags&flagLower != 0 {
			p.bound(w.guide.lower)
		}
		if flags&flagUpper != 0 {
			p.bound(w.guide.upper)
		}
		p.str(w.name)
		p.str(w.english)
		p.str(w.russian)
	}
	p.putAt(headerSynStart, len(p.b))
	for _, s := range synonyms {
		p.str(s.alias)
		p.u64(s.index)
	}
	return string(p.b)
}

// reader reads a pack from an offset.
type reader struct {
	s   string
	off int
}

func (r *reader) u64() int {
	var v uint64
	for k := u64Size - 1; k >= 0; k-- {
		v = v<<8 | uint64(r.s[r.off+k])
	}
	r.off += u64Size
	return int(v)
}

func (r *reader) u8() byte {
	v := r.s[r.off]
	r.off++
	return v
}

func (r *reader) str() string {
	n := r.u64()
	v := r.s[r.off : r.off+n]
	r.off += n
	return v
}

// u64At returns the u64 at offset off of pack s.
func u64At(s string, off int) int {
	r := reader{s: s, off: off}
	return r.u64()
}

// record is a decoded word record. Its strings share the memory of the pack.
type record struct {
	place   int
	group   Group
	flags   byte
	lower   [2]string // numerator and denominator if flags&flagLower != 0
	upper   [2]string // numerator and denominator if flags&flagUpper != 0
	name    string
	english string
	russian string
}

// recordAt decodes the word with index i of the set pack s.
func recordAt(s string, i int) record {
	r := reader{s: s, off: u64At(s, headerSize+i*u64Size)}
	rec := record{place: r.u64()}
	rec.group = Group(r.u8())
	rec.flags = r.u8()
	if rec.flags&flagLower != 0 {
		rec.lower = [2]string{r.str(), r.str()}
	}
	if rec.flags&flagUpper != 0 {
		rec.upper = [2]string{r.str(), r.str()}
	}
	rec.name = r.str()
	rec.english = r.str()
	rec.russian = r.str()
	return rec
}

// unpackBound returns the fraction stored as a numerator and a denominator.
func unpackBound(nd [2]string) fraction.Fraction {
	num := new(big.Int).SetBytes([]byte(nd[0]))
	den := new(big.Int).SetBytes([]byte(nd[1]))
	f, _ := fraction.FromBigInt(num, den)
	return f
}

// guide returns the guide of the word and whether it has one.
func (rec record) guide() (Guide, bool) {
	if rec.flags == 0 {
		return Guide{}, false
	}
	var g Guide
	if rec.flags&flagLower != 0 {
		g.lower = unpackBound(rec.lower)
	}
	if rec.flags&flagUpper != 0 {
		g.upper = unpackBound(rec.upper)
	}
	return g, true
}

// synonymsOf returns the synonyms of the set pack s.
func synonymsOf(s string) []synonym {
	n := u64At(s, headerSynonyms)
	if n == 0 {
		return nil
	}
	r := reader{s: s, off: u64At(s, headerSynStart)}
	out := make([]synonym, n)
	for i := range out {
		out[i].alias = r.str()
		out[i].index = r.u64()
	}
	return out
}
