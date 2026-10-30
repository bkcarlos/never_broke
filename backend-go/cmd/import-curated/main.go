// import-curated uses the same validation and transaction as the authenticated
// import API. Run inside the deployed API container; input is read from stdin.
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"neverbroke/internal/curatedimport"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/repository/postgres"
	"os"
)

func main() {
	email := flag.String("user-email", "", "existing owner email")
	commit := flag.Bool("commit", false, "commit after validation (default preview)")
	flag.Parse()
	if *email == "" {
		panic("user-email required")
	}
	db, err := postgres.OpenGorm(dbconfig.LoadFromEnv())
	if err != nil {
		panic("database connection failed")
	}
	var id int64
	if err := db.Raw("SELECT id FROM users WHERE email = ?", *email).Scan(&id).Error; err != nil || id == 0 {
		panic("owner not found")
	}
	var b curatedimport.Bundle
	d := json.NewDecoder(io.LimitReader(os.Stdin, 16<<20))
	if err := d.Decode(&b); err != nil {
		panic(err)
	}
	out, err := curatedimport.Run(context.Background(), db, fmt.Sprint(id), b, *commit)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	json.NewEncoder(os.Stdout).Encode(out)
}
