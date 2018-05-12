package pinaf

import (
	"bytes"
	"encoding/json"

	"github.com/syndtr/goleveldb/leveldb"
)

type JSONKey struct {
	Key
}

func (k JSONKey) Put(b *leveldb.Batch, subKey []byte, value interface{}) error {
	var buf bytes.Buffer
	err := json.NewEncoder(&buf).Encode(value)
	if err == nil {
		b.Put(k.Entry(subKey), buf.Bytes())
	}
	return err
}

func (k JSONKey) Get(db *leveldb.DB, subKey []byte, value interface{}) error {
	buf, err := db.Get(k.Entry(subKey), nil)
	if err == nil {
		err = json.NewDecoder(bytes.NewReader(buf)).Decode(value)
	}
	return err
}
