package bot

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	apiKey    string
	apiSecret string
	botToken  string
)

func init() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Ошибка при загрузке переменных окружения из файла .env")
	}

	apiKey = os.Getenv("API_KEY")
	apiSecret = os.Getenv("API_SECRET")
	botToken = os.Getenv("BOT_TOKEN")

}

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/help"),
		tgbotapi.NewKeyboardButton("/sayhi"),
		tgbotapi.NewKeyboardButton("/status"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/ping"),
		tgbotapi.NewKeyboardButton("/time"),
		tgbotapi.NewKeyboardButton("6"),
	),
)

func RunBot() {
	// Logic bot

	bot, err := tgbotapi.NewBotAPI(botToken)
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
			msg.Text = "Бот поддерживает команды: /sayhi , /status , /ping , /time "
		case "sayhi":
			msg.Text = "Добрый день, это бот для автоматической работы с биржей."
		case "status":
			msg.Text = "В работе."
		case "ping":
			statusCode, responseBody := sendRequestToAPI("https://api.mexc.com/api/v3/ping")
			msg.Text = responseBody
			if statusCode != 200 {
				msg.Text = "Error"
			}
		case "time":
			statusCode, responseBody := sendRequestToAPI("https://api.mexc.com/api/v3/time")
			msg.Text = responseBody
			if statusCode != 200 {
				msg.Text = "Error"
			}
		case "6":
			msg.Text = "заглушка"
		default:
			msg.Text = "I don't know that command"
		}

		// Okay, we're sending our message off! We don't care about the message
		// we just sent, so we'll discard it.
		if _, err := bot.Send(msg); err != nil {
			// Note that panics are a bad way to handle errors. Telegram can
			// have service outages or network errors, you should retry sending
			// messages or more gracefully handle failures.
			panic(err)
		}
	}
}

func sendRequestToAPI(url string) (int, string) {
	// Создаем запрос на API биржи
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Println(err)
		return 0, ""
	}

	// Отправляем запрос на API биржи
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Println(err)
		return 0, ""
	}

	// Обрабатываем ответ от API биржи
	defer resp.Body.Close()

	// Читаем содержимое ответа
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return 0, ""
	}

	// Возвращаем код ответа и содержимое ответа
	return resp.StatusCode, string(body)
}

// Inline Keyboard----------------------------------------------------------------------------------
// var numericKeyboard = tgbotapi.NewInlineKeyboardMarkup(
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
// 		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
// 		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
// 	),
// 	tgbotapi.NewInlineKeyboardRow(
// 		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
// 		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
// 		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
// 	),
// )

// func RunBot() {
// 	bot, err := tgbotapi.NewBotAPI(botToken)
// 	if err != nil {
// 		log.Panic(err)
// 	}

// 	bot.Debug = true

// 	log.Printf("Authorized on account %s", bot.Self.UserName)

// 	u := tgbotapi.NewUpdate(0)
// 	u.Timeout = 60

// 	updates := bot.GetUpdatesChan(u)

// 	// Loop through each update.
// 	for update := range updates {
// 		// Check if we've gotten a message update.
// 		if update.Message != nil {
// 			// Construct a new message from the given chat ID and containing
// 			// the text that we received.
// 			msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

// 			// If the message was open, add a copy of our numeric keyboard.
// 			switch update.Message.Text {
// 			case "open":
// 				msg.ReplyMarkup = numericKeyboard

// 			}

// 			// Send the message.
// 			if _, err = bot.Send(msg); err != nil {
// 				panic(err)
// 			}
// 		} else if update.CallbackQuery != nil {
// 			// Respond to the callback query, telling Telegram to show the user
// 			// a message with the data received.
// 			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
// 			if _, err := bot.Request(callback); err != nil {
// 				panic(err)
// 			}

// 			// And finally, send a message containing the data received.
// 			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
// 			if _, err := bot.Send(msg); err != nil {
// 				panic(err)
// 			}
// 		}
// 	}
// }
