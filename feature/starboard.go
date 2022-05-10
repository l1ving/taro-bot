package feature

import (
	"github.com/5HT2/taro-bot/bot"
	"github.com/5HT2/taro-bot/cmd"
	"github.com/5HT2/taro-bot/util"
	"github.com/diamondburned/arikawa/v3/discord"
	"github.com/diamondburned/arikawa/v3/gateway"
	"log"
	"strconv"
	"strings"
	"time"
)

var (
	stars3Emoji = "⭐"
	stars5Emoji = "🌟"
	stars6Emoji = "💫"
	stars9Emoji = "✨"
)

func StarboardReactionHandler(e *gateway.MessageReactionAddEvent) {
	defer util.LogPanic()

	start := time.Now().UnixMilli()

	bot.GuildContext(e.GuildID, func(g *bot.GuildConfig) (*bot.GuildConfig, string) {
		if g.Starboard.Threshold == 0 {
			g.Starboard.Threshold = 3
		}

		// Not starred by a guild member
		if e.Member == nil {
			log.Printf("Not a guild member\n")
			return g, "StarboardReactionHandler: check guild member"
		}

		// Not a star
		if e.Emoji.APIString().PathString() != util.EscapedStar {
			log.Printf("Not a star emoji\n")
			return g, "StarboardReactionHandler: check reaction emoji"
		}

		msg, err := bot.Client.Message(e.ChannelID, e.MessageID)
		if err != nil {
			return g, "StarboardReactionHandler: get reaction message"
		}
		channel, err := bot.Client.Channel(e.ChannelID)
		if err != nil {
			return g, "StarboardReactionHandler: get reaction channel"
		}

		var sMsg *bot.StarboardMessage = nil
		newPost := true
		cID := int64(channel.ID)

		log.Printf("Checking channel for starboard message %s\n", cmd.CreateMessageLink(int64(e.GuildID), msg, false))

		// If user reacts to a post in a starboard channel
		if cID == g.Starboard.Channel || cID == g.Starboard.NsfwChannel {
			for _, m := range g.Starboard.Messages {
				if m.PostID == int64(msg.ID) {
					sMsg = &m
					newPost = false
					break
				}
			}
		} else { // else if a user reacts to a post in a regular channel
			for _, m := range g.Starboard.Messages {
				if m.ID == int64(msg.ID) {
					sMsg = &m
					newPost = false
					break
				}
			}

			// If starred before channel ID was added, and the reaction is from the origin channel, update the stored one
			if !newPost && sMsg.CID == 0 {
				sMsg.CID = int64(msg.ChannelID)
			}
		}

		if newPost {
			sMsg = &bot.StarboardMessage{
				Author: int64(msg.Author.ID),
				CID:    int64(msg.ChannelID),
				ID:     int64(msg.ID),
				PostID: 0,
				IsNsfw: channel.NSFW,
				Stars:  make([]int64, 0),
			}
		}

		// Channel to send starboard message to
		cID = g.Starboard.Channel
		if sMsg.IsNsfw == true {
			cID = g.Starboard.NsfwChannel
		}

		// Channel hasn't been set
		if cID == 0 {
			log.Printf("Channel ID is 0\n")
			return g, "StarboardReactionHandler: check cID"
		}

		// Get post channel and ensure it exists
		postChannel, err := bot.Client.Channel(discord.ChannelID(cID))
		if err != nil {
			log.Printf("Couldn't get post channel\n")
			return g, "StarboardReactionHandler: get post channel"
		}

		// When adding a new star, ensure star user is not the same as author
		// And also check if they've already been added
		sUserID := int64(e.Member.User.ID)
		if sMsg.Author != sUserID && !util.SliceContains(sMsg.Stars, sUserID) {
			sMsg.Stars = append(sMsg.Stars, sUserID)
		}
		log.Printf("sUserID: %v\nsMsg:%v\n", sUserID, sMsg)

		// Update our reactions in case any are missing from the API
		for _, reaction := range msg.Reactions {
			if reaction.Emoji.APIString().PathString() == util.EscapedStar {
				userReactions, err := bot.Client.Reactions(msg.ChannelID, msg.ID, reaction.Emoji.APIString(), 0)
				if err != nil {
					log.Printf("Failed to get userReactions: %s\n", err)
					return g, "StarboardReactionHandler: update sMsg.Stars"
				}

				for _, userReaction := range userReactions {
					sUserID = int64(userReaction.ID)

					if sMsg.Author != sUserID && !util.SliceContains(sMsg.Stars, sUserID) {
						sMsg.Stars = append(sMsg.Stars, sUserID)
					}
				}
				break
			}
		}

		stars := len(sMsg.Stars)

		// Not enough stars in sMsg to make post
		if int64(stars) < g.Starboard.Threshold {
			log.Printf("Not enough stars: %v\n", sMsg.Stars)
			return g, "StarboardReactionHandler: check notEnoughStars"
		}

		content := getEmoji(stars) + " **" + strconv.Itoa(stars) + "** <#" + strconv.FormatInt(sMsg.CID, 10) + ">"

		// Attempt to get existing message, and make a new one if it isn't there
		pMsg, err := bot.Client.Message(postChannel.ID, discord.MessageID(sMsg.PostID))
		if err != nil {
			log.Printf("Couldn't get pMsg %v\n", err)

			// Construct new starboard post if it couldn't retrieve an existing one

			// Try to find a URL in the message content
			description := msg.Content
			url := cmd.UrlRegex.MatchString(msg.Content)

			// Set the embed image to the URL and try to find the first attached image in the message attachments
			var image *discord.EmbedImage = nil
			for _, attachment := range msg.Attachments {
				if strings.HasPrefix(attachment.ContentType, "image/") {
					image = &discord.EmbedImage{URL: attachment.URL}
					url = false // Don't remove URL in embed if we found an image attachment (eg, twitter link + image attachment)
					break
				}
			}

			// If we found only a URL (no other text) in the message content, and the found URL has an image extension, and we didn't find an attached image
			// Set the description to nothing and set the image to the found URL
			if url && util.FileExtMatches(util.ImageExtensions, msg.Content) {
				description = ""
				image = &discord.EmbedImage{URL: msg.Content}
			}

			member, err := bot.Client.Member(e.GuildID, discord.UserID(sMsg.Author))
			if err != nil {
				log.Printf("Couldn't get member %v\n", err)
				return g, "StarboardReactionHandler: get sMsg.Author"
			}

			field := discord.EmbedField{Name: "Source", Value: cmd.CreateMessageLink(int64(e.GuildID), msg, true)}
			footer := discord.EmbedFooter{Text: strconv.FormatInt(sMsg.Author, 10)}
			embed := discord.Embed{
				Description: description,
				Author:      cmd.CreateEmbedAuthor(*member),
				Fields:      []discord.EmbedField{field},
				Footer:      &footer,
				Timestamp:   msg.Timestamp,
				Color:       bot.StarboardColor,
				Image:       image,
			}

			log.Printf("Embed image: %v\n", embed.Image)

			msg, err = bot.Client.SendMessage(postChannel.ID, content, embed)
			if err != nil {
				log.Printf("Error sending starboard post: %v\n", err)
			} else {
				sMsg.PostID = int64(msg.ID)
			}
		} else {
			// Edit the post if it exists
			_, err = bot.Client.EditMessage(postChannel.ID, discord.MessageID(sMsg.PostID), content, pMsg.Embeds...)
			if err != nil {
				log.Printf("Error updating starboard post: %v\n", err)
			}
		}

		// Now that we have updated the stars and starboard post ID, save it in the config
		if newPost {
			g.Starboard.Messages = append(g.Starboard.Messages, *sMsg)
		} else {
			for i, m := range g.Starboard.Messages {
				if m.ID == sMsg.ID {
					g.Starboard.Messages[i] = *sMsg
				}
			}
		}

		return g, "StarboardReactionHandler: update post"
	})

	log.Printf("Execute: %vms (StarboardReactionHandler)\n", time.Now().UnixMilli()-start)
}

func getEmoji(stars int) (emoji string) {
	switch stars {
	case 0, 1, 2, 3, 4:
		emoji = stars3Emoji
	case 5:
		emoji = stars5Emoji
	case 6, 7, 8:
		emoji = stars6Emoji
	default:
		emoji = stars9Emoji
	}

	return emoji
}