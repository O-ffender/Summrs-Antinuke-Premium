package events

import (
	"fmt"
	"strings"
	"sync"

	"github.com/bwmarrin/discordgo"
	"github.com/summrs-dev-team/summrs-premium/database"
	"github.com/summrs-dev-team/summrs-premium/utils"
)

func AntiInvite(s *discordgo.Session, m *discordgo.MessageCreate) {
	data, err := database.Database.FindData(m.GuildID)
	switch {
	case err != nil:
		return
	case data["Anti-Invite"] == "off":
		return
	case database.Database.IsWhitelisted(m.GuildID, "whitelisted-invite-channels", m.ChannelID, nil) || utils.HasPerms(s, m.Message, m.GuildID, m.Author.ID, discordgo.PermissionManageMessages):
		return
	}

	if strings.Contains(m.Content, "discord.gg/") {
		s.ChannelMessageDelete(m.ChannelID, m.ID)
	}
}

func BanHandler(s *discordgo.Session, event *discordgo.GuildBanAdd) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["Anti-Ban"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "banned a member", 22)
}

func ChannelCreate(s *discordgo.Session, event *discordgo.ChannelCreate) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["Anti-Channel-Create"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "created a channel", 10)
}

func ChannelRemove(s *discordgo.Session, event *discordgo.ChannelDelete) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["Anti-Channel-Delete"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "deleted a channel", 12)
}

func CreateGuild(s *discordgo.Session, event *discordgo.GuildCreate) {
	muteX.Lock()

	s.State.GuildAdd(event.Guild)
	database.Database.CreateGuild(s.State.User, event.Guild)

	defer muteX.Unlock()

	if _, ok := guilds[event.Guild.ID]; ok {
		return
	}

	guilds[event.Guild.ID] = event.Guild.MemberCount

	MemberCount += guilds[event.Guild.ID]
	GuildCount++
}

func DeleteGuild(s *discordgo.Session, event *discordgo.GuildDelete) {
	database.Database.DeleteGuild(event.Guild.ID)

	MemberCount -= guilds[event.Guild.ID]

	muteX.RLock()
	defer muteX.RUnlock()

	delete(guilds, event.Guild.ID)
	GuildCount--
}

func GuildUpdate(s *discordgo.Session, event *discordgo.GuildUpdate) {
	guildData, err := database.Database.FindData(event.ID)
	if err != nil {
		return
	}

	entry, _, err := utils.FindAudit(s, event.Guild.ID, 1)
	if err != nil && err.Error() != "Whitelisted" {
		return
	}

	for _, change := range entry.Changes {
		var key = *change.Key

		switch key {

		case discordgo.AuditLogChangeKeyName:
			switch {

			case guildData["anti-name-change"].(bool) == false:
				continue

			case change.OldValue == nil:
				continue

			case err != nil && err.Error() == "Whitelisted":
				database.Database.SetData("$set", event.Guild.ID, "guild-name", change.NewValue.(string))

			default:
				s.GuildEdit(event.Guild.ID, discordgo.GuildParams{
					Name: guildData["guild-name"].(string),
				})

			}

		case discordgo.AuditLogChangeKeyVanityURLCode:

			switch {
			case guildData["anti-vanity-steal"].(bool) == false:
				continue
			case change.OldValue == nil:
				continue

			case err != nil && err.Error() == "Whitelisted":
				database.Database.SetData("$set", event.Guild.ID, "vanity_url", change.NewValue.(string))

			default:
				jsonData := []byte(fmt.Sprintf(`{"code":"%s"}`, guildData["vanity_url"]))
				utils.MakeRequest("PATCH", fmt.Sprintf("https://discord.com/api/v7/guilds/%s/vanity-url", event.Guild.ID), s.Token, jsonData)

			}

		case discordgo.AuditLogChangeKeyWidgetEnabled:

			switch {
			case guildData["anti-widget-spam"].(bool) == false:

			case err != nil:
				return

			case change.OldValue.(bool) == change.NewValue.(bool):
				continue

			}

			utils.HandleModeration(s, event.Guild.ID, entry.UserID, fmt.Sprintf("%s - tried to audit log spam", s.State.User.Username))
		}
	}
}

func KickHandler(s *discordgo.Session, event *discordgo.GuildMemberRemove) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-kick"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "kicked a member", 20)
}

func MemberJoin(s *discordgo.Session, event *discordgo.GuildMemberAdd) {
	MemberCount++
	s.State.MemberAdd(event.Member)

	if !event.Member.User.Bot {
		return
	}

	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-bot"].(bool) == false {
		return
	}

	// ANTI-Bot check
	var (
		entry, _, err2 = utils.FindAudit(s, event.GuildID, 28)
	)

	if entry == nil || err2 != nil {
		return
	}

	utils.HandleModeration(s, event.GuildID, entry.UserID, fmt.Sprintf("%s | invited a bot", s.State.User.Username))

	err = utils.HandleModeration(s, event.GuildID, event.User.ID, fmt.Sprintf("%s - invited by someone (Probable nuke bot)", s.State.User.Username))
	if err != nil {
		return
	}

	utils.LogChannel(s, event.GuildID, fmt.Sprintf("<@%s> tried inviting a bot (<@%s>) (Moderation action taken.)", entry.UserID, event.User.ID))
}

func MemberLeave(s *discordgo.Session, event *discordgo.GuildMemberRemove) {
	MemberCount--
	s.State.MemberRemove(event.Member)
}

func MemberRoleUpdate(s *discordgo.Session, event *discordgo.GuildMemberUpdate) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-member-role"].(bool) == false {
		return
	}

	var (
		entry, change, err2 = utils.FindAudit(s, event.GuildID, 25)
	)

	if err2 != nil || change == nil || len(change.([]interface{})) == 0 {
		return
	}

	roleID := change.([]interface{})[0].(map[string]interface{})["id"].(string)

	guildRole, err := s.State.Role(event.GuildID, roleID)
	if err != nil {
		return
	}

	if guildRole.Permissions&0x8 != 0x8 {
		return
	}

	err = s.GuildMemberRoleRemove(event.GuildID, entry.TargetID, roleID)
	if err != nil {
		return
	}

	err = utils.HandleModeration(s, event.GuildID, entry.UserID, fmt.Sprintf("%s | gave a member an admin role", s.State.User.Username))
	if err != nil {
		return
	}

	utils.LogChannel(s, event.GuildID, fmt.Sprintf("<@%s> gave a member an admin role", entry.UserID))
}

func Ready(s *discordgo.Session, event *discordgo.Ready) {
	s.UpdateStreamingStatus(2, fmt.Sprintf(">help | Shard #%d", s.ShardID), "https://twitch.tv/discord")
	fmt.Printf("Connected to shard #%d\n", s.ShardID)
}

func RoleCreate(s *discordgo.Session, event *discordgo.GuildRoleCreate) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-role-create"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "created a role", 30)
}

func RoleRemove(s *discordgo.Session, event *discordgo.GuildRoleDelete) {
	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-role-delete"].(bool) == false {
		return
	}

	utils.ReadAudit(s, event.GuildID, "deleted a role", 32)
}

func WebhookCreate(s *discordgo.Session, event *discordgo.WebhooksUpdate) {

	var (
		err          error
		selfMember   *discordgo.Member
		targetMember *discordgo.Member
	)

	data, err := database.Database.FindData(event.GuildID)
	if err != nil {
		return
	}

	if data["anti-webhook-create"].(bool) == false {
		return
	}

	webhooks, err := s.ChannelWebhooks(event.ChannelID)
	if err != nil {
		return
	}

	for _, webhook := range webhooks {

		selfMember, err = s.State.Member(event.GuildID, s.State.User.ID)
		if err != nil {
			selfMember, err = s.GuildMember(event.GuildID, s.State.User.ID)
			if err != nil {
				return
			}
		}

		targetMember, err = s.State.Member(event.GuildID, webhook.User.ID)
		if err != nil {
			targetMember, err = s.GuildMember(event.GuildID, webhook.User.ID)
			if err != nil {
				return
			}
		}

		whitelisted := database.Database.IsWhitelisted(event.GuildID, "users", webhook.User.ID, targetMember)
		whitelistedChannel := database.Database.IsWhitelisted(event.GuildID, "whitelisted-webhook-channels", event.ChannelID, nil)
		if whitelisted || whitelistedChannel {
			return
		}

		err = s.WebhookDelete(webhook.ID)

		targetHighest := utils.HighestRole(s, event.GuildID, targetMember)
		selfHighest := utils.HighestRole(s, event.GuildID, selfMember)

		if !utils.IsAbove(selfHighest, targetHighest) || !utils.HasPerms(s, nil, event.GuildID, selfMember.User.ID, discordgo.PermissionBanMembers) {
			return
		}

		err = utils.HandleModeration(s, event.GuildID, webhook.User.ID, fmt.Sprintf("%s | created a webhook", s.State.User.Username))
		if err != nil {
			return
		}

		utils.LogChannel(s, event.GuildID, fmt.Sprintf("<@%s> created a webhook", webhook.User.ID))
	}
}

var (
	guilds      = make(map[string]int)
	GuildCount  int
	MemberCount int
	muteX       = &sync.RWMutex{}
)
