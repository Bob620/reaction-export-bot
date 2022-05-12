package helpers

import (
	"fmt"
	"github.com/bwmarrin/discordgo"
	"regexp"
	"strings"
)

var regReplaceDiscord = regexp.MustCompile(`<:.+?:(\d+)>|<@&(\d+)>`)

type DiscordReplacerForGuild struct {
	session *discordgo.Session
	guild   *discordgo.Guild
	emotes  map[string]*discordgo.Emoji
	roles   map[string]*discordgo.Role
}

func MakeDiscordReplacer(session *discordgo.Session, guildId string) *DiscordReplacerForGuild {
	guild, _ := session.Guild(guildId)
	emotes := map[string]*discordgo.Emoji{}
	roles := map[string]*discordgo.Role{}

	emojis, _ := session.GuildEmojis(guildId)
	for _, emote := range emojis {
		emotes[emote.ID] = emote
	}

	guildRoles, _ := session.GuildRoles(guildId)
	for _, role := range guildRoles {
		roles[role.ID] = role
	}

	return &DiscordReplacerForGuild{
		session: session,
		guild:   guild,
		emotes:  emotes,
		roles:   roles,
	}
}

func (replacer DiscordReplacerForGuild) GetEmote(id string) *discordgo.Emoji {
	return replacer.emotes[id]
}

func (replacer DiscordReplacerForGuild) ReplaceDiscordContent(message string) (string, map[string]*discordgo.Emoji) {
	foundEmotes := map[string]*discordgo.Emoji{}
	for _, str := range regReplaceDiscord.FindAllStringSubmatch(message, -1) {
		emoteId := str[1]
		roleId := str[2]

		// Emote ID
		if emoteId != "" {
			emote := replacer.emotes[emoteId]

			if emote != nil {
				foundEmotes[emote.ID] = emote
				message = strings.ReplaceAll(message, str[0], emote.Name)
			}
		}

		// Role ID
		if roleId != "" {
			if replacer.roles[roleId] != nil {
				message = strings.ReplaceAll(message, str[0], fmt.Sprintf("@%s", replacer.roles[roleId].Name))
			}
		}
	}

	return message, foundEmotes
}
