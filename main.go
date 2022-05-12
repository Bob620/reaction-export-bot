package main

import (
	"Shock-Bot/config"
	"Shock-Bot/helpers"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/webview/webview"
	"time"
)

func main() {
	conf := config.LoadConfig("./config/config.json")
	discord, err := discordgo.New(conf.BotToken)
	if err != nil {
		fmt.Println("Unable to connect to discord!")
	}

	discord.Identify.Intents = discordgo.IntentsMessageContent | discordgo.IntentsGuildMessageReactions |
		discordgo.IntentGuildMessages | discordgo.IntentsGuildEmojis | discordgo.IntentsGuilds

	// Prepare the webview
	w := webview.New(true)
	defer w.Destroy()
	w.SetTitle("Shock-Bot")
	w.SetSize(800, 600, webview.HintNone)

	windowReady := make(chan bool, 1)
	bot := MakeBot(&conf, discord, windowReady, w)

	discord.AddHandler(bot.OnReady)

	w.Bind("log", func(t string) {
		fmt.Println(t)
	})
	w.Bind("init", func() {
		if windowReady != nil {
			windowReady <- true
		}
	})

	go func() {
		time.Sleep(time.Second * 2)
		w.Dispatch(func() {
			w.Eval("test();")
		})
	}()

	html, err := helpers.ParseHtml("initialize", nil)
	if err != nil {
		fmt.Println(err)
	}
	html.Navigate(w, []string{})

	// OPen discord connection
	err = discord.Open()
	if err != nil {
		fmt.Println("Error opening Discord session: ", err)
	}

	fmt.Println("Discord Connected")

	// Open the window while we wait for discord to connect
	w.Run()
	w.Destroy()

	// Cleanly close down the Discord session.
	err = discord.Close()
	if err != nil {
		fmt.Println(err)
	}
}
