package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
)

type UserInterfaceStruct struct {
	Log     string `json:"login"`
	Pasw    string `json:"password"`
	Address string `json:"address"`
	Port    string `json:"port"`
	Room    string `json:"room"`
}

var (
	GuiApp UserInterfaceStruct
	Wg     sync.WaitGroup
)

func main() {

	err := loader()
	if err != nil {
		log.Printf("Error while loading config: %v", err)
	}

	chatApp := app.New()

	chatWindow := chatApp.NewWindow("Кусок GoВнища")
	chatWindow.Resize(fyne.NewSize(800, 480))

	// layout group

	entryMessage := widget.NewEntry()
	entryMessage.PlaceHolder = "type something..."

	applicationHolder := widget.NewMultiLineEntry()
	applicationHolder.Disable()

	yourRoomName := widget.NewEntry()
	yourRoomName.Disable()

	yourRoomName.SetText("Room is: " + GuiApp.Room)

	// logic start

	full := fmt.Sprintf("%s:%s", GuiApp.Address, GuiApp.Port)
	conn, err := net.DialTimeout("tcp", full, 3*time.Second)
	if err != nil {
		log.Printf("Error while dialing conn, retry: %v", err)
		return
	}
	defer conn.Close()

	Wg.Add(1)
	go func() {
		defer Wg.Done()
		_, err = conn.Write([]byte("/kkk12111dddqqqprstlon " + GuiApp.Log + " " + GuiApp.Room)) // sends package with main data
	}()

	Wg.Wait() // cuz i can

	go func() {
		for {
			buffer := make([]byte, 8192)

			n, err := conn.Read(buffer)
			if err != nil {
				log.Printf("%v", err)
				return
			}

			appendMessage := string(buffer[:n]) + "\n"

			if strings.HasPrefix(appendMessage, "///") {
				roomName := strings.SplitN(appendMessage, " ", 2)
				yourRoomName.SetText(roomName[1])
				GuiApp.Room = strings.TrimSpace(roomName[1])

				err := updateRoomTitleJsn()
				if err != nil {
					log.Printf("Error while updating room title: %v", err)
				}

				continue // skips appending
			}

			applicationHolder.Append(appendMessage)
		}
	}()

	entryMessage.OnSubmitted = func(content string) {
		if content == "" {
			return
		}
		appMsg := time.Now().Format("15:04:05 02.01.2006") + " You" + ": " + content + "\n"

		applicationHolder.Append(appMsg)

		if strings.HasPrefix(content, "/") {
			_, err = conn.Write([]byte(content + "\n"))
		} else if strings.HasPrefix(content, "@bot") {
			_, err = conn.Write([]byte(content + "\n"))
		} else {
			_, err = conn.Write([]byte(time.Now().Format("15:04:05 02.01.2006") + GuiApp.Log + ": " + content + "\n"))

			if err != nil {
				log.Printf("%v", err)
			}

		}
		entryMessage.SetText("")
	}

	// logic ends here (i mean connection, logic and etc.)
	winLayout := container.NewBorder(yourRoomName, entryMessage, nil, nil, applicationHolder)

	chatWindow.SetContent(winLayout)
	chatWindow.ShowAndRun()
}

func loader() error {
	bytes, err := os.ReadFile("login.json")
	if err != nil {
		return err
	}

	err = json.Unmarshal(bytes, &GuiApp)
	if err != nil {
		return err
	}

	return nil
}

func updateRoomTitleJsn() error {
	data, err := json.MarshalIndent(GuiApp, "", " ")
	if err != nil {
		log.Printf("Marshaling error idk!")
		return err
	}

	timeFile := "login.json.tmp"
	err = os.WriteFile(timeFile, data, 0644)
	if err != nil {
		log.Printf("Writing error")
		return err
	}

	loader() // calls loader

	return os.Rename("login.json.tmp", "login.json")
}
