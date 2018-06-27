package data

import (
	"bytes"
	"encoding/binary"
	"github.com/boltdb/bolt"
	"log"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"
	"github.com/fpawel/ogmvik"
)

type DB struct {
	mx *sync.Mutex
	db *bolt.DB
}

func root(tx *bolt.Tx) *bolt.Bucket {
	return tx.Bucket([]byte("root"))
}

func (x DB) Close() error {
	return x.db.Close()
}

func Open() (x DB) {
	x.mx = new(sync.Mutex)
	var err error
	x.db, err = bolt.Open(ogmvik.AppDataFileName("data.db"), 0600, nil)
	check(err)
	x.update(func(tx *bolt.Tx) {
		buck := root(tx)
		if buck == nil {
			tx.CreateBucket([]byte("root"))
		}
	})
	return
}

func (x DB) Records(f Filter) (xs []Entity) {
	x.view(func(tx *bolt.Tx) {
		xs = x.records(tx, f)
	})
	return
}

func (x DB) SetOrder(id uint64, order uint64) {
	x.setByID(id, func(entity *Entity) {
		entity.Order = order
	})
}

func (x DB) SetDocCode(id uint64, docCode byte) {
	x.setByID(id, func(entity *Entity) {
		entity.DocCode = docCode
	})
}

func (x DB) SetProductsCount(id uint64, productsCount uint64) {
	x.setByID(id, func(entity *Entity) {
		entity.ProductsCount = productsCount
	})
}

func (x DB) SetRouteSheet(id uint64, routeSheet string) {
	x.setByID(id, func(entity *Entity) {
		entity.RouteSheet = routeSheet
	})
}

func (x DB) setByID(id uint64, f func(*Entity)) {
	x.mx.Lock()
	defer x.mx.Unlock()
	x.update(func(tx *bolt.Tx) {
		b := root(tx)
		c := b.Cursor()
		for k, v := c.First(); k != nil; k, v = c.Next() {
			a := readEntity(k, bytes.NewReader(v))
			if a.ID == id {
				f(&a)
				buf := bytes.NewBuffer(nil)
				writeEntity(buf, a)
				check(b.Put(k, buf.Bytes()))
			}
		}
	})
	return
}

func (x DB) Add(a Entity) (ret uint64) {
	x.mx.Lock()
	defer x.mx.Unlock()
	x.update(func(tx *bolt.Tx) {
		ret = x.add(a, root(tx))
	})
	return
}

func (x DB) Delete(id uint64) {
	x.mx.Lock()
	defer x.mx.Unlock()
	k := make([]byte, 8)
	binary.BigEndian.PutUint64(k, id)
	x.update(func(tx *bolt.Tx) {
		check(root(tx).Delete(k))
	})
	return
}

func (x DB) addRange(xs []Entity) {
	x.mx.Lock()
	defer x.mx.Unlock()
	x.update(func(tx *bolt.Tx) {
		for _, a := range xs {
			x.add(a, root(tx))
		}
	})
	return
}

func (x DB) records(tx *bolt.Tx, f Filter) (xs []Entity) {
	c := root(tx).Cursor()
	for k, v := c.First(); k != nil; k, v = c.Next() {
		a := readEntity(k, bytes.NewReader(v))

		if (f.Year == 0 || a.Time.Year() == f.Year) &&
			(f.Month == 0 || a.Time.Month() == f.Month) &&
			(f.DocCode == 0 || a.DocCode == f.DocCode) &&
			(f.OrderFrom == 0 || a.Order >= f.OrderFrom) &&
			(f.OrderTo == 0 || a.Order <= f.OrderTo) &&
			(len(f.RouteSheetMask) == 0 || strings.Contains(a.RouteSheet, f.RouteSheetMask)) {
			xs = append(xs, a)
		}
	}
	sort.Slice(xs, func(i, j int) bool {
		a, b := xs[i], xs[j]
		ta, tb := a.Time.Truncate(24*time.Hour), b.Time.Truncate(24*time.Hour)
		if ta == tb {
			return a.Order < b.Order
		}
		return a.Time.Before(b.Time)
	})
	return
}

func (x DB) add(a Entity, b *bolt.Bucket) uint64 {

	if a.RouteSheet == "" {
		log.Fatal("RouteSheet must be set")
	}

	a.ID, _ = b.NextSequence()
	buf := bytes.NewBuffer(nil)
	writeEntity(buf, a)

	k := make([]byte, 8)
	binary.BigEndian.PutUint64(k, a.ID)

	check(b.Put(k, buf.Bytes()))
	return a.ID
}

func (x DB) update(f func(*bolt.Tx)) {
	x.db.Update(func(tx *bolt.Tx) error {
		f(tx)
		return nil
	})
}

func (x DB) view(f func(*bolt.Tx)) {
	x.db.View(func(tx *bolt.Tx) error {
		f(tx)
		return nil
	})
}

func check(err error) {
	if err != nil {
		_, file, line, _ := runtime.Caller(1)
		log.Panicf("%s:%d %v\n", filepath.Base(file), line, err)
	}
}
