package bot

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"

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

var keyboard = tgbotapi.NewInlineKeyboardMarkup(
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonURL("1.com", "http://1.com"),
		tgbotapi.NewInlineKeyboardButtonData("2", "2"),
		tgbotapi.NewInlineKeyboardButtonData("3", "3"),
	),
	tgbotapi.NewInlineKeyboardRow(
		tgbotapi.NewInlineKeyboardButtonData("4", "4"),
		tgbotapi.NewInlineKeyboardButtonData("5", "5"),
		tgbotapi.NewInlineKeyboardButtonData("6", "6"),
	),
)

var numericKeyboard = tgbotapi.NewReplyKeyboard(
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/help"),
		tgbotapi.NewKeyboardButton("/sayhi"),
		tgbotapi.NewKeyboardButton("/status"),
	),
	tgbotapi.NewKeyboardButtonRow(
		tgbotapi.NewKeyboardButton("/ping"),
		tgbotapi.NewKeyboardButton("/time"),
		tgbotapi.NewKeyboardButton("/price"),
	),
)

func RunBot() {
	// Logic bot

	bot, err := tgbotapi.NewBotAPI(botToken)
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	// Создаем новую структуру UpdateConfig со смещением 0. Смещения используются
	// чтобы убедиться, что Telegram знает, что мы обработали предыдущие значения, и нам
	// не нужно их повторять.
	updateConfig := tgbotapi.NewUpdate(0)

	// Сообщите Telegram, что нам следует ждать до 30 секунд при каждом запросе на
	// обновление. Таким образом, мы можем получать информацию так же быстро, как если бы делали много
	// частых запросов, не отправляя почти столько же.
	updateConfig.Timeout = 30

	// Начните опрос Telegram на предмет обновлений
	updates := bot.GetUpdatesChan(updateConfig)

	// Давайте рассмотрим каждое обновление, которое мы получаем от Telegram.
	for update := range updates {
		// Telegram может отправлять множество типов обновлений в зависимости от того, что делает ваш бот
		// Сейчас мы хотим просматривать только сообщения, поэтому можем
		// отбросить любые другие обновления.
		if update.Message == nil {
			continue
		} else if update.CallbackQuery != nil {
			// Ответ на запрос обратного вызова, сообщающий Telegram о необходимости показать пользователю
			// сообщение с полученными данными.
			callback := tgbotapi.NewCallback(update.CallbackQuery.ID, update.CallbackQuery.Data)
			if _, err := bot.Request(callback); err != nil {
				panic(err)
			}

			// отправьте сообщение, содержащее полученные данные.
			msg := tgbotapi.NewMessage(update.CallbackQuery.Message.Chat.ID, update.CallbackQuery.Data)
			if _, err := bot.Send(msg); err != nil {
				panic(err)
			}
		}

		// if !update.Message.IsCommand() { // игнорировать любые некомандные сообщения
		//     continue
		// }

		// Теперь, когда мы знаем, что получили новое сообщение, мы можем создать
		// ответ! Мы возьмем идентификатор чата и текст из входящего сообщения
		// и используем его для создания нового сообщения.
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)

		// Мы также скажем, что это сообщение является ответом на предыдущее сообщение.
		// Для любых других спецификаций, кроме идентификатора чата или текста, вам необходимо
		// задать поля в `MessageConfig`.
		msg.ReplyToMessageID = update.Message.MessageID

		switch update.Message.Text {
		case "open":
			msg.ReplyMarkup = numericKeyboard
		case "close":
			msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		}

		// Извлеките команду из сообщения.
		switch update.Message.Command() {
		case "help":
			msg.Text = "Бот поддерживает команды: /sayhi , /status , /ping , /time "
		case "sayhi":
			msg.Text = "Добрый день, это бот для автоматической работы с биржей."
		case "status":
			msg.ReplyMarkup = keyboard
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
		case "price":
			statusCode, responseBody := sendRequestToAPI("https://api.mexc.com/api/v3/avgPrice?symbol=KASUSDT")
			msg.Text = responseBody
			if statusCode != 200 {
				msg.Text = "Error"
			}
		default:
			msg.Text = "I don't know that command"
		}

		// Хорошо, мы отправляем наше сообщение! Нам не важно сообщение, которое
		// мы только что отправили, поэтому мы его отбросим.
		if _, err := bot.Send(msg); err != nil {
			// Обратите внимание, что паника — плохой способ обработки ошибок. У Telegram могут
			// возникнуть сбои в работе сервиса или сетевые ошибки, вам следует повторить отправку
			// сообщений или более изящно обработать сбои.
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

func sendRequestToAPItoTime(url string) {
	for {
		// Создаем запрос на API биржи
		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			log.Println(err)
			continue
		}

		// Отправляем запрос на API биржи
		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			log.Println(err)
			continue
		}

		// Обрабатываем ответ от API биржи
		defer resp.Body.Close()

		// Читаем содержимое ответа
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Println(err)
			continue
		}

		// Возвращаем код ответа и содержимое ответа
		log.Println(resp.StatusCode, string(body))

		// Приостанавливаем выполнение на 1 секунду
		time.Sleep(1 * time.Second)
	}
}

// Inline Keyboard----------------------------------------------------------------------------------

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
// 				msg.ReplyMarkup = keyboard

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
