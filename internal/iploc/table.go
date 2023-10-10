package iploc

import (
	"context"
	"fmt"
	"net"

	"github.com/jmoiron/sqlx"
)

type table struct {
	db *sqlx.DB
}

func New(db *sqlx.DB) (*table, error) {
	t := table{
		db: db,
	}

	err := t.migrate()
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (t *table) migrate() error {
	// create table
	if err := t.createIPTable(); err != nil {
		return err
	}

	if err := t.importFromFile(); err != nil {
		return err
	}

	return nil

}

func (t *table) LookupCountry(ipAddr *net.IPAddr) (string, error) {
	// // should it re-validate ip is not private, not loopback, not local
	// if ok := isValidate(&ipAddr.IP); !ok {
	// 	return "", fmt.Errorf("invalid ip: %v", ipAddr.IP.String())
	// }

	// convert ip to number
	ipNum := ipToInt(ipAddr)

	// query the ip number
	query := `
	SELECT country_name
	FROM ip2location_db1
	WHERE ip_from <= $1 AND ip_to >= $1;
`
	var country string
	err := t.db.QueryRowxContext(context.TODO(), query, ipNum).Scan(&country)
	if err != nil {
		return "", fmt.Errorf("ip-table lookup: %v", err)
	}

	return country, nil
}

func isValidate(ip *net.IP) bool {
	return !ip.IsLoopback() && !ip.IsPrivate() && !ip.IsUnspecified()
}

func ipToInt(ipAddr *net.IPAddr) uint64 {
	ipBytes := ipAddr.IP.To4()
	ipNumber := uint64(ipBytes[0])<<24 | uint64(ipBytes[1])<<16 | uint64(ipBytes[2])<<8 | uint64(ipBytes[3])
	return ipNumber
}
