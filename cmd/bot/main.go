package main

import (
	"inspense-bot/bot"
	"inspense-bot/database"
	"log"
	"os"
)

func main() {
	db, err := initDB(os.Getenv("DB_URL"))
	if err != nil {
		log.Fatal(err)
	}

	boot := bot.BootStrap{
		DB: db,
	}

	b, err := bot.New("bot.yml", boot)
	if err != nil {
		log.Fatal(err)
	}

	b.Start()
}

func initDB(path string) (*database.DB, error) {
	db, err := database.Open(path)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	if err := db.Migrate(); err != nil {
		return nil, err
	}

	return db, nil
}
