package iploc

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"strconv"
)

func (t *table) createIPTable() error {
	// CREATE DATABASE ip2location WITH ENCODING 'UTF8';
	query := `
	CREATE TABLE IF NOT EXISTS ip2location_db1(
		ip_from bigint NOT NULL,
		ip_to bigint NOT NULL,
		country_code character(2) NOT NULL,
		country_name character varying(64) NOT NULL,
		CONSTRAINT ip2location_db1_pkey PRIMARY KEY (ip_from, ip_to)
	);
	`
	_, err := t.db.ExecContext(context.TODO(), query)
	return err
}

func (t *table) importFromFile() error {
	// import table from csv
	// 'IP2LOCATION-LITE-DB1.CSV'

	wd, _ := os.Getwd()
	pathToFile := path.Join(wd, "file", "ip2loc.csv")
	if pathToFile == "" {
		log.Fatal(fmt.Errorf("path error %v", wd))
	}

	f, _ := os.Open(pathToFile)
	defer f.Close()
	r := csv.NewReader(f)

	// prepared statement
	query := `
	INSERT INTO ip2location_db1(ip_from,ip_to,country_code,country_name)
	VALUES($1,$2,$3,$4)
	`
	stmt, err := t.db.PreparexContext(context.TODO(), query)
	if err != nil {
		log.Fatal(err)
	}
	defer stmt.Close()

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
			_, err := stmt.ExecContext(context.TODO(), from, to, record[2], record[3])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// CREATE INDEX

func (t *table) drop() error {
	query := `
		DROP TABLE IF EXISTS ip2location_db1
	`
	_, err := t.db.ExecContext(context.TODO(), query)
	return err
}
