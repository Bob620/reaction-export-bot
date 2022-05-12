package message

import (
	"github.com/bwmarrin/discordgo"
	"time"
)

type Message struct {
	ID        string
	Content   string
	Author    Author
	Time      time.Time
	Selected  bool
	Reactions []*discordgo.MessageReactions
}

type Messages []*Message

type Author struct {
	ID            string
	Name          string
	Discriminator string
}

func (messages Messages) GetSelected() (mes Messages) {
	for _, message := range messages {
		if message.Selected {
			mes = append(mes, message)
		}
	}
	return mes
}

func (messages Messages) Get(id string) *Message {
	for _, message := range messages {
		if message.ID == id {
			return message
		}
	}
	return nil
}

func (messages Messages) Len() int {
	return len(messages)
}

func (messages Messages) Swap(i, j int) {
	messages[i], messages[j] = messages[j], messages[i]
}

func (messages Messages) Less(i, j int) bool {
	return messages[i].Time.Before(messages[j].Time)
}
