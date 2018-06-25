package main

import (
	"github.com/fpawel/ogmvik/data"
	"log"
	"time"

	"github.com/fpawel/procmq"
	"strconv"
)

type App struct {
	dir     string
	db      data.DB
	filter  data.Filter
	records []data.Entity
}

func (x *App) onCellEdited(conn procmq.Conn) error {
	col, err := conn.ReadUInt32()
	if err != nil {
		return err
	}
	row, err := conn.ReadUInt32()
	if err != nil {
		return err
	}
	value, err := conn.ReadString()
	if err != nil {
		return err
	}
	a := &x.records[row-1]
	switch col {
	case 0:
		if n, err := strconv.Atoi(value); err == nil && n > -1 {
			v := uint64(n)
			x.db.SetOrder(a.ID, v)
			a.Order = v
		}
	case 2:
		if n, err := strconv.Atoi(value); err == nil && n > -1 && n < 256 {
			v := byte(n)
			x.db.SetDocCode(a.ID, v)
			a.DocCode = v
		}
	case 3:
		x.db.SetRouteSheet(a.ID, value)
		a.RouteSheet = value
	case 4:
		if n, err := strconv.Atoi(value); err == nil && n > -1 {
			v := uint64(n)
			x.db.SetProductsCount(a.ID, v)
			a.ProductsCount = v
		}
	}
	return nil
}

func (x *App) onCellRequest(conn procmq.Conn) error {
	col, err := conn.ReadUInt32()
	if err != nil {
		return err
	}
	row, err := conn.ReadUInt32()
	if err != nil {
		return err
	}

	a := x.records[row-1]

	switch col {
	case 0:
		if err := conn.WriteString(intToStr(a.Order)); err != nil {
			return err
		}
	case 2:
		if err := conn.WriteString(intToStr(a.DocCode)); err != nil {
			return err
		}
	case 3:
		if err := conn.WriteString(a.RouteSheet); err != nil {
			return err
		}
	case 4:
		if err := conn.WriteString(intToStr(a.ProductsCount)); err != nil {
			return err
		}
	default:
		log.Fatalf("wrong col %d", col)
	}
	return nil
}

func (x *App) addNewRecord(conn procmq.Conn) error {
	routeSheet, err := conn.ReadString()
	if err != nil {
		return err
	}
	strDocCode, err := conn.ReadString()
	if err != nil {
		return err
	}
	strProdCount, err := conn.ReadString()
	if err != nil {
		return err
	}

	if routeSheet == "" {
		return nil
	}

	docCode, err := strconv.Atoi(strDocCode)
	if err != nil {
		return nil
	}
	prodCount, err := strconv.Atoi(strProdCount)
	if err != nil {
		return nil
	}

	a := data.Entity{
		Time:          time.Now(),
		RouteSheet:    routeSheet,
		Order:         1,
		DocCode:       byte(docCode),
		ProductsCount: uint64(prodCount),
	}
	// все записи за последний год
	xs := x.db.Records(data.NewFilter())
	n := len(xs)
	if n > 0 {
		a.Order = xs[n-1].Order + 1
	}
	x.db.Add(a)
	return nil
}

func (x *App) sendRecords(conn procmq.Conn) error {

	if err := conn.WriteString(intToStr(x.filter.Year)); err != nil {
		return err
	}
	if err := conn.WriteString(intToStr(x.filter.Month)); err != nil {
		return err
	}

	if err := conn.WriteString(x.filter.RouteSheetMask); err != nil {
		return err
	}

	if err := conn.WriteString(intToStr(x.filter.DocCode)); err != nil {
		return err
	}

	if err := conn.WriteString(intToStr(x.filter.OrderFrom)); err != nil {
		return err
	}

	if err := conn.WriteString(intToStr(x.filter.OrderTo)); err != nil {
		return err
	}

	if err := conn.WriteUInt32(uint32(len(x.records))); err != nil {
		return err
	}
	for _, a := range x.records {
		if err := conn.WriteUInt32(uint32(a.Order)); err != nil {
			return err
		}
		if err := conn.WriteString(a.Time.Format("02.01.2006")); err != nil {
			return err
		}
		if err := conn.WriteUInt32(uint32(a.DocCode)); err != nil {
			return err
		}
		if err := conn.WriteString(a.RouteSheet); err != nil {
			return err
		}
		if err := conn.WriteUInt32(uint32(a.ProductsCount)); err != nil {
			return err
		}
	}

	return nil
}
