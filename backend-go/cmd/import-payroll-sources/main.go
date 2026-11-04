package main

import (
	"context"
	"fmt"
	"log"
	"neverbroke/internal/dbconfig"
	"neverbroke/internal/payroll"
	"neverbroke/internal/repository/postgres"
)

func main() {
	db, err := postgres.OpenGorm(dbconfig.LoadFromEnv())
	if err != nil {
		log.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		log.Fatal(err)
	}
	defer sqlDB.Close()
	if err = payroll.SeedPolicyHistory(db); err != nil {
		log.Fatal(err)
	}
	n, err := payroll.ArchiveOfficialSources(context.Background(), db)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("Archived %d official source documents; reviewed policy history imported.\n", n)
}
