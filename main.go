package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/fatih/structs"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/mitchellh/mapstructure"
	//test comment
)

var (
	users     map[string]interface{} // map ник - статус
	names     map[string]interface{} // map имя - ник
	listUsers map[string]interface{} // map имя - статус, для отчета
	statusMsg string
	parms     map[string]interface{}
)

type listFromJson struct { //под этими именами сохраняется в структуре. Имена переменных должны совпадать с key map names
	Руслан  string
	Дима    string
	Ильнур  string
	Айтуган string
	Айдар   string
	Денис   string
}

type parmsFromJson struct {
	Token     string
	SendTime1 string
	SendTime2 string
	ChatID    string
}

func main() {
	parms = make(map[string]interface{})
	//parms = readFromFileParms("/app/mnt/parms.json")
	parms = readFromFileParms("./parms.json")
	chIDstr := parms["ChatID"].(string)
	chID, err1 := strconv.ParseInt(chIDstr, 10, 64)
	if err1 != nil {
		log.Fatal(err1)
	}

	names = make(map[string]interface{}) // настоящие имена
	//names = readFromFile("/app/mnt/names.json")
	names = readFromFile("./names.json")
	listUsers = make(map[string]interface{}) // map для отчета
	//listUsers = readFromFile("/app/mnt/listUsers.json")
	listUsers = readFromFile("./listUsers.json")

	users = make(map[string]interface{}) // ники в telegram
	for key1, val1 := range listUsers {
		for key2, _ := range names {
			if key1 == key2 {
				strVal := names[key2].(string)
				users[strVal] = val1
			}

		}
	}

	// подключаемся к боту с помощью токена
	bot, err := tgbotapi.NewBotAPI(parms["Token"].(string))
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true
	log.Printf("Authorized on account %s", bot.Self.UserName)

	go sendReport(bot, chID)

	// инициализируем канал, куда будут прилетать обновления от API
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	// читаем обновления из канала
	for update := range updates {

		reply := ""
		if update.Message == nil { // ignore any non-Message Updates
			continue
		}

		// логируем от кого какое сообщение пришло
		//	log.Printf("[%s] %s", update.Message.From.UserName, update.Message.Text)

		msg := tgbotapi.NewMessage(update.Message.Chat.ID, update.Message.Text)
		msg.ReplyToMessageID = update.Message.MessageID

		userNick := update.Message.From.UserName
		userStatus := update.Message.Text

		//sendUserState := tgbotapi.NewMessage(update.Message.Chat.ID, userState)

		if strings.Contains(userStatus, "/") {
			switch update.Message.Command() {
			case "status_sect": // пришлет статус всех в меню бота
				reply = fmt.Sprintln(listUsers)
				replyMsg := tgbotapi.NewMessage(update.Message.Chat.ID, reply)
				bot.Send(replyMsg)
			case "send": // пошлет в чат chatID статус
				go sendNotifications(bot, chID)
			case "read":
				//listUsers = readFromFile("/app/mnt/listUsers.json")
				listUsers = readFromFile("./listUsers.json")
			case "write":
				writeToFile(listUsers)
			}
		} else {
			if _, ok := users[userNick]; ok {
				users[userNick] = userStatus
				for key1, val1 := range users {
					for key2, val2 := range names {
						if key1 == val2 {
							listUsers[key2] = val1
						}
					}
				}
				writeToFile(listUsers)
			}
		}
	}
}
func sendNotifications(bot *tgbotapi.BotAPI, chID int64) {
	var stringMap string
	stringMap = mapToList(listUsers)

	bot.Send(tgbotapi.NewMessage(chID, stringMap))
}

func sendReport(bot *tgbotapi.BotAPI, chID int64) {
	for {
		t := time.Now()
		cTime := t.Format(time.Kitchen)

		if cTime == parms["SendTime1"].(string) {
			sendNotifications(bot, chID)
			time.Sleep(time.Minute * 1)
		}
		if cTime == parms["SendTime2"].(string) {
			sendNotifications(bot, chID)
			time.Sleep(time.Minute * 1)
		}
	}
}

func mapToList(m map[string]interface{}) string {
	var str string
	var strAll string
	for key, val := range m {
		str = fmt.Sprintf("%s - %s", key, val)
		strAll = strAll + "\n" + str
	}
	return strAll
}

func readFromFile(s string) map[string]interface{} { //принимает имя файла, возвращает map
	file, err1 := os.Open(s)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file.Close()
	jsonString, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		log.Fatal(err2)
	}
	var str listFromJson
	err3 := json.Unmarshal(jsonString, &str)
	if err3 != nil {
		log.Fatal(err3)
	}
	m := structs.Map(str)
	return m
}

func readFromFileParms(s string) map[string]interface{} {
	file, err1 := os.Open(s)
	if err1 != nil {
		log.Fatal(err1)
	}
	defer file.Close()
	jsonString, err2 := ioutil.ReadAll(file)
	if err2 != nil {
		log.Fatal(err2)
	}
	var str parmsFromJson
	err3 := json.Unmarshal(jsonString, &str)
	if err3 != nil {
		log.Fatal(err3)
	}
	m := structs.Map(str)
	return m
}

func writeToFile(m map[string]interface{}) {
	var str listFromJson
	err1 := mapstructure.Decode(m, &str)
	if err1 != nil {
		panic(err1)
	}
	jsn, err2 := json.Marshal(str)
	if err2 != nil {
		panic(err2)
	}
	//err3 := ioutil.WriteFile("/app/mnt/listUsers.json", jsn, 0644)
	err3 := ioutil.WriteFile("./listUsers.json", jsn, 0644)
	if err3 != nil {
		panic(err3)
	}
}
