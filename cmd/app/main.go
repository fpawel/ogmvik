package main

import (
	"fmt"
	"github.com/ansel1/merry"
	"github.com/fpawel/elco/pkg/winapp"
	"github.com/fpawel/goutils/procmq"
	"github.com/fpawel/ogmvik/data"
	"gopkg.in/natefinch/npipe.v2"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"strconv"
	"time"
)

const (
	actionCellEdited uint32 = iota
	actionCellRequest
	actionYear
	actionMonth
	actionRouteSheetMask
	actionDocCode
	actionOrderFrom
	actionOrderTo
	actionActs
	actionAddNewRecord
	actionDelete
)

func main() {

	x := &App{
		filter: data.NewFilter(),
	}
	x.db = data.Open()
	defer x.db.Close()

	// сделать pipeRunner сервер
	pipeReadListener, err := npipe.Listen(`\\.\pipe\$OGMVIKTABLE$`)
	if err != nil {
		panic(err)
	}
	defer pipeReadListener.Close()

	pipeWriteListener, err := npipe.Listen(`\\.\pipe\$OGMVIKTABLE2$`)
	if err != nil {
		panic(err)
	}
	defer pipeWriteListener.Close()



	if err := runUIApp(); err != nil {
		panic(err)
	}

	var pipeRead, pipeWrite procmq.Conn

	pipeRead.Conn, err = pipeReadListener.Accept()
	if err != nil {
		panic(err)
	}

	defer pipeRead.Close()

	pipeWrite.Conn, err = pipeWriteListener.Accept()
	if err != nil {
		panic(err)
	}

	defer pipeWrite.Close()

	donePipeWrite := make(chan bool)
	interruptPipeWrite := make(chan bool, 2)
	sendRecords := make(chan bool)
	go func() {
		defer func() {
			donePipeWrite <- true
		}()
		for {
			select {
			case <-interruptPipeWrite:
				return
			case <-sendRecords:
				if pipeWrite.WriteUInt32(0) != nil || x.sendRecords(pipeWrite) != nil {
					return
				}
			}
		}
	}()

	invalidateRecords := func() {
		x.records = x.db.Records(x.filter)
		sendRecords <- true
	}
	invalidateRecords()

	for {
		cmd, err := pipeRead.ReadUInt32()
		if err != nil {
			break
		}
		switch cmd {
		case actionCellEdited:
			if x.onCellEdited(pipeRead) != nil {
				return
			}

		case actionCellRequest:
			if x.onCellRequest(pipeRead) != nil {
				break
			}

		case actionYear:
			str, err := pipeRead.ReadString()
			if err != nil {
				break
			}
			x.filter.Year, _ = strconv.Atoi(str)
			invalidateRecords()

		case actionMonth:
			str, err := pipeRead.ReadString()
			if err != nil {
				break
			}
			n, _ := strconv.Atoi(str)
			x.filter.Month = time.Month(n)
			invalidateRecords()

		case actionRouteSheetMask:
			x.filter.RouteSheetMask, err = pipeRead.ReadString()
			if err != nil {
				break
			}
			invalidateRecords()

		case actionDocCode:
			s, err := pipeRead.ReadString()
			if err != nil {
				break
			}
			v, _ := strconv.Atoi(s)
			x.filter.DocCode = byte(v)
			invalidateRecords()

		case actionOrderFrom:
			s, err := pipeRead.ReadString()
			if err != nil {
				break
			}
			v, _ := strconv.Atoi(s)
			x.filter.OrderFrom = uint64(v)
			invalidateRecords()

		case actionOrderTo:
			s, err := pipeRead.ReadString()
			if err != nil {
				break
			}
			v, _ := strconv.Atoi(s)
			x.filter.OrderTo = uint64(v)
			invalidateRecords()

		case actionActs:
			x.savePDF(filepath.Join(os.TempDir(), "акты-ВИК.pdf"))

		case actionAddNewRecord:
			if err := x.addNewRecord(pipeRead); err != nil {
				break
			}
			invalidateRecords()
		case actionDelete:
			n, err := pipeRead.ReadUInt32()
			if err != nil {
				break
			}
			m := int(n)
			if m >= 0 && m < len(x.records) {
				x.db.Delete(x.records[m].ID)
				invalidateRecords()
			}

		default:
			log.Fatal("unknown message")
		}
	}
	interruptPipeWrite <- true
	<-donePipeWrite
	println("exited well")
}

func intToStr(x interface{}) string {
	return fmt.Sprintf("%d", x)
}

func runUIApp() error {
	fileName, err := winapp.CurrentDirOrProfileFileName(".ogmvik", "ui.exe")
	if err != nil {
		return merry.Wrap(err)
	}
	return exec.Command(fileName).Start()
}