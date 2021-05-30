package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/summrs-dev-team/summrs-premium/database"
	"github.com/summrs-dev-team/summrs-premium/utils"
	"go.mongodb.org/mongo-driver/bson"
)

func (cmd *Commands) AntiInvite(s *discordgo.Session, m *discordgo.Message, ctx *Context) {
	if !(ctx.Fields[0] == "on" || ctx.Fields[0] == "off") {
		return
	}

	if _, err := database.Database.SetData("$set", m.GuildID, "anti-invite", ctx.Fields[0]); err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set Anti-Invite to %s", ctx.Fields[0]))
}

func (cmd *Commands) LoggingChannel(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if set, err := database.Database.SetData("$set", message.GuildID, "log-channel", message.ChannelID); !set {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Set the logging channel to the current channel")
}

func (cmds *Commands) ModerationType(s *discordgo.Session, m *discordgo.Message, ctx *Context) {
	if !(ctx.Fields[0] == "ban" || ctx.Fields[0] == "kick") {
		return
	}

	if _, err := database.Database.SetData("$set", m.GuildID, "moderation-type", ctx.Fields[0]); err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set Moderation-Type to %s", ctx.Fields[0]))
}

func (cmd *Commands) Prefix(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if set, err := database.Database.SetData("$set", message.GuildID, "prefix", ctx.Fields[0]); !set {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		Title:  fmt.Sprintf("Prefix has been set to `%s`", ctx.Fields[0]),
		Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Requested by: %s", message.Author.Username)},
		Color:  0x36393F,
	})
}

func (cmd *Commands) Settings(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	data, err := database.Database.FindData(message.GuildID)
	guild, _ := s.State.Guild(message.GuildID)

	if err != nil {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}

	var (
		embed = &discordgo.MessageEmbed{
			Title:  fmt.Sprintf("%s current settings", guild.Name),
			Footer: &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Requested by: %s", message.Author.Username)},
			Color:  0x36393F,
		}
		tempValue string
	)

	for index, value := range data {
		if utils.FindInSlice(blacklistedArgs, index) {
			continue
		}

		switch value.(type) {

		case string:
			switch value.(string) {
			case "on":
				//tempValue = "<:enabled:799507631274197022>"
				tempValue = "<:enabled:825799704586878996>"

			case "off":
				tempValue = "<:enabled:825799704586878996>"
				//tempValue = "<:disabled:799507673648594954>"

			case "nil":
				tempValue = "<:disabled:825799718683934791>"
				//tempValue = "<:disabled:799507673648594954>"

			default:
				tempValue = value.(string)
				if index == "log-channel" {
					tempValue = fmt.Sprintf("<#%s>", value.(string))
				}
			}
		case bool:

			//tempValue = "<:disabled:799507673648594954>"
			tempValue = "<:disabled:825799718683934791>"
			if value.(bool) == true {
				tempValue = "<:enabled:825799704586878996>"
				//tempValue = "<:enabled:799507631274197022>"
			}
		}

		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   index,
			Value:  tempValue,
			Inline: true,
		})

	}
	s.ChannelMessageSendEmbed(message.ChannelID, embed)
}

func (cmds *Commands) Threshold(s *discordgo.Session, m *discordgo.Message, ctx *Context) {
	_, err := strconv.Atoi(ctx.Fields[0])
	if err != nil {
		s.ChannelMessageSend(m.ChannelID, "You need a number")
		return
	}

	if _, err := database.Database.SetData("$set", m.GuildID, "offense-threshold", ctx.Fields[0]); err != nil {
		s.ChannelMessageSend(m.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set Offense-Threshold to %s", ctx.Fields[0]))
}

func (cmd *Commands) Toggle(s *discordgo.Session, m *discordgo.Message, ctx *Context) {
	if !utils.FindInSlice(validArgs, ctx.Fields[0]) {
		return
	}

	if len(ctx.Fields) > 1 {
		fmt.Println("test")
		var boolean bool
		switch ctx.Fields[1] {
		case "on":
			boolean = true
		case "off":
			boolean = false
		default:
			s.ChannelMessageSend(m.ChannelID, "Wtf? It's on/off not whatever you put.")
		}

		_, err := database.Database.SetToggle("$set", m.GuildID, ctx.Fields[0], boolean)
		if err != nil {
			s.ChannelMessageSend(m.ChannelID, err.Error())
			return
		}

		s.ChannelMessageSend(m.ChannelID, fmt.Sprintf("Set %s to %v", ctx.Fields[0], boolean))
		return
	}

	s.ChannelMessageSend(m.ChannelID, "You also need to specify on/off")
}

// ** User whitelist ** //
func (cmd *Commands) Whitelist(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.Mentions[0].ID, "$push", "users"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Whitelisted that user.")
}

// ** Invite channel whitelist ** //
func (cmd *Commands) WhitelistInvite(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.ChannelID, "$push", "whitelisted-invite-channels"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Whitelisted this channel for sending discord.gg/ invites (globally).")
}

// ** Role Whitelist ** //
func (cmd *Commands) WhitelistRole(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.MentionRoles[0], "$push", "whitelisted-roles"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Whitelisted that role.")
}

// ** Webhook whitelist ** //
func (cmd *Commands) WhitelistWebhook(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.ChannelID, "$push", "whitelisted-webhook-channels"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Whitelisted this channel for creation of webhooks (globally).")
}

// ** User unwhitelist ** //
func (cmd *Commands) Unwhitelist(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.Mentions[0].ID, "$pull", "users"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}

	s.ChannelMessageSend(message.ChannelID, "Unwhitelisted that user.")
}

// ** Invite channel unwhitelist ** //
func (cmd *Commands) UnWhitelistInvite(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.ChannelID, "$pull", "whitelisted-invite-channels"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Removed the invite whitelist on this channel.")
}

// ** Role unwhitelist ** //
func (cmd *Commands) UnWhitelistRole(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.MentionRoles[0], "$pull", "whitelisted-roles"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Unwhitelisted that role.")
}

// ** Webhook unwhitelist ** //
func (cmd *Commands) UnWhitelistWebhook(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	if whitelisted, err := database.Database.SetWhitelistData(message.GuildID, message.ChannelID, "$push", "whitelisted-webhook-channels"); !whitelisted {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}
	s.ChannelMessageSend(message.ChannelID, "Removed the webhook whitelist on this channel.")
}

func (cmd *Commands) ViewWhitelisted(s *discordgo.Session, message *discordgo.Message, ctx *Context) {
	data, err := database.Database.FindData(message.GuildID)

	if err != nil {
		s.ChannelMessageSend(message.ChannelID, err.Error())
		return
	}

	var whitelistedUsers []string

	for _, userID := range data["users"].(bson.A) {
		member, err := s.State.Member(message.GuildID, userID.(string))
		if err != nil {
			continue
		}

		whitelistedUsers = append(whitelistedUsers, fmt.Sprintf("📋 | %s#%s", member.User.Username, member.User.Discriminator))
	}

	for _, roleID := range data["whitelisted-roles"].(bson.A) {
		whitelistedUsers = append(whitelistedUsers, fmt.Sprintf("📋 | Role: <@&%s>", roleID))
	}

	for _, inviteID := range data["whitelisted-invite-channels"].(bson.A) {
		whitelistedUsers = append(whitelistedUsers, fmt.Sprintf("📋 | Invite channel: <#%s>", inviteID))
	}

	for _, webhookID := range data["whitelisted-webhook-channels"].(bson.A) {
		whitelistedUsers = append(whitelistedUsers, fmt.Sprintf("📋 | Webhook channel: <#%s>", webhookID))
	}

	s.ChannelMessageSendEmbed(message.ChannelID, &discordgo.MessageEmbed{
		Title:       "Whitelisted data",
		Footer:      &discordgo.MessageEmbedFooter{Text: fmt.Sprintf("Requested by: %s", message.Author.Username)},
		Description: strings.Join(whitelistedUsers, "\n"),
		Color:       0x36393F,
	})
}

var (
	blacklistedArgs = []string{
		"users",
		"_id",
		"guild_id",
		"guild-name",
		"vanity-url",
		"whitelisted-roles",
		"whitelisted-invite-channels",
		"whitelisted-webhook-channels",
	}

	validArgs = []string{
		"anti-ban",
		"anti-bot",
		"anti-kick",
		"anti-name-change",
		"anti-widget-spam",
		"anti-member-role",
		"anti-role-create",
		"anti-role-delete",
		"anti-vanity-steal",
		"anti-channel-create",
		"anti-channel-delete",
		"anti-webhook-create",
	}
)
