package config

import (
	"encoding/json"
	"fmt"
	"os"
)

type Config struct {
	BotToken          string     `json:"bot_token"`
	GuildId           string     `json:"guild_id"`
	ChannelId         string     `json:"channel_id"`
	DefaultOutputFile string     `json:"default_output_file"`
	RoleSets          [][]string `json:"role_sets"`
}

func LoadConfig(configLocation string) (conf Config) {
	configFile, err := os.Open(configLocation)
	if err != nil {
		fmt.Println(err)
	}
	defer configFile.Close()

	jsonParser := json.NewDecoder(configFile)
	jsonParser.Decode(&conf)

	return
}
