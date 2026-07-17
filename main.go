package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

		"github.com/joho/godotenv"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type Transaction struct {
	ChatID int64     `json:"chat_id"`
	Type   string    `json:"type"`
	Amount int       `json:"amount"`
	Note   string    `json:"note"`
	Time   time.Time `json:"time"`
}

const dataFile = "data.json"

func loadTransactions() []Transaction {
	var transactions []Transaction

	file, err := os.ReadFile(dataFile)
	if err != nil {
		return transactions
	}

	json.Unmarshal(file, &transactions)
	return transactions
}

func saveTransactions(transactions []Transaction) {
	data, _ := json.MarshalIndent(transactions, "", "  ")
	os.WriteFile(dataFile, data, 0644)
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Warning: .env not pound")
	}

	token := os.Getenv("BOT_TOKEN")
	if token == "" {
		log.Panic("BOT_TOKEN kosong! file .env not pound")
	}

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

		newTx := Transaction{
			ChatID: chatID,
			Type:   parts[0],
			Amount: amount,
			Note:   note,
			Time:   time.Now(),
		}

		transactions := loadTransactions()
		transactions = append(transactions, newTx)
		saveTransactions(transactions)

		reply := tgbotapi.NewMessage(chatID, fmt.Sprintf("Tercatat: %s Rp%d (%s)", newTx.Type, newTx.Amount, newTx.Note))
		bot.Send(reply)
	}
}