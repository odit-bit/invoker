package iploc

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path"
	"strconv"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
)

type ipdata struct {
	From    int64
	to      int64
	code    string
	country string
}

func insertFixture(db *sqlx.DB, data *ipdata) {
	query := `
		INSERT INTO ip2location_db1(ip_from,ip_to,country_code,country_name)
		VALUES($1,$2,$3,$4)
	`
	_, err := db.ExecContext(context.TODO(), query, data.From, data.to, data.code, data.country)
	if err != nil {
		log.Fatal(err)
	}
}

func insertFromFile(db *sqlx.DB) error {
	// import data from file
	wd, _ := os.Getwd()
	pathToFile := path.Join(wd, "file", "ip2loc.csv")
	if pathToFile == "" {
		log.Fatal(fmt.Errorf("path error %v", wd))
	}

	f, _ := os.Open(pathToFile)
	defer f.Close()
	r := csv.NewReader(f)

	for {
		record, err := r.Read()
		if err != nil {
			if err == io.EOF {
				break
			}
			return fmt.Errorf("read csv: %v", err)
		}
		if record[3] == "Indonesia" {
			from, _ := strconv.Atoi(record[0])
			to, _ := strconv.Atoi(record[1])

			insertFixture(db, &ipdata{
				From:    int64(from),
				to:      int64(to),
				code:    record[2],
				country: record[3],
			})
		}
	}
	return nil
}

func Test_iptable(t *testing.T) {
	dsn := "host=localhost user=development password=credential dbname=development sslmode=disable"
	db, err := sqlx.Connect("pgx", dsn)
	if err != nil {
		t.Fatal(err)
	}
	err = db.Ping()
	if err != nil {
		t.Fatal(err)
	}

	ipt, err := New(db)
	if err != nil {
		t.Fatal(err)
	}

	//FIXTURE DATA
	// insertFixture(db, &ipdata{From: 1735374848, to: 1735383039, code: "ID", country: "Indonesia"})
	// if err := insertFromFile(db); err != nil {
	// 	t.Fatal(err)
	// }

	ip, err := net.ResolveIPAddr("ip", "103.111.184.0")
	if err != nil {
		t.Fatal(err)
	}
	country, err := ipt.LookupCountry(ip)
	if err != nil {
		t.Fatal(err)
	}

	if country != "Indonesia" {
		t.Fatalf("expected indonesia, got: %v", country)
	}

	ipt.drop()
}
