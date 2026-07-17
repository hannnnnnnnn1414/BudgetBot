package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	_ "github.com/lib/pq"
)

var db *sql.DB

func connectDB() *sql.DB {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	database, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Panic(err)
	}

	err = database.Ping()
	if err != nil {
		log.Panic(err)
	}

	_, err = database.Exec(`
		CREATE TABLE IF NOT EXISTS transactions (
			id SERIAL PRIMARY KEY,
			chat_id BIGINT NOT NULL,
			type TEXT NOT NULL,
			amount INT NOT NULL,
			note TEXT,
			created_at TIMESTAMP DEFAULT NOW()
		)
	`)
	if err != nil {
		log.Panic(err)
	}

	return database
}

func insertTransaction(chatID int64, txType string, amount int, note string) error {
	_, err := db.Exec(
		"INSERT INTO transactions (chat_id, type, amount, note) VALUES ($1, $2, $3, $4)",
		chatID, txType, amount, note,
	)
	return err
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env gak ketemu")
	}

	db = connectDB()
	fmt.Println("Berhasil konek ke database!")

	token := os.Getenv("BOT_TOKEN")
	bot, err := tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Panic(err)
	}

	fmt.Println("Bot jalan sebagai:", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message == nil {
			continue
		}

		text := update.Message.Text
		chatID := update.Message.Chat.ID

		parts := strings.SplitN(text, " ", 3)

		if len(parts) < 2 || (parts[0] != "keluar" && parts[0] != "masuk") {
			reply := tgbotapi.NewMessage(chatID, "Format salah bro. Contoh: keluar 20000 makan siang")
			bot.Send(reply)
			continue
		}

		amount, err := strconv.Atoi(parts[1])
		if err != nil {
			reply := tgbotapi.NewMessage(chatID, "Jumlah uangnya harus angka ya. Contoh: keluar 20000 makan siang")
			bot.Send(reply)
			continue
		}

		note := ""
		if len(parts) == 3 {
			note = parts[2]
		}

		err = insertTransaction(chatID, parts[0], amount, note)
		if err != nil {
			reply := tgbotapi.NewMessage(chatID, "Gagal nyimpen ke database :(")
			bot.Send(reply)
			continue
		}

		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Tercatat: %s Rp%d (%s)", parts[0], amount, note))
		bot.Send(reply)
	}
}