package main

import (
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// UserState holds individual user information and progress
type UserState struct {
	AwaitingUserName  bool
	AwaitingPhone     bool
	AwaitingPanelName bool
	AwaitingPayment   bool
	AwaitingReceipt   bool
	SelectedService   string
	UserName          string
	PhoneNumber       string
	PanelName         string
}

// Map to store user states by chat ID
var userStates = make(map[int64]*UserState)

// Admin User ID (who will receive the purchase details)
const AdminUserID int64 = 94152088

func main() {
	bot, err := tgbotapi.NewBotAPI("7860223140:AAGQrOGg6hJaBkQabESr5RJIvJTzgajJbJk")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = false
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			handleUpdate(update.Message, bot)
		}
	}
}

func handleUpdate(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	chatID := message.Chat.ID
	state, exists := userStates[chatID]
	if !exists {
		// Initialize new state if it doesn't exist
		state = &UserState{}
		userStates[chatID] = state
	}

	// Handle different stages of interaction
	if state.AwaitingUserName {
		state.UserName = message.Text
		askForPhone(message, bot)
		state.AwaitingUserName = false
		state.AwaitingPhone = true
		return
	}

	if state.AwaitingPhone {
		state.PhoneNumber = message.Text
		askForReceipt(message, bot)
		state.AwaitingPhone = false
		state.AwaitingReceipt = true
		return
	}

	if state.AwaitingReceipt && message.Photo != nil {
		sendPurchaseDetailsToAdmin(message, bot, state)
		confirmRequest(message, bot)
		showMainMenu(message, bot)
		delete(userStates, chatID) // Clean up state after completion
		return
	}

	// Handle command or options selection
	if message.IsCommand() {
		switch message.Command() {
		case "start":
			showMainMenu(message, bot)
		default:
			msg := tgbotapi.NewMessage(message.Chat.ID, "Unknown command")
			bot.Send(msg)
		}
	} else {
		handleSelection(message, bot, state)
	}
}

func handleSelection(message *tgbotapi.Message, bot *tgbotapi.BotAPI, state *UserState) {
	switch message.Text {
	case "خرید سرویس":
		showServiceOptions(message, bot)
	case "تمدید سرویس":
		startRenewalProcess(message, bot, state)
	case "تک کاربره":
		showSingleUserOptions(message, bot)
	case "دو کاربره":
		showTwoUserOptions(message, bot)
	case "نامحدود":
		showUnlimitedOptions(message, bot)
	case "Back":
		showServiceOptions(message, bot)
	case "۴۰ گیگ ۱ ماهه:۷۵ تومن", "۶۰ گیگ ۱ ماهه:۹۰ تومن", "۷۵ گیگ ۱ ماهه:۱۰۰ تومن", "۱۰۰گیگ ۱ ماهه:۱۲۰تومن",
		"۷۰گیگ ۱ ماهه ۱۲۰ تومن", "۹۰ گیگ ۱ ماهه ۱۴۰ تومن", "۱۲۰گیگ ۱ ماهه ۱۶۰ تومن", "۲۰۰ گیگ ۱ ماهه ۲۲۰ تومن",
		"۱ ماهه ۱۵۰ گیگ ۲۵۰ تومن", "۱ماهه ۲۵۰ گیگ ۳۱۵ تومن", "۱ماهه ۳۵۰ گیگ  ۴۰۰ تومن":
		state.SelectedService = message.Text
		// Only ask for the name here, no need to ask again in askForName
		state.AwaitingUserName = true
		msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً نام خود را وارد کنید:")
		msg.ReplyMarkup = tgbotapi.NewRemoveKeyboard(true)
		bot.Send(msg)

	default:
		// Handle receipt or other messages
		if len(message.Photo) > 0 && state.AwaitingReceipt {
			sendPurchaseDetailsToAdmin(message, bot, state)
			confirmRequest(message, bot)
			showMainMenu(message, bot)
			delete(userStates, message.Chat.ID) // Clean up state after completion
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "گزینه نامعتبر است. لطفاً دوباره امتحان کنید.")
			bot.Send(msg)
		}
	}
}

func startRenewalProcess(message *tgbotapi.Message, bot *tgbotapi.BotAPI, state *UserState) {
	state.SelectedService = "تمدید سرویس"
	msg := tgbotapi.NewMessage(message.Chat.ID, "برای تمدید سرویس لطفاً نام خود را وارد کنید:")
	bot.Send(msg)
	state.AwaitingUserName = true
}

func sendPurchaseDetailsToAdmin(message *tgbotapi.Message, bot *tgbotapi.BotAPI, state *UserState) {
	// Prepare the message to send to the admin
	adminMessage := tgbotapi.NewMessage(AdminUserID, "درخواست خرید:\n" + state.SelectedService + "\nنام: " + state.UserName + "\nشماره تلفن: " + state.PhoneNumber + "\nنام پنل: " + state.PanelName)
	bot.Send(adminMessage)

	// Send the photo (receipt) to the admin as well
	if len(message.Photo) > 0 {
		photo := message.Photo[len(message.Photo)-1] // Get the largest version of the photo
		photoConfig := tgbotapi.NewPhoto(AdminUserID, tgbotapi.FileID(photo.FileID))
		bot.Send(photoConfig)
	}
}

func showMainMenu(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً یک گزینه را انتخاب کنید:")
	mainKeyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("خرید سرویس"),
			tgbotapi.NewKeyboardButton("تمدید سرویس"),
		),
	)
	msg.ReplyMarkup = mainKeyboard
	bot.Send(msg)
}

func showServiceOptions(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً نوع سرویس را انتخاب کنید:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("تک کاربره"),
			tgbotapi.NewKeyboardButton("دو کاربره"),
			tgbotapi.NewKeyboardButton("نامحدود"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showSingleUserOptions(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "یک گزینه را انتخاب کنید:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۴۰ گیگ ۱ ماهه:۷۵ تومن"),
			tgbotapi.NewKeyboardButton("۶۰ گیگ ۱ ماهه:۹۰ تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۷۵ گیگ ۱ ماهه:۱۰۰ تومن"),
			tgbotapi.NewKeyboardButton("۱۰۰گیگ ۱ ماهه:۱۲۰تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showTwoUserOptions(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "یک گزینه را انتخاب کنید:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۷۰گیگ ۱ ماهه ۱۲۰ تومن"),
			tgbotapi.NewKeyboardButton("۹۰ گیگ ۱ ماهه ۱۴۰ تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۱۲۰گیگ ۱ ماهه ۱۶۰ تومن"),
			tgbotapi.NewKeyboardButton("۲۰۰ گیگ ۱ ماهه ۲۲۰ تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func showUnlimitedOptions(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "یک گزینه را انتخاب کنید:")
	keyboard := tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۱ ماهه ۱۵۰ گیگ ۲۵۰ تومن"),
			tgbotapi.NewKeyboardButton("۱ماهه ۲۵۰ گیگ ۳۱۵ تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("۱ماهه ۳۵۰ گیگ  ۴۰۰ تومن"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("Back"),
		),
	)
	msg.ReplyMarkup = keyboard
	bot.Send(msg)
}

func askForPhone(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً شماره تلفن خود را وارد کنید:")
	bot.Send(msg)
}

func askForReceipt(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً رسید پرداخت را ارسال کنید:")
	bot.Send(msg)
}

func confirmRequest(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "درخواست شما با موفقیت ثبت شد. پشتیبانی به زودی با شما تماس خواهد گرفت.")
	bot.Send(msg)
}
