package pinaf

import (
	"bytes"

	"github.com/syndtr/goleveldb/leveldb"
	"github.com/syndtr/goleveldb/leveldb/iterator"
	"github.com/syndtr/goleveldb/leveldb/util"
)

const (
	// key-path separators
	path  = '|'
	entry = '@'
)

type Key struct {
	prefix []byte
}

func New(element string, elements ...string) Key {
	var buf bytes.Buffer
	buf.WriteString(element)
	for _, p := range elements {
		buf.WriteRune(path)
		buf.WriteString(p)
	}
	return Key{prefix: buf.Bytes()}
}

func (k Key) subpath(sep rune, subKey []byte) []byte {
	buf := bytes.NewBuffer(k.prefix)
	buf.WriteRune(sep)
	buf.Write(subKey)
	return buf.Bytes()
}

func (k Key) SubKey(subKey []byte) Key {
	return Key{k.subpath(path, subKey)}
}

func (k Key) Entry(entryKey []byte) []byte {
	return k.subpath(entry, entryKey)
}

func (k Key) Scan(s *leveldb.DB) Iterator {
	pfx := k.subpath(entry, nil)
	return Iterator{
		Iterator: s.NewIterator(util.BytesPrefix(pfx), nil),
		prefix:   pfx,
	}
}

type Iterator struct {
	iterator.Iterator
	prefix []byte
}

func (i Iterator) Key() []byte {
	return i.Iterator.Key()[len(i.prefix):]
}

func (i Iterator) Delete(b *leveldb.Batch) {
	b.Delete(i.Iterator.Key())
}

func (i Iterator) Seek(key []byte) bool {
	var buf bytes.Buffer
	buf.Write(i.prefix)
	buf.Write(key)
	return i.Iterator.Seek(buf.Bytes())
}

