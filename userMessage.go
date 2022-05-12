package main

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"strings"
	"time"
)

type UserMessage struct {
	Expires         *time.Time
	Guild           *discordgo.Guild
	Message         *discordgo.Message
	SplitMessage    []string
	Session         *discordgo.Session
	MakeUserMessage func(*discordgo.Message, *time.Time) *UserMessage
}

func MakeUserMessage(session *discordgo.Session, makeUserMessage func(*discordgo.Message, *time.Time) *UserMessage, message *discordgo.Message, expires *time.Time) *UserMessage {
	var guild *discordgo.Guild
	if message.GuildID == "" {
		guild = &discordgo.Guild{
			ID: "dm",
		}
	} else {
		guild, _ = session.Guild(message.GuildID)
	}

	var split []string
	for _, word := range strings.Split(message.Content, " ") {
		if word != "" {
			split = append(split, word)
		}
	}

	return &UserMessage{
		Expires:         expires,
		Guild:           guild,
		Message:         message,
		SplitMessage:    split,
		Session:         session,
		MakeUserMessage: makeUserMessage,
	}
}

func (userMessage *UserMessage) Reply(content string) (*UserMessage, error) {
	msg, err := userMessage.Session.ChannelMessageSendReply(userMessage.Message.ChannelID, content, userMessage.Message.Reference())
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		msg.GuildID = userMessage.Message.GuildID
		return userMessage.MakeUserMessage(msg, nil), err
	}
}

func (userMessage *UserMessage) Response(content string) (*UserMessage, error) {
	msg, err := userMessage.Session.ChannelMessageSend(userMessage.Message.ChannelID, content)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		msg.GuildID = userMessage.Message.GuildID
		return userMessage.MakeUserMessage(msg, nil), err
	}
}

func (userMessage *UserMessage) ReplyComplex(message *discordgo.MessageSend) (*UserMessage, error) {
	message.Reference = userMessage.Message.Reference()
	msg, err := userMessage.Session.ChannelMessageSendComplex(userMessage.Message.ChannelID, message)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		msg.GuildID = userMessage.Message.GuildID
		return userMessage.MakeUserMessage(msg, nil), err
	}
}

func (userMessage *UserMessage) ResponseComplex(message *discordgo.MessageSend) (*UserMessage, error) {
	msg, err := userMessage.Session.ChannelMessageSendComplex(userMessage.Message.ChannelID, message)
	if err != nil {
		fmt.Println(err)
		return nil, err
	} else {
		msg.GuildID = userMessage.Message.GuildID
		return userMessage.MakeUserMessage(msg, nil), err
	}
}

func (userMessage *UserMessage) Edit(message *discordgo.MessageSend) {
	_, _ = userMessage.Session.ChannelMessageEditComplex(&discordgo.MessageEdit{
		Content:         &message.Content,
		Components:      message.Components,
		Embeds:          message.Embeds,
		AllowedMentions: message.AllowedMentions,
		ID:              userMessage.Message.ID,
		Channel:         userMessage.Message.ChannelID,
		Embed:           message.Embed,
	})
}

func (userMessage *UserMessage) EditContent(content string) {
	_, _ = userMessage.Session.ChannelMessageEdit(userMessage.Message.ChannelID, userMessage.Message.ID, content)
}
