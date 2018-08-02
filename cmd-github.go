package main

import (
	"strconv"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func commandGitHub(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	request := strings.Split(args[0], "/")

	switch len(request) {
	case 1: //Only user was specified
		user, err := GitHubFetchUser(request[0])
		if err != nil {
			return NewErrorEmbed("GitHub Error", "There was an error fetching information about the specified user.")
		}

		fields := []*discordgo.MessageEmbedField{}

		//Gather user info
		if user.Bio != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Bio", Value: *user.Bio})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Username", Value: *user.Login})
		if user.Name != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *user.Name})
		}
		if user.Company != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Company", Value: *user.Company})
		}
		if *user.Blog != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Blog", Value: *user.Blog})
		}
		if user.Location != nil {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Location", Value: *user.Location})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Repos", Value: strconv.Itoa(*user.PublicRepos)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Public Gists", Value: strconv.Itoa(*user.PublicGists)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Following", Value: strconv.Itoa(*user.Following)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Followers", Value: strconv.Itoa(*user.Followers)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *user.HTMLURL})

		//Build embed about user
		responseEmbed := NewEmbed().
			SetTitle("GitHub User: " + *user.Login).
			SetImage(*user.AvatarURL).
			SetColor(0x24292D).MessageEmbed
		responseEmbed.Fields = fields

		return responseEmbed
	case 2: //Repo was specified
		repo, err := GitHubFetchRepo(request[0], request[1])
		if err != nil {
			return NewErrorEmbed("GitHub Error", "There was an error fetchign information about the specified repo.")
		}

		fields := []*discordgo.MessageEmbedField{}

		//Gather repo info
		if repo.Description != nil && *repo.Description != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Description", Value: *repo.Description})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Name", Value: *repo.FullName})
		if repo.Homepage != nil && *repo.Homepage != "" {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Homepage", Value: *repo.Homepage})
		}
		if len(repo.Topics) > 0 {
			fields = append(fields, &discordgo.MessageEmbedField{Name: "Topics", Value: strings.Join(repo.Topics, ", ")})
		}
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Default Branch", Value: *repo.DefaultBranch})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Is Fork", Value: strconv.FormatBool(*repo.Fork)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Forks", Value: strconv.Itoa(*repo.ForksCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Networks", Value: strconv.Itoa(*repo.NetworkCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Open Issues", Value: strconv.Itoa(*repo.OpenIssuesCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Stargazers", Value: strconv.Itoa(*repo.StargazersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Subscribers", Value: strconv.Itoa(*repo.SubscribersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Watchers", Value: strconv.Itoa(*repo.WatchersCount)})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "GitHub URL", Value: *repo.HTMLURL})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Clone URL", Value: *repo.CloneURL})
		fields = append(fields, &discordgo.MessageEmbedField{Name: "Git URL", Value: *repo.GitURL})

		//Build embed about repo
		responseEmbed := NewEmbed().
			SetTitle("GitHub Repo: " + *repo.FullName).
			SetColor(0x24292D).MessageEmbed
		responseEmbed.Fields = fields

		return responseEmbed
	}

	return NewErrorEmbed("GitHub Error", "You must specify a GitHub user or a GitHub repo to fetch info about.\n\nExamples:\n```"+botData.CommandPrefix+"github JoshuaDoes\n"+botData.CommandPrefix+"gh JoshuaDoes/clinet-discord```")
}
