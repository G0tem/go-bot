package bot

import (
	"log"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

// const (
// 	apiKey = "Ваш API ключ"
// 	apiURL = "Адрес запроса на API биржи"
// )

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("1"),
		tgbotapi.NewKeyboardButton("2"),
		tgbotapi.NewKeyboardButton("3"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("4"),
		tgbotapi.NewKeyboardButton("5"),
		tgbotapi.NewKeyboardButton("6"),
	),
)

func RunBot() {
	// Logic bot
	// Loading .env
	err := godotenv.Load()

	if err != nil {
		log.Fatal("Ошибка при загрузке переменных окружения из файла .env")
	}

	bot, err := tgbotapi.NewBotAPI(os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// Create a new UpdateConfig struct with an offset of 0. Offsets are used
	// to make sure Telegram knows we've handled previous values and we don't
	// need them repeated.
	updateConfig := tgbotapi.NewUpdate(0)

	// Tell Telegram we should wait up to 30 seconds on each request for an
	// update. This way we can get information just as quickly as making many
	// frequent requests without having to send nearly as many.
	updateConfig.Timeout = 30

	// Start polling Telegram for updates.
	updates := bot.GetUpdatesChan(updateConfig)

	// Let's go through each update that we're getting from Telegram.
	for update := range updates {
		// Telegram can send many types of updates depending on what your Bot
		// is up to. We only want to look at messages for now, so we can
		// discard any other updates.
		if update.Message == nil {
			continue
		}

		// if !update.Message.IsCommand() { // ignore any non-command Messages
		//     continue
		// }

		// Now that we know we've gotten a new message, we can construct a
		// reply! We'll take the Chat ID and Text from the incoming message
		// and use it to create a new message.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		// We'll also say that this message is a reply to the previous message.
		// For any other specifications than Chat ID or Text, you'll need to
		// set fields on the `MessageConfig`.
		msg.ReplyToMessageID = update.Message.MessageID

		switch update.Message.Text {
		case "open":
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}

		// Extract the command from the Message.
		switch update.Message.Command() {
		case "help":
			msg.Text = "I understand /sayhi and /status."
		case "sayhi":
			msg.Text = "Hi :)"
		case "status":
			msg.Text = "I'm ok."
		default:
			msg.Text = "I don't know that command"
		}

		// // Okay, we're sending our message off! We don't care about the message
		// // we just sent, so we'll discard it.
		// if _, err := bot.Send(msg); err != nil {
		// 	// Note that panics are a bad way to handle errors. Telegram can
		// 	// have service outages or network errors, you should retry sending
		// 	// messages or more gracefully handle failures.
		// 	panic(err)
		// }
	}
}

// func sendRequestToAPI(chatID int64) {
// 	// Создаем запрос на API биржи
// 	req, err := http.NewRequest("GET", apiURL, nil)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	// Добавляем API ключ в запрос
// 	req.Header.Set("Authorization", "Bearer "+apiKey)

// 	// Отправляем запрос на API биржи
// 	resp, err := http.DefaultClient.Do(req)
// 	if err != nil {
// 		log.Println(err)
// 		return
// 	}

// 	// Обрабатываем ответ от API биржи
// 	defer resp.Body.Close()
// 	// ...
// }
