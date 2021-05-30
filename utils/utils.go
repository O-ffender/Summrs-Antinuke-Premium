package utils

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/summrs-dev-team/summrs-premium/database"
)

func FindAudit(s *discordgo.Session, guildID string, auditType int) (*discordgo.AuditLogEntry, interface{}, error) {
	if !HasPerms(s, nil, guildID, s.State.User.ID, discordgo.PermissionViewAuditLogs) {
		return nil, nil, fmt.Errorf("No perms in %s", guildID) // useless as we don't have error handling but eh
	}

	audit, err := s.GuildAuditLog(guildID, "", "", auditType, 25)

	if err != nil || len(audit.AuditLogEntries) == 0 {
		return nil, nil, fmt.Errorf("Error in fetching audit log entries")
	}

	auditLog := audit.AuditLogEntries[0]

	if len(auditLog.Changes) == 0 {
		return auditLog, []interface{}{}, nil
	}

	var (
		targetMember *discordgo.Member
	)
	targetMember, err = s.State.Member(guildID, auditLog.UserID)
	if err != nil {
		targetMember, err = s.GuildMember(guildID, auditLog.UserID)
		if err != nil {
			return nil, nil, err
		}

		return nil, nil, err
	}

	if whitelisted := database.Database.IsWhitelisted(guildID, "users", auditLog.UserID, targetMember); whitelisted {
		return auditLog, auditLog.Changes[0].NewValue, fmt.Errorf("Whitelisted")
	}

	current := time.Now()
	entryTime, err := discordgo.SnowflakeTimestamp(auditLog.ID)
	if err != nil {
		return nil, nil, err
	}

	if current.Sub(entryTime).Round(time.Second).Seconds() > 2 {
		return nil, nil, fmt.Errorf("Too late")
	}

	return auditLog, auditLog.Changes[0].NewValue, nil
}

func FindInSlice(slice []string, item string) bool {
	for _, i := range slice {
		if i == item {
			return true
		}
	}
	return false
}

func GetGuildOwner(s *discordgo.Session, guildID string) string {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return ""
	}
	return guild.OwnerID
}

func HandleModeration(s *discordgo.Session, guildID, userID, reason string) error {
	data, err := database.Database.FindData(guildID)
	if err != nil {
		return err
	}

	if data["moderation-type"] == nil {
		database.Database.SetData("$set", guildID, "moderation-type", "ban")
	}

	if data["moderation-type"] == "ban" {
		return s.GuildBanCreateWithReason(guildID, userID, reason, 0)
	}

	return s.GuildMemberDeleteWithReason(guildID, userID, reason)
}

func HasPerms(s *discordgo.Session, m *discordgo.Message, guildID, userID string, permissions ...int) bool {
	if GetGuildOwner(s, guildID) == userID {
		return true
	}

	var (
		err   error
		guild *discordgo.Guild
		perms int64
	)

	switch m != nil {
	case true:

		perms, err = s.State.MessagePermissions(m)
		if err != nil {
			return false
		}

	case false:

		guild, err = s.State.Guild(guildID)
		if err != nil {
			return false
		}

		if len(guild.Channels) == 0 {
			return false
		}

		perms, err = s.State.UserChannelPermissions(userID, guild.Channels[0].ID)
		if err != nil {
			return false
		}
	}

	for _, perm := range permissions {
		if perms&(int64(perm)) != int64(perm) {
			continue
		}
		return true
	}

	return false
}

func HighestRole(s *discordgo.Session, guildID string, member *discordgo.Member) *discordgo.Role {
	guild, err := s.State.Guild(guildID)
	if err != nil {
		return nil
	}

	var highest *discordgo.Role
	for _, roleID := range member.Roles {
		for _, role := range guild.Roles {
			if roleID != role.ID {
				continue
			}
			if highest == nil || IsAbove(role, highest) {
				highest = role
			}
			break
		}
	}
	if highest == nil {
		defaultRole, _ := s.State.Role(guildID, guildID)
		return defaultRole
	}
	return highest
}

func IsAbove(r, r2 *discordgo.Role) bool {

	switch {
	case r.Position != r2.Position:
		return r.Position > r2.Position
	case r.ID == r2.ID:
		return false
	}

	return r.Position < r2.Position
}

func LogChannel(s *discordgo.Session, guildID, postData string) {
	data, err := database.Database.FindData(guildID)
	if err != nil {
		return
	}

	if data["log-channel"] == "nil" {
		return
	}

	s.ChannelMessageSend(data["log-channel"].(string), postData)
}

func MakeRequest(method string, url string, token string, body []byte) (resBody []byte, err error) {
	var res *http.Response

	req, err := http.NewRequest(method, url, bytes.NewBuffer(body))
	if err != nil {
		return
	}

	if len(token) > 0 {
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", token)
	}

	res, err = http.DefaultClient.Do(req)
	if err != nil {
		return
	}

	resBody, err = ioutil.ReadAll(res.Body)
	if err != nil {
		return
	}
	defer res.Body.Close()

	return
}

func ReadAudit(s *discordgo.Session, guildID, reason string, auditType int) {
	if !HasPerms(s, nil, guildID, s.State.User.ID, discordgo.PermissionViewAuditLogs) {
		return
	}

	var (
		audits, err  = s.GuildAuditLog(guildID, "", "", auditType, 25)
		auditMap     []string
		userMap      = make(map[string]int)
		selfMember   *discordgo.Member
		targetMember *discordgo.Member
	)

	if err != nil {
		return
	}

	guildData, err := database.Database.FindData(guildID)
	if err != nil {
		return
	}

	if guildData["offense-threshold"] == nil {
		database.Database.SetData("$set", guildID, "offense-threshold", "1")
		guildData, err = database.Database.FindData(guildID)
	}

	for _, entry := range audits.AuditLogEntries {

		selfMember, err = s.State.Member(guildID, s.State.User.ID)
		if err != nil {
			selfMember, err = s.GuildMember(guildID, entry.UserID)
			if err != nil {
				return
			}
		}

		targetMember, err = s.State.Member(guildID, entry.UserID)
		if err != nil {
			targetMember, err = s.GuildMember(guildID, entry.UserID)
			if err != nil {
				return
			}
		}

		if whitelisted := database.Database.IsWhitelisted(guildID, "users", entry.UserID, targetMember); whitelisted {
			return
		}

		thresholdInt, _ := strconv.Atoi(guildData["offense-threshold"].(string))

		if userMap[entry.UserID] < thresholdInt {

			current := time.Now()
			entryTime, err := discordgo.SnowflakeTimestamp(entry.ID)
			if err != nil {
				return
			}

			if current.Sub(entryTime).Round(time.Second).Seconds() > 2 {
				return
			}

			if FindInSlice(auditMap, entry.ID) {
				continue
			}

			auditMap = append(auditMap, entry.ID)
			userMap[entry.UserID]++
		}

		if userMap[entry.UserID] >= thresholdInt {

			targetHighest := HighestRole(s, guildID, targetMember)
			selfHighest := HighestRole(s, guildID, selfMember)

			if targetHighest == nil || selfHighest == nil {
				return
			}

			if !IsAbove(selfHighest, targetHighest) || !HasPerms(s, nil, guildID, s.State.User.ID, discordgo.PermissionBanMembers) {
				return
			}

			err = HandleModeration(s, guildID, entry.UserID, fmt.Sprintf("%s | %s", s.State.User.Username, reason))
			if err != nil {
				return
			}

			LogChannel(s, guildID, fmt.Sprintf("<@%s> %s", entry.UserID, reason))
		}
	}
}

func RemoveFromSlice(slice []string, item string) []string {
	returnItems := []string{}
	for _, i := range slice {
		if i == item {
			continue
		}
		returnItems = append(returnItems, i)
	}
	return returnItems
}
