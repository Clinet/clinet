package main

import (
	"context"
	"strconv"
	"strings"

	"github.com/andygrunwald/go-trending"
	"github.com/bwmarrin/discordgo"
	"github.com/google/go-github/github"
)

func commandGitHub(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	switch args[0] {
	case "trend", "trends", "trending":
		if len(args) <= 1 {
			return NewErrorEmbed("GitHub Error", "Not enough arguments. Type ``cli$help "+env.Command+"`` for command usage.")
		}

		time := ""
		if len(args) >= 3 {
			switch args[2] {
			case "daily", "today":
				time = "daily"
			case "weekly", "week":
				time = "weekly"
			case "monthly", "month":
				time = "monthly"
			default:
				return NewErrorEmbed("GitHub Error", "Invalid trending time ``"+args[2]+"``. Type ``cli$help "+env.Command+"`` for command usage.")
			}
		}

		language := ""
		if len(args) >= 4 {
			language = args[3]
		}

		switch args[1] {
		case "repo", "repos", "repository", "repositories":
			projects, err := trending.NewTrending().GetProjects(time, language)
			if err != nil {
				return NewErrorEmbed("GitHub Error", "There was an error fetching the trending repositories.")
			}

			trendingEmbed := NewEmbed().
				SetTitle("GitHub Trending").
				SetDescription("These projects are currently trending.").
				SetColor(0x24292D)

			for i, project := range projects {
				details := "Project [" + project.RepositoryName + "](https://github.com/" + project.Owner + "/" + project.RepositoryName + ") created by [" + project.Owner + "](https://github.com/" + project.Owner + ")"
				if project.Language != "" {
					details += ", written in ``" + project.Language + "``."
				} else {
					details += "."
				}
				trendingEmbed.AddField("Project #"+strconv.Itoa(i+1)+": "+project.RepositoryName, details)
			}

			return trendingEmbed.MessageEmbed
		case "user", "users":
			developers, err := trending.NewTrending().GetDevelopers(time, language)
			if err != nil {
				return NewErrorEmbed("GitHub Error", "There was an error fetching the trending developers.")
			}

			trendingEmbed := NewEmbed().
				SetTitle("GitHub Trending").
				SetDescription("These developers are currently trending.").
				SetColor(0x24292D)

			for i, developer := range developers {
				trendingEmbed.AddField("Developer #"+strconv.Itoa(i+1)+": "+developer.DisplayName, "Developer "+developer.FullName+", AKA ["+developer.DisplayName+"](https://github.com/"+developer.DisplayName+").")
			}

			return trendingEmbed.MessageEmbed
		default:
			return NewErrorEmbed("GitHub Error", "Invalid trending type ``"+args[1]+"``. Type ``cli$help "+env.Command+"`` for command usage.")
		}
	default:
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
				SetColor(0x24292D)
			responseEmbed.Fields = fields
			responseEmbed.InlineAllFields()

			return responseEmbed.MessageEmbed
		case 2: //Repo was specified
			repo, err := GitHubFetchRepo(request[0], request[1])
			if err != nil {
				return NewErrorEmbed("GitHub Error", "There was an error fetching information about the specified repo.")
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
				SetColor(0x24292D)
			responseEmbed.Fields = fields
			responseEmbed.InlineAllFields()

			return responseEmbed.MessageEmbed
		}

		return NewErrorEmbed("GitHub Error", "Not enough arguments. Type ``cli$help "+env.Command+"`` for command usage.")
	}
}

func GitHubFetchUser(username string) (*github.User, error) {
	user, _, err := botData.BotClients.GitHub.Users.Get(context.Background(), username)
	if err != nil {
		return nil, err
	}
	return user, nil
}
func GitHubFetchRepo(owner string, repository string) (*github.Repository, error) {
	repo, _, err := botData.BotClients.GitHub.Repositories.Get(context.Background(), owner, repository)
	if err != nil {
		return nil, err
	}
	return repo, nil
}
