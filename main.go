package main

import (
	"log"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

var currentCommand string
var awaitingUserName, awaitingPhone, awaitingPanelName, awaitingPayment, awaitingReceipt bool
var selectedService, panelName, userName, phoneNumber string

// Admin User ID (who will receive the purchase details)
const AdminUserID int64 = 94152088

func main() {
	bot, err := tgbotapi.NewBotAPI("7860223140:AAGQrOGg6hJaBkQabESr5RJIvJTzgajJbJk")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	updateConfig := tgbotapi.NewUpdate(0)
	updateConfig.Timeout = 60

	updates := bot.GetUpdatesChan(updateConfig)

	for update := range updates {
		if update.Message != nil {
			// Handle different states (awaiting user input for name, phone, panel, etc.)
			if awaitingUserName {
				userName = update.Message.Text
				askForPhone(update.Message, bot)
				awaitingUserName = false
				awaitingPhone = true
				continue
			}

			if awaitingPhone {
				phoneNumber = update.Message.Text
				if currentCommand == "تمدید سرویس" {
					askForReceipt(update.Message, bot)
					awaitingPhone = false
					awaitingReceipt = true
				} else {
					askForReceipt(update.Message, bot)
					awaitingPhone = false
					awaitingReceipt = true
				}
				continue
			}

			if awaitingPanelName {
				panelName = update.Message.Text
				askForReceipt(update.Message, bot)
				awaitingPanelName = false
				awaitingReceipt = true
				continue
			}

			if awaitingReceipt && update.Message.Photo != nil {
				// Send purchase details to the admin
				sendPurchaseDetailsToAdmin(update.Message, bot)
				// Confirm the request to the user
				confirmRequest(update.Message, bot)
				// Return user to main menu
				showMainMenu(update.Message, bot)
				continue
			}

			// Handle command or options selection
			if update.Message.IsCommand() {
				switch update.Message.Command() {
				case "start":
					showMainMenu(update.Message, bot)
				default:
					msg := tgbotapi.NewMessage(update.Message.Chat.ID, "Unknown command")
					bot.Send(msg)
				}
			} else {
				handleSelection(update.Message, bot)
			}
		}
	}
}


func sendPurchaseDetailsToAdmin(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	// Prepare the message to send to the admin
	adminMessage := tgbotapi.NewMessage(AdminUserID, "درخواست خرید:\n" + selectedService + "\nنام: " + userName + "\nشماره تلفن: " + phoneNumber + "\nنام پنل: " + panelName)
	// Send the message to the admin
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

func handleSelection(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	switch message.Text {
	case "خرید سرویس":
		showServiceOptions(message, bot)
	case "تمدید سرویس":
		startRenewalProcess(message, bot)
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
		selectedService = message.Text
		askForName(message, bot)
	default:
		// Handle non-text inputs such as images (photos)
		if len(message.Photo) > 0 && awaitingReceipt {
			// Handle the image upload and receipt
			sendPurchaseDetailsToAdmin(message, bot)
			confirmRequest(message, bot)
			showMainMenu(message, bot)
		} else {
			msg := tgbotapi.NewMessage(message.Chat.ID, "گزینه نامعتبر است. لطفاً دوباره امتحان کنید.")
			bot.Send(msg)
		}
	}
}


func startRenewalProcess(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	currentCommand = "تمدید سرویس"
	msg := tgbotapi.NewMessage(message.Chat.ID, "برای تمدید سرویس لطفاً نام خود را وارد کنید:")
	bot.Send(msg)
	awaitingUserName = true // Set awaitingUserName to true
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

func askForName(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً نام خود را وارد کنید:")
	bot.Send(msg)
	awaitingUserName = true
}

func askForPhone(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفا شماره تلفن یا آیدی تلگرام خود را وارد کنید (ترجیحا آیدی):")
	bot.Send(msg)
}

func askForPanelName(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "اگر از قبل سرویسی از ما خریداری کرده اید، نام پنل آن سرویس را وارد کنید.\nتوجه داشته باشید که در غیر این صورت پنل جدیدی برای شما ساخته میشود!")
	bot.Send(msg)
}

func askForReceipt(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "لطفاً عکس رسید خرید خود را ارسال کنید:")
	bot.Send(msg)
}

func confirmRequest(message *tgbotapi.Message, bot *tgbotapi.BotAPI) {
	msg := tgbotapi.NewMessage(message.Chat.ID, "درخواست شما با موفقیت ثبت شد. پشتیبانی به زودی با شما تماس خواهد گرفت.")
	bot.Send(msg)
}
