package main

import (
	"encoding/json"
	"math"
	"math/rand"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

//returns true if it's time to notify the user
func cyclePatreonNotice(userID string) bool {
	if MemberIsPatron(botData.DiscordSession, userID) {
		return false
	}

	if timeSince, exists := patreonNotices[userID]; exists {
		if time.Now().Sub(timeSince).Hours() < 24 {
			return false
		}
	}

	patreonNotices[userID] = time.Now()
	return true
}

func MemberIsPatron(s *discordgo.Session, userID string) bool {
	if hasRole, err := MemberHasRole(s, patreonBotData.BotOptions.RoleRequired.GuildID, userID, patreonBotData.BotOptions.RoleRequired.RoleID); err == nil {
		return hasRole
	}
	return false
}

func MemberHasRole(s *discordgo.Session, guildID, userID, roleID string) (bool, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	for _, memberRoleID := range member.Roles {
		if roleID == memberRoleID {
			return true, nil
		}
	}

	return false, nil
}

// MemberHasPermission checks if a member has the given permission
// for example, If you would like to check if user has the administrator
// permission you would use
// --- MemberHasPermission(s, guildID, userID, discordgo.PermissionAdministrator)
// If you want to check for multiple permissions you would use the bitwise OR
// operator to pack more bits in. (e.g): PermissionAdministrator|PermissionAddReactions
// =================================================================================
//     s          :  discordgo session
//     guildID    :  guildID of the member you wish to check the roles of
//     userID     :  userID of the member you wish to retrieve
//     channelID  :  channelID of the member who sent the message
//     permission :  the permission you wish to check for
func MemberHasPermission(s *discordgo.Session, guildID string, userID string, channelID string, permission int64) (bool, error) {
	member, err := s.State.Member(guildID, userID)
	if err != nil {
		if member, err = s.GuildMember(guildID, userID); err != nil {
			return false, err
		}
	}

	channel, err := s.State.Channel(channelID)
	if err != nil {
		if channel, err = s.Channel(channelID); err != nil {
			return false, err
		}
	}

	for _, roleID := range member.Roles {
		role, err := s.State.Role(guildID, roleID)
		if err != nil {
			return false, err
		}
		if role.Permissions&discordgo.PermissionAdministrator != 0 {
			return true, nil
		}
		if role.Permissions&permission != 0 {
			return true, nil
		}

		for _, permissionOverwrite := range channel.PermissionOverwrites {
			if permissionOverwrite.Type == discordgo.PermissionOverwriteTypeRole || permissionOverwrite.ID == roleID {
				if permissionOverwrite.Allow&permission != 0 {
					return true, nil
				}
				if permissionOverwrite.Deny&permission != 0 {
					return false, nil
				}
			}
		}
	}

	for _, permissionOverwrite := range channel.PermissionOverwrites {
		if permissionOverwrite.Type == discordgo.PermissionOverwriteTypeMember || permissionOverwrite.ID == userID {
			if permissionOverwrite.Allow&permission != 0 {
				return true, nil
			}
			if permissionOverwrite.Deny&permission != 0 {
				return false, nil
			}
		}
	}

	return false, nil
}

// CreationTime returns the creation time of a Snowflake ID relative to the creation of Discord.
// Taken from https://github.com/Moonlington/FloSelfbot/blob/master/commands/commandutils.go#L117
func CreationTime(ID string) (t time.Time, err error) {
	i, err := strconv.ParseInt(ID, 10, 64)
	if err != nil {
		return
	}
	timestamp := (i >> 22) + 1420070400000
	t = time.Unix(timestamp/1000, 0)
	return
}

type CaseInsensitiveReplacer struct {
	toReplace   *regexp.Regexp
	replaceWith string
}

func NewCaseInsensitiveReplacer(toReplace, with string) *CaseInsensitiveReplacer {
	return &CaseInsensitiveReplacer{
		toReplace:   regexp.MustCompile("(?i)" + toReplace),
		replaceWith: with,
	}
}
func (cir *CaseInsensitiveReplacer) Replace(str string) string {
	return cir.toReplace.ReplaceAllString(str, cir.replaceWith)
}

func zeroPad(str string) (result string) {
	if len(str) < 2 {
		result = "0" + str
	} else {
		result = str
	}
	return
}

func secondsToHuman(input float64) (result string) {
	hours := math.Floor(float64(input) / 60 / 60)
	seconds := int(input) % (60 * 60)
	minutes := math.Floor(float64(seconds) / 60)
	seconds = int(input) % 60

	if hours > 0 {
		result = strconv.Itoa(int(hours)) + ":" + zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	} else {
		result = zeroPad(strconv.Itoa(int(minutes))) + ":" + zeroPad(strconv.Itoa(int(seconds)))
	}

	return
}

func roundTime(d, r time.Duration) time.Duration {
	if r <= 0 {
		return d
	}
	neg := d < 0
	if neg {
		d = -d
	}
	if m := d % r; m+m < r {
		d = d - m
	} else {
		d = d + r - m
	}
	if neg {
		return -d
	}
	return d
}

func remove(s []string, r string) []string {
	for i, v := range s {
		if v == r {
			return append(s[:i], s[i+1:]...)
		}
	}
	return s
}

// GetStringInBetween returns empty string if no start string found
func GetStringInBetween(str string, start string, end string) (result string) {
	s := strings.Index(str, start)
	if s == -1 {
		return
	}
	s += len(start)
	str = str[s:]
	e := strings.Index(str, end)
	return str[:e]
}

// Gets the JSON from the API and assigns the data to the target
func unmarshal(body *http.Response, target interface{}) error {
	defer body.Body.Close()
	return json.NewDecoder(body.Body).Decode(target)
}

func randomInRange(begin, end int) int {
	randomizer := rand.New(rand.NewSource(time.Now().UnixNano()))
	return randomizer.Intn(end-begin) + begin
}
