package main

/*
	The below initializer functions initializes a set of data within various maps used throughout Clinet.
	They perform checks to ensure that the data does not yet exist as to prevent from overwriting pre-existing data.

	These initializer functions were created to greatly reduce repetitive coding practices within various functions in Clinet.
*/

func initializeGuildData(guildID string) {
	_, guildDataExists := guildData[guildID]
	if !guildDataExists {
		guildData[guildID] = &GuildData{}
		guildData[guildID].Queries = make(map[string]*Query)
	}
}

func initializeGuildSettings(guildID string) {
	_, guildSettingsExists := guildSettings[guildID]
	if !guildSettingsExists {
		guildSettings[guildID] = &GuildSettings{}
	}
}

func initializeUserSettings(userID string) {
	_, userSettingsExists := userSettings[userID]
	if !userSettingsExists {
		userSettings[userID] = &UserSettings{}
	}
}

func initializeStarboard(guildID string) {
	_, starboardExists := starboards[guildID]
	if !starboardExists {
		starboards[guildID] = &Starboard{}
		starboards[guildID].Emoji = "⭐"
		starboards[guildID].NSFWEmoji = "💦"
		starboards[guildID].AllowSelfStar = false
		starboards[guildID].MinimumStars = 2
	}
}
