//go:build pebbledb

package db

import (
	"bytes"
	"fmt"
	"path/filepath"

	"github.com/cockroachdb/pebble"
)

func init() {
	dbCreator := func(name string, dir string) (DB, error) {
		return NewPebbleDB(name, dir)
	}
	registerDBCreator(PebbleDBBackend, dbCreator, false)
}

// PebbleDB is a PebbleDB backend.
type PebbleDB struct {
	db *pebble.DB
}

var _ DB = (*PebbleDB)(nil)

func NewPebbleDB(name string, dir string) (DB, error) {
	dbPath := filepath.Join(dir, name+".db")
	//	cache := pebble.NewCache(1024 * 1024 * 32)
	//	defer cache.Unref()
	opts := &pebble.Options{
		//		Cache:                       cache,
		//		FormatMajorVersion:          pebble.FormatNewest,
		//		L0CompactionThreshold:       2,
		//		L0StopWritesThreshold:       1000,
		//		LBaseMaxBytes:               64 << 20, // 64 MB
		//		Levels:                      make([]pebble.LevelOptions, 7),
		//		MaxConcurrentCompactions:    3,
		//		MaxOpenFiles:                1024,
		//		MemTableSize:                64 << 20,
		//		MemTableStopWritesThreshold: 4,
	}
	/*
		for i := 0; i < len(opts.Levels); i++ {
			l := &opts.Levels[i]
			l.BlockSize = 32 << 10       // 32 KB
			l.IndexBlockSize = 256 << 10 // 256 KB
			l.FilterPolicy = bloom.FilterPolicy(10)
			l.FilterType = pebble.TableFilter
			if i > 0 {
				l.TargetFileSize = opts.Levels[i-1].TargetFileSize * 2
			}
			l.EnsureDefaults()
		}
	*/
	//	opts.Levels[6].FilterPolicy = nil
	//	opts.FlushSplitBytes = opts.Levels[0].TargetFileSize

	for i := 0; i < len(opts.Levels); i++ {
		l := &opts.Levels[i]
		l.Compression = pebble.NoCompression
		l.EnsureDefaults()
	}

	opts.EnsureDefaults()

	p, err := pebble.Open(dbPath, opts)
	if err != nil {
		return nil, err
	}
	return &PebbleDB{
		db: p,
	}, err
}

// Get implements DB.
func (db *PebbleDB) Get(key []byte) ([]byte, error) {
	fmt.Println("PebbleDB.Get")
	if len(key) == 0 {
		return nil, errKeyEmpty
	}
	res, closer, err := db.db.Get(key)
	if err != nil {
		return res, nil
	}
	closer.Close()
	return res, nil
}

// Has implements DB.
func (db *PebbleDB) Has(key []byte) (bool, error) {
	fmt.Println("PebbleDB.Has")
	if len(key) == 0 {
		return false, errKeyEmpty
	}
	bytes, err := db.Get(key)
	if err != nil {
		return false, err
	}
	return bytes != nil, nil
}

// Set implements DB.
func (db *PebbleDB) Set(key []byte, value []byte) error {
	fmt.Println("PebbleDB.Set")
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	err := db.db.Set(key, value, pebble.NoSync)
	if err != nil {
		return err
	}
	return nil
}

// SetSync implements DB.
func (db *PebbleDB) SetSync(key []byte, value []byte) error {
	fmt.Println("PebbleDB.SetSync")
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	err := db.db.Set(key, value, pebble.Sync)
	if err != nil {
		return err
	}
	return nil
}

// Delete implements DB.
func (db *PebbleDB) Delete(key []byte) error {
	fmt.Println("PebbleDB.Delete")
	if len(key) == 0 {
		return errKeyEmpty
	}
	err := db.db.Delete(key, pebble.NoSync)
	if err != nil {
		return err
	}
	return nil
}

// DeleteSync implements DB.
func (db PebbleDB) DeleteSync(key []byte) error {
	fmt.Println("PebbleDB.DeleteSync")
	if len(key) == 0 {
		return errKeyEmpty
	}
	err := db.db.Delete(key, pebble.Sync)
	if err != nil {
		return nil
	}
	return nil
}

func (db *PebbleDB) DB() *pebble.DB {
	return db.db
}

// Close implements DB.
func (db PebbleDB) Close() error {
	fmt.Println("PebbleDB.Close")
	db.db.Close()
	return nil
}

// Print implements DB.
func (db *PebbleDB) Print() error {
	itr, err := db.Iterator(nil, nil)
	if err != nil {
		return err
	}
	defer itr.Close()
	for ; itr.Valid(); itr.Next() {
		key := itr.Key()
		value := itr.Value()
		fmt.Printf("[%X]:\t[%X]\n", key, value)
	}
	return nil
}

// Stats implements DB.
func (db *PebbleDB) Stats() map[string]string {
	/*
		keys := []string{"rocksdb.stats"}
		stats := make(map[string]string, len(keys))
		for _, key := range keys {
			stats[key] = db.(key)
		}
	*/
	return nil
}

// NewBatch implements DB.
func (db *PebbleDB) NewBatch() Batch {
	return newPebbleDBBatch(db)
}

// Iterator implements DB.
func (db *PebbleDB) Iterator(start, end []byte) (Iterator, error) {
	fmt.Println("PebbleDB.Iterator")
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	o := pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	}
	itr := db.db.NewIter(&o)
	itr.First()

	return newPebbleDBIterator(itr, start, end, false), nil
}

// ReverseIterator implements DB.
func (db *PebbleDB) ReverseIterator(start, end []byte) (Iterator, error) {
	fmt.Println("PebbleDB.ReverseIterator")
	if (start != nil && len(start) == 0) || (end != nil && len(end) == 0) {
		return nil, errKeyEmpty
	}
	o := pebble.IterOptions{
		LowerBound: start,
		UpperBound: end,
	}
	itr := db.db.NewIter(&o)
	itr.Last()
	return newPebbleDBIterator(itr, start, end, true), nil
}

var _ Batch = (*pebbleDBBatch)(nil)

type pebbleDBBatch struct {
	db    *PebbleDB
	batch *pebble.Batch
}

var _ Batch = (*pebbleDBBatch)(nil)

func newPebbleDBBatch(db *PebbleDB) *pebbleDBBatch {
	return &pebbleDBBatch{
		batch: db.db.NewBatch(),
	}
}

// Set implements Batch.
func (b *pebbleDBBatch) Set(key, value []byte) error {
	fmt.Println("pebbleDBBatch.Set")
	if len(key) == 0 {
		return errKeyEmpty
	}
	if value == nil {
		return errValueNil
	}
	if b.batch == nil {
		return errBatchClosed
	}
	b.batch.Set(key, value, nil)
	return nil
}

// Delete implements Batch.
func (b *pebbleDBBatch) Delete(key []byte) error {
	fmt.Println("pebbleDBBatch.Delete")
	if len(key) == 0 {
		return errKeyEmpty
	}
	if b.batch == nil {
		return errBatchClosed
	}
	b.batch.Delete(key, nil)
	return nil
}

// Write implements Batch.
func (b *pebbleDBBatch) Write() error {
	fmt.Println("pebbleDBBatch.Write")
	if b.batch == nil {
		return errBatchClosed
	}
	err := b.batch.Commit(pebble.NoSync)
	if err != nil {
		return err
	}
	// Make sure batch cannot be used afterwards. Callers should still call Close(), for errors.

	return b.Close()
}

// WriteSync implements Batch.
func (b *pebbleDBBatch) WriteSync() error {
	fmt.Println("pebbleDBBatch.WriteSync")
	if b.batch == nil {
		return errBatchClosed
	}
	err := b.batch.Commit(pebble.Sync)
	if err != nil {
		return err
	}
	// Make sure batch cannot be used afterwards. Callers should still call Close(), for errors.
	return b.Close()
}

// Close implements Batch.
func (b *pebbleDBBatch) Close() error {
	fmt.Println("pebbleDBBatch.Close")
	if b.batch != nil {
		err := b.batch.Close()
		if err != nil {
			return err
		}
		b.batch = nil
	}

	return nil
}

type pebbleDBIterator struct {
	source     *pebble.Iterator
	start, end []byte
	isReverse  bool
	isInvalid  bool
}

var _ Iterator = (*pebbleDBIterator)(nil)

func newPebbleDBIterator(source *pebble.Iterator, start, end []byte, isReverse bool) *pebbleDBIterator {
	if isReverse {
		if end == nil {
			source.Last()
		}
	} else {
		if start == nil {
			source.First()
		}
	}
	return &pebbleDBIterator{
		source:    source,
		start:     start,
		end:       end,
		isReverse: isReverse,
		isInvalid: false,
	}
}

// Domain implements Iterator.
func (itr *pebbleDBIterator) Domain() ([]byte, []byte) {
	fmt.Println("pebbleDBIterator.Domain")
	return itr.start, itr.end
}

// Valid implements Iterator.
func (itr *pebbleDBIterator) Valid() bool {
	fmt.Println("pebbleDBIterator.Valid")
	// Once invalid, forever invalid.
	if itr.isInvalid {
		return false
	}

	// If source has error, invalid.
	if err := itr.source.Error(); err != nil {
		itr.isInvalid = true

		return false
	}

	// If source is invalid, invalid.
	if !itr.source.Valid() {
		itr.isInvalid = true

		return false
	}

	// If key is end or past it, invalid.
	start := itr.start
	end := itr.end
	key := itr.source.Key()
	if itr.isReverse {
		if start != nil && bytes.Compare(key, start) < 0 {
			itr.isInvalid = true

			return false
		}
	} else {
		if end != nil && bytes.Compare(end, key) <= 0 {
			itr.isInvalid = true

			return false
		}
	}

	// It's valid.
	return true
}

// Key implements Iterator.
func (itr *pebbleDBIterator) Key() []byte {
	fmt.Println("pebbleDBIterator.Key")
	itr.assertIsValid()
	return itr.source.Key()
}

// Value implements Iterator.
func (itr *pebbleDBIterator) Value() []byte {
	fmt.Println("pebbleDBIterator.Value")
	itr.assertIsValid()
	return itr.source.Value()
}

// Next implements Iterator.
func (itr pebbleDBIterator) Next() {
	fmt.Println("pebbleDBIterator.Next")
	itr.assertIsValid()
	if itr.isReverse {
		itr.source.Prev()
	} else {
		itr.source.Next()
	}
}

// Error implements Iterator.
func (itr *pebbleDBIterator) Error() error {
	return itr.source.Error()
}

// Close implements Iterator.
func (itr *pebbleDBIterator) Close() error {
	fmt.Println("pebbleDBIterator.Close")
	err := itr.source.Close()
	if err != nil {
		return err
	}
	return nil
}

func (itr *pebbleDBIterator) assertIsValid() {
	if !itr.Valid() {
		panic("iterator is invalid")
	}
}
