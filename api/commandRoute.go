package api

import (
	"github.com/bwmarrin/discordgo"
	"github.com/summrs-dev-team/summrs-premium/commands"
)

func init() {

	commandRoute.Add("antiinvite", commandRoute.AntiInvite, &commands.Config{
		Alias:        []string{"antiinv", "noinvite"},
		Cooldown:     3,
		OwnerOnly:    true,
		RequiresArgs: true,
	})

	commandRoute.Add("avatar", commandRoute.Avatar, &commands.Config{
		Alias:           []string{"av", "pfp", "icon"},
		Cooldown:        1,
		RequiresMention: true,
	})

	commandRoute.Add("ban", commandRoute.Ban, &commands.Config{
		Cooldown:        2,
		RequiresMention: true,
		Perms:           discordgo.PermissionBanMembers,
	})

	commandRoute.Add("banner", commandRoute.ServerBanner, &commands.Config{
		Alias:    []string{"serverbanner", "sbanner"},
		Cooldown: 1,
	})

	commandRoute.Add("botinfo", commandRoute.BotInfo, &commands.Config{
		Cooldown: 4,
	})

	commandRoute.Add("credits", commandRoute.Credits, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("fox", commandRoute.Fox, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("help", commandRoute.Help, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("invite", commandRoute.Invite, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("kick", commandRoute.Kick, &commands.Config{
		Cooldown:        2,
		RequiresMention: true,
		Perms:           discordgo.PermissionBanMembers,
	})

	commandRoute.Add("lockdown", commandRoute.Lockdown, &commands.Config{
		Alias:    []string{"lock"},
		Cooldown: 1,
		Perms:    discordgo.PermissionManageChannels,
	})

	commandRoute.Add("logchannel", commandRoute.LoggingChannel, &commands.Config{
		Alias:     []string{"setlogs", "log"},
		Cooldown:  5,
		OwnerOnly: true,
	})

	commandRoute.Add("massunban", commandRoute.Unban, &commands.Config{
		Alias:    []string{"unbanall"},
		Cooldown: 30,
		Perms:    discordgo.PermissionBanMembers,
	})

	commandRoute.Add("membercount", commandRoute.MemberCount, &commands.Config{
		Alias:    []string{"mc", "members"},
		Cooldown: 1,
	})

	commandRoute.Add("moderationtype", commandRoute.ModerationType, &commands.Config{
		Alias:     []string{"mt"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("nuke", commandRoute.Nuke, &commands.Config{
		Alias:    []string{"nk"},
		Cooldown: 30,
		Perms:    discordgo.PermissionManageChannels,
	})

	commandRoute.Add("ping", commandRoute.Ping, &commands.Config{
		Cooldown: 5,
	})

	commandRoute.Add("prefix", commandRoute.Prefix, &commands.Config{
		Alias:        []string{"setprefix"},
		Cooldown:     3,
		OwnerOnly:    true,
		RequiresArgs: true,
	})

	commandRoute.Add("servericon", commandRoute.ServerIcon, &commands.Config{
		Alias:    []string{"serverpfp", "sicon", "serverpic"},
		Cooldown: 1,
	})

	commandRoute.Add("serverinfo", commandRoute.ServerInfo, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("settings", commandRoute.Settings, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("setup", commandRoute.Setup, &commands.Config{
		Cooldown: 1,
	})

	commandRoute.Add("slowmode", commandRoute.SlowMode, &commands.Config{
		Cooldown:     1,
		RequiresArgs: true,
		Perms:        discordgo.PermissionManageChannels,
	})

	commandRoute.Add("threshold", commandRoute.Threshold, &commands.Config{
		Alias:        []string{"t"},
		Cooldown:     3,
		RequiresArgs: true,
		OwnerOnly:    true,
	})

	commandRoute.Add("toggle", commandRoute.Toggle, &commands.Config{
		Cooldown:     3,
		RequiresArgs: true,
		OwnerOnly:    true,
	})

	commandRoute.Add("unlockdown", commandRoute.UnLockdown, &commands.Config{
		Alias:    []string{"unlock"},
		Cooldown: 1,
		Perms:    discordgo.PermissionManageChannels,
	})

	commandRoute.Add("unslowmode", commandRoute.UnSlowMode, &commands.Config{
		Cooldown: 1,
		Perms:    discordgo.PermissionManageChannels,
	})

	commandRoute.Add("unwhitelist", commandRoute.Unwhitelist, &commands.Config{
		Alias:           []string{"delwl", "removewl", "dewhitelist"},
		Cooldown:        3,
		OwnerOnly:       true,
		RequiresMention: true,
	})

	commandRoute.Add("unwhitelistrole", commandRoute.UnWhitelistRole, &commands.Config{
		Alias:               []string{"delrolewl", "removerolewl", "dewhitelistrole"},
		Cooldown:            3,
		OwnerOnly:           true,
		RequiresRoleMention: true,
	})

	commandRoute.Add("unwhitelistinvite", commandRoute.UnWhitelistInvite, &commands.Config{
		Alias:     []string{"delwebhook", "removewebhook", "dewhitelistwebhook"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("unwhitelistwebhook", commandRoute.UnWhitelistWebhook, &commands.Config{
		Alias:     []string{"delwebhook", "removewebhook", "dewhitelistwebhook"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("userinfo", commandRoute.UserInfo, &commands.Config{
		Alias:           []string{"whois"},
		Cooldown:        1,
		RequiresMention: true,
	})

	commandRoute.Add("whitelist", commandRoute.Whitelist, &commands.Config{
		Alias:           []string{"wl", "addwhitelist", "bypass"},
		Cooldown:        3,
		OwnerOnly:       true,
		RequiresMention: true,
	})

	commandRoute.Add("whitelistinvite", commandRoute.WhitelistInvite, &commands.Config{
		Alias:     []string{"wlinvite", "bypassinvite"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("whitelistrole", commandRoute.WhitelistRole, &commands.Config{
		Alias:     []string{"wlrole", "addrolewhitelist", "bypassrole"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("whitelistwebhook", commandRoute.WhitelistWebhook, &commands.Config{
		Alias:     []string{"wlwebhook", "bypasswebhook"},
		Cooldown:  3,
		OwnerOnly: true,
	})

	commandRoute.Add("whitelisted", commandRoute.ViewWhitelisted, &commands.Config{
		Cooldown:        3,
		WhitelistedOnly: true,
	})

}
