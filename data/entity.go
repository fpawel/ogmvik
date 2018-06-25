package data

import (
	"encoding/binary"
	"io"
	"log"
	"time"
)

type Entity struct {
	ID            uint64
	Order         uint64
	Time          time.Time
	DocCode       byte
	ProductsCount uint64
	RouteSheet    string
}

type Filter struct {
	Year               int
	Month              time.Month
	DocCode            byte
	RouteSheetMask     string
	OrderFrom, OrderTo uint64
}

func NewFilter() Filter {
	return Filter{
		Year: time.Now().Year(),
	}
}

func readBytes(x io.Reader, b []byte) {
	n, err := x.Read(b)
	if n != len(b) {
		log.Panicln(n, "!=len(b)==", len(b))
	}
	if err != nil && err != io.EOF {
		log.Panic(err)
	}
}

func readUInt64(x io.Reader) uint64 {
	b := make([]byte, 8)
	readBytes(x, b)
	return binary.BigEndian.Uint64(b)
}

func readString(x io.Reader) string {
	n := readUInt64(x)
	b := make([]byte, int(n))
	readBytes(x, b)
	return string(b)
}

func writeBytes(x io.Writer, b []byte) {
	n, err := x.Write(b)
	check(err)
	if n != len(b) {
		log.Fatal("n!=len(b)")
	}
}

func writeUInt64(x io.Writer, v uint64) {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, v)
	writeBytes(x, b)
}

func writeEntity(x io.Writer, a Entity) {
	writeUInt64(x, a.Order)
	writeUInt64(x, uint64(a.Time.UnixNano()))
	writeBytes(x, []byte{a.DocCode})
	writeUInt64(x, a.ProductsCount)
	writeUInt64(x, uint64(len(a.RouteSheet)))
	writeBytes(x, []byte(a.RouteSheet))
}

func readEntity(k []byte, x io.Reader) (a Entity) {
	a.ID = binary.BigEndian.Uint64(k)
	a.Order = readUInt64(x)
	a.Time = time.Unix(0, int64(readUInt64(x)))
	b := []byte{0}
	readBytes(x, b)
	a.DocCode = b[0]
	a.ProductsCount = readUInt64(x)
	a.RouteSheet = readString(x)
	return a
}
