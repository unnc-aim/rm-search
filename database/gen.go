package main

import (
	"gorm.io/driver/postgres"
	"gorm.io/gen"
	"gorm.io/gorm"
)

const DSN = "host=localhost port=5432 user=postgres password=123456 dbname=rm_search sslmode=disable"

func main() {
	g := gen.NewGenerator(gen.Config{
		OutPath:           "database/query",
		Mode:              gen.WithoutContext | gen.WithDefaultQuery | gen.WithQueryInterface, // generate mode
		FieldWithTypeTag:  true,
		FieldWithIndexTag: true,
	})

	db, _ := gorm.Open(postgres.Open(DSN))
	g.UseDB(db) // reuse your gorm db

	g.ApplyBasic(
		// Generate structs from all tables of current database
		g.GenerateAllTable()...,
	)
	// Generate the code
	g.Execute()
}
