package main

import (
	"encoding/json"
	"github.com/diamondburned/arikawa/v3/discord"
	"net/http"
	"strings"
)

var (
	commands = map[string]string{
		"ping":  "PingCommand",
		"frog":  "FrogCommand",
		"kirby": "KirbyCommand",
	}
)

func (c Command) PingCommand() {
	msg, err := SendEmbed(c,
		"Ping!",
		"Unfinished", // TODO do
		defaultColor)
	if err != nil {
		_, _ = SendEmbed(c, "Pong!", msg.Timestamp.Format(timeFormat), defaultColor)
	}
}

func (c Command) FrogCommand() error {
	frogData, err := RequestUrl("https://frog.pics/api/random", http.MethodGet)
	if err != nil {
		SendErrorEmbed(c, err)
		return nil
	}

	type FrogPicture struct {
		ImageUrl    string `json:"image_url"`
		MedianColor string `json:"median_color"`
	}
	var frogPicture FrogPicture
	err = json.Unmarshal(frogData, &frogPicture)
	if err != nil {
		SendErrorEmbed(c, err)
		return nil
	}

	color, err := ParseHexColorFast(frogPicture.MedianColor)
	if err != nil {
		SendErrorEmbed(c, err)
		return err
	}

	embed := discord.Embed{
		Color: discord.Color(ConvertColorToInt32(color)),
		Image: &discord.EmbedImage{URL: frogPicture.ImageUrl},
	}
	_, err = SendCustomEmbed(c.e.ChannelID, embed)
	return err
}

func (c Command) KirbyCommand() {
	content := strings.Join(strings.Split(c.e.Content, " ")[1:], " ")
	_, _ = SendMessage(c, "<:kirbyfeet:893291555744542730>")
	_, _ = SendMessage(c, content)
}
