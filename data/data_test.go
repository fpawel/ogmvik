package data

import (
	"bytes"
	"io/ioutil"
	"path/filepath"
	"testing"

	"bufio"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"
)

func TestCreateFromTextFile(t *testing.T) {
	db, dir := Open()
	defer db.db.Close()
	println(dir)
	b, err := ioutil.ReadFile(filepath.Join(dir, "data.txt"))
	check(err)

	var xs []Entity

	scanner := bufio.NewScanner(bytes.NewReader(b))

	m := make(map[int]uint64)

	for scanner.Scan() {
		line := scanner.Text()
		strs := strings.Split(line, ",")
		if len(strs) != 4 {
			log.Fatal("len must be 4")
		}

		var a Entity

		docCode, err := strconv.Atoi(strs[0])
		check(err)
		a.DocCode = byte(docCode)

		productsCount, err := strconv.Atoi(strs[1])
		check(err)
		a.ProductsCount = uint64(productsCount)

		a.RouteSheet = strs[2]

		a.Time, err = time.Parse("02.01.2006", strs[3])
		check(err)

		y := a.Time.Year()
		m[y]++
		a.Order = m[y]

		xs = append(xs, a)
		fmt.Println(a)
	}
	db.addRange(xs)
	println(len(xs), "added")
}
