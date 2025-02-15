package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"

	"cinemanager-bot/omdb"
	"cinemanager-bot/storage"
)

var (
	// Menu texts
	firstMenu  = "<b>Menu 1</b>\n\nA beautiful menu with a shiny inline button."
	secondMenu = "<b>Menu 2</b>\n\nA better menu with even more shiny inline buttons."

	// Button texts
	nextButton     = "Next"
	backButton     = "Back"
	tutorialButton = "Tutorial"

	// Button ids
	movieYesButton = "movie-yes"
	movieNoButton  = "movie-no"

	// Store bot screaming status
	bot *tgbotapi.BotAPI

	// Keyboard layout for the first menu. One button, one row
	firstMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(nextButton, nextButton),
		),
	)

	// Keyboard layout for the second menu. Two buttons, one per row
	secondMenuMarkup = tgbotapi.NewInlineKeyboardMarkup(
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonData(backButton, backButton),
		),
		tgbotapi.NewInlineKeyboardRow(
			tgbotapi.NewInlineKeyboardButtonURL(tutorialButton, "https://core.telegram.org/bots/api"),
		),
	)
)

func main() {
	var err error

	err = godotenv.Load()
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	// get token from env variable
	token := os.Getenv("TELEGRAM_BOT_TOKEN")

	isDebug := os.Getenv("DEBUG") == "true"
	log.Println("Debug mode:", isDebug)

	log.Println("Starting bot...")
	log.Println(token)
	bot, err = tgbotapi.NewBotAPI(token)
	log.Printf("Authorized on account %s", bot.Self.UserName)
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = isDebug

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	updates := bot.GetUpdatesChan(u)

	go receiveUpdates(ctx, updates)

	log.Println("Start listening for updates. Press Ctrl+C to stop.")

	storage.Init()
	defer storage.DbDispose()

	bufio.NewReader(os.Stdin).ReadBytes('\n')
	cancel()
}

func receiveUpdates(ctx context.Context, updates tgbotapi.UpdatesChannel) {
	for {
		select {
		case update := <-updates:
			log.Println("update received")
			handleUpdate(update)
		case <-ctx.Done():
			log.Println("Stop listening for updates.")
			return
		}
	}
}

func handleUpdate(update tgbotapi.Update) {
	switch {
	// Handle messages
	case update.Message != nil:
		handleMessage(update.Message)

	// Handle button clicks
	case update.CallbackQuery != nil:
		handleButton(update.CallbackQuery)
	}
}

func handleMessage(message *tgbotapi.Message) {
	user := message.From
	text := message.Text

	if user == nil {
		return
	}

	// Print to console
	log.Printf("%s wrote %s", user.FirstName, text)

	var err error
	if strings.HasPrefix(text, "/") {
		err = handleCommand(message.Chat.ID, text)
	} else {
		// This is equivalent to forwarding, without the sender's name
		copyMsg := tgbotapi.NewCopyMessage(message.Chat.ID, message.Chat.ID, message.MessageID)
		_, err = bot.CopyMessage(copyMsg)
	}

	if err != nil {
		log.Printf("An error occured: %s", err.Error())
	}
}

// When we get a command, we react accordingly
func handleCommand(chatId int64, command string) error {
	var err error

	switch {
	case strings.HasPrefix(command, "/menu"):
		err = sendMenu(chatId)
	case strings.HasPrefix(command, "/backlog"):
		msgText := "Current backlog:\n"
		movies := storage.GetMoviesBacklog()
		if len(movies) == 0 {
			msgText += "Empty"
		} else {
			for _, movie := range movies {
				msgText += movie.Title + "\n"
			}
		}

		_, err = bot.Send(tgbotapi.NewMessage(chatId, msgText))
	case strings.HasPrefix(command, "/movie"):
		title := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(command, "/movie"), "@"+bot.Self.UserName))
		if title == "" {
			_, err = bot.Send(tgbotapi.NewMessage(chatId, "Please provide a movie title"))
		} else {
			movieData, fetchErr := omdb.FetchMovieData(title)

			if fetchErr != nil {
				_, err = bot.Send(tgbotapi.NewMessage(chatId, fetchErr.Error()))
			} else {

				movieMenuMarkup := tgbotapi.NewInlineKeyboardMarkup(
					tgbotapi.NewInlineKeyboardRow(
						tgbotapi.NewInlineKeyboardButtonData("Yes, post in chat", movieYesButton),
						tgbotapi.NewInlineKeyboardButtonData("No, ditch it", movieNoButton),
					),
				)
				cardText := formatMovieData(&movieData)
				log.Println(cardText)
				movieCard := tgbotapi.NewMessage(chatId, cardText)
				movieCard.ReplyMarkup = movieMenuMarkup
				movieCard.ParseMode = tgbotapi.ModeHTML
				msg, sendErr := bot.Send(movieCard)
				if sendErr != nil {
					log.Printf("Error sending message: %s", sendErr.Error())
				} else {
					storage.SaveMovie(&movieData, msg.MessageID)
				}

			}
		}

	}

	return err
}

func handleButton(query *tgbotapi.CallbackQuery) {
	var text string

	markup := tgbotapi.NewInlineKeyboardMarkup()
	message := query.Message
	if query.Data == movieNoButton {
		msg := tgbotapi.NewDeleteMessage(message.Chat.ID, message.MessageID)
		storage.DeleteMovieByMessageId(message.MessageID)
		bot.Send(msg)
		return
	}

	if query.Data == movieYesButton {
		movieData, err := storage.GetMovieByMessageId(message.MessageID)
		if err != nil {
			log.Printf("Error getting movie by message ID: %s", err.Error())
			return
		}
		msgText := formatMovieData(movieData)
		msg := tgbotapi.NewEditMessageText(message.Chat.ID, message.MessageID, msgText)

		msg.ParseMode = tgbotapi.ModeHTML
		bot.Send(msg)
		return
	}

	if query.Data == nextButton {
		text = secondMenu
		markup = secondMenuMarkup
	} else if query.Data == backButton {
		text = firstMenu
		markup = firstMenuMarkup
	}

	callbackCfg := tgbotapi.NewCallback(query.ID, "")
	bot.Send(callbackCfg)

	// Replace menu text and keyboard
	msg := tgbotapi.NewEditMessageTextAndMarkup(message.Chat.ID, message.MessageID, text, markup)
	msg.ParseMode = tgbotapi.ModeHTML
	bot.Send(msg)
}

func sendMenu(chatId int64) error {
	msg := tgbotapi.NewMessage(chatId, firstMenu)
	msg.ParseMode = tgbotapi.ModeHTML
	msg.ReplyMarkup = firstMenuMarkup
	_, err := bot.Send(msg)
	return err
}

func formatMovieData(movie *omdb.Movie) string {
	return fmt.Sprintf(
		"<b>Title:</b> %s\n<b>Year:</b> %s\n<b>Plot:</b> %s\n<a href=\"%s\">Poster</a>",
		movie.Title, movie.Year, movie.Plot, movie.Poster,
	)
}
