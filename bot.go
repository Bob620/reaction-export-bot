package main

import (
	"Shock-Bot/config"
	"Shock-Bot/helpers"
	"Shock-Bot/message"
	"fmt"
	"github.com/bwmarrin/discordgo"
	"github.com/webview/webview"
	"golang.org/x/exp/slices"
	"html/template"
	"os"
	"sort"
	"strings"
)

type UIState struct {
	Guilds          map[string]*Guild
	GuildList       []*Guild
	SelectedGuild   *Guild
	SelectedChannel *discordgo.Channel
	Messages        message.Messages
}

type Bot struct {
	config      *config.Config
	session     *discordgo.Session
	state       UIState
	windowReady chan bool
	w           webview.WebView
}

type Guild struct {
	ID       string
	Name     string
	IconUrl  string
	Channels []*discordgo.Channel
	Resolver *helpers.DiscordReplacerForGuild
}

type Person struct {
	ID            string
	Name          string
	Discriminator string
	EmoteCount    map[string]*discordgo.Emoji
}

func MakeBot(config *config.Config, session *discordgo.Session, windowReady chan bool, w webview.WebView) *Bot {
	return &Bot{
		config:      config,
		session:     session,
		state:       UIState{},
		windowReady: windowReady,
		w:           w,
	}
}

func (bot *Bot) OnReady(session *discordgo.Session, ready *discordgo.Ready) {
	html, err := helpers.ParseHtml("index", template.FuncMap{
		"selectedGuild": func(id string) bool {
			return id == bot.state.SelectedGuild.ID
		},
		"selectedChannel": func(id string) bool {
			return id == bot.state.SelectedChannel.ID
		},
		"getMessageContent": func(message *discordgo.Message) string {
			out, _ := message.ContentWithMoreMentionsReplaced(session)
			return template.HTMLEscapeString(out)
		},
	})
	if err != nil {
		fmt.Println(err)
	}

	if bot.windowReady != nil {
		fmt.Println("Waiting on window to initialize")
		<-bot.windowReady
		bot.windowReady = nil
	}
	fmt.Println("Window initialized")

	guilds := map[string]*Guild{}
	var guildList []*Guild

	for _, guild := range session.State.Guilds {
		var channels []*discordgo.Channel
		for _, channel := range guild.Channels {
			if channel.Type == discordgo.ChannelTypeGuildText || channel.Type == discordgo.ChannelTypeGuildNews {
				channels = append(channels, channel)
			}
		}

		guilds[guild.ID] = &Guild{
			ID:       guild.ID,
			Name:     guild.Name,
			IconUrl:  guild.IconURL(),
			Channels: channels,
			Resolver: helpers.MakeDiscordReplacer(session, guild.ID),
		}
		guildList = append(guildList, guilds[guild.ID])
	}

	bot.w.Dispatch(func() {
		bot.w.Bind("Export", func() {
			selectedMessages := bot.state.Messages.GetSelected()
			reactions := map[string]Person{}

			MemberCache := map[string]*struct {
				HasRoles bool
				Member   *discordgo.Member
			}{}

			for _, mes := range selectedMessages {
				if len(mes.Reactions) > 1 {
					for _, reaction := range mes.Reactions {
						reacts, _ := session.MessageReactions(bot.state.SelectedChannel.ID, mes.ID, reaction.Emoji.APIName(), 100, "", "")
						for _, user := range reacts {
							if MemberCache[user.ID] == nil {
								member, err := session.GuildMember(bot.state.SelectedGuild.ID, user.ID)
								if err != nil {
									break
								}
								hasRoles := true
								for _, possibility := range bot.config.RoleSets {
									worked := true
									for _, role := range possibility {
										if !slices.Contains(member.Roles, role) {
											worked = false
										}
										if !worked {
											break
										}
									}
									if worked {
										hasRoles = true
										break
									}
								}
								MemberCache[user.ID] = &struct {
									HasRoles bool
									Member   *discordgo.Member
								}{
									HasRoles: hasRoles,
									Member:   member,
								}
							}

							if MemberCache[user.ID].HasRoles {
								if reactions[user.ID].ID == "" {
									reactions[user.ID] = Person{
										ID:            user.ID,
										Name:          user.Username,
										Discriminator: user.Discriminator,
										EmoteCount: map[string]*discordgo.Emoji{
											mes.ID: reaction.Emoji,
										},
									}
								} else {
									reactions[user.ID].EmoteCount[mes.ID] = reaction.Emoji
								}
							}
						}
					}
				}
			}

			file, err := os.OpenFile("output.csv", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0777)
			if err != nil {
				return
			}

			for _, person := range reactions {
				line := fmt.Sprintf("%s#%s", person.Name, person.Discriminator)
				for _, mes := range selectedMessages {
					emote := person.EmoteCount[mes.ID]
					if emote != nil {
						line = line + fmt.Sprintf(", %s", emote.Name)
					} else {
						line = line + ", "
					}
				}
				_, _ = file.WriteString(line + "\n")
			}

			_ = file.Close()
			bot.w.Dispatch(func() {
				bot.w.Eval("exportEnd()")
			})
		})

		bot.w.Bind("ToggleMessage", func(id string, on bool) {
			mes := bot.state.Messages.Get(id)
			if mes != nil {
				mes.Selected = on
			}
		})

		bot.w.Bind("SelectGuild", func(guildId string) {
			if bot.state.Guilds[guildId] != nil {
				bot.state.SelectedGuild = bot.state.Guilds[guildId]
				html.Navigate(bot.w, bot.state)
			}
		})

		bot.w.Bind("SelectChannel", func(channelId string) {
			if bot.state.SelectedGuild != nil {
				for _, channel := range bot.state.SelectedGuild.Channels {
					if channel.ID == channelId {
						bot.state.SelectedChannel = channel
						rawMessages, err := session.ChannelMessages(channelId, 100, channel.LastMessageID, "", "")
						if err != nil {
							fmt.Println(err)
						}
						rawMessage, err := session.ChannelMessage(channelId, channel.LastMessageID)
						if err != nil {
							fmt.Println(err)
						} else {
							rawMessages = append(rawMessages, rawMessage)
						}

						var messages message.Messages
						if len(rawMessages) != 0 {
							for _, raw := range rawMessages {
								out, _ := raw.ContentWithMoreMentionsReplaced(session)
								out, emotes := bot.state.SelectedGuild.Resolver.ReplaceDiscordContent(out)

								var reactions []*discordgo.MessageReactions
								for _, reaction := range raw.Reactions {
									emote := emotes[reaction.Emoji.ID]
									if emote != nil {
										reactions = append(reactions, reaction)
									}
								}
								messages = append(messages, &message.Message{
									ID:      raw.ID,
									Content: strings.ReplaceAll(strings.ReplaceAll(out, "'", "%27"), "#", "%23"),
									Author: message.Author{
										ID:            raw.Author.ID,
										Name:          raw.Author.Username,
										Discriminator: raw.Author.Discriminator,
									},
									Time:      raw.Timestamp,
									Reactions: reactions,
								})
							}

							sort.Sort(sort.Reverse(messages))
						}
						bot.state.Messages = messages
						break
					}
				}
			}
			html.Navigate(bot.w, bot.state)
		})
	})

	var DefaultGuild *Guild
	var DefaultChannel *discordgo.Channel
	if guilds[bot.config.GuildId] != nil {
		DefaultGuild = guilds[bot.config.GuildId]
		bot.state.SelectedGuild = DefaultGuild
		for _, channel := range guilds[bot.config.GuildId].Channels {
			if channel.ID == bot.config.ChannelId {
				DefaultChannel = channel
				break
			}
		}
	}

	var messages message.Messages
	var rawMessages []*discordgo.Message
	if DefaultChannel != nil {
		rawMessages, err = session.ChannelMessages(DefaultChannel.ID, 100, DefaultChannel.LastMessageID, "", "")
		if err != nil {
			fmt.Println(err)
		}
		rawMessage, err := session.ChannelMessage(DefaultChannel.ID, DefaultChannel.LastMessageID)
		if err != nil {
			fmt.Println(err)
		}

		rawMessages = append(rawMessages, rawMessage)
	}

	for _, raw := range rawMessages {
		out, _ := raw.ContentWithMoreMentionsReplaced(session)
		out, emotes := bot.state.SelectedGuild.Resolver.ReplaceDiscordContent(out)

		var reactions []*discordgo.MessageReactions
		for _, reaction := range raw.Reactions {
			emote := emotes[reaction.Emoji.ID]
			if emote != nil {
				reactions = append(reactions, reaction)
			}
		}

		messages = append(messages, &message.Message{
			ID:      raw.ID,
			Content: strings.ReplaceAll(strings.ReplaceAll(out, "'", "%27"), "#", "%23"),
			Author: message.Author{
				ID:            raw.Author.ID,
				Name:          raw.Author.Username,
				Discriminator: raw.Author.Discriminator,
			},
			Time:      raw.Timestamp,
			Reactions: reactions,
		})
	}
	sort.Sort(sort.Reverse(messages))

	bot.state = UIState{
		Guilds:          guilds,
		GuildList:       guildList,
		SelectedGuild:   DefaultGuild,
		SelectedChannel: DefaultChannel,
		Messages:        messages,
	}

	html.Navigate(bot.w, bot.state)
}
