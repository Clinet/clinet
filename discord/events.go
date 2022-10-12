package discord

import (
	"fmt"

	"github.com/bwmarrin/discordgo"
	"github.com/Clinet/clinet/cmds"
	"github.com/Clinet/discordgo-embed"
)

func discordReady(session *discordgo.Session, event *discordgo.Ready) {
	Log.Trace("--- discordReady(", event, ") ---")
	for Discord == nil {
		//Wait for Discord to finish connecting on our end
		if Discord != nil {
			break
		}
	}
	Discord.User = event.User
	Log.Info("Logged into Discord as ", Discord.User, "!")
}

func discordMessageCreate(session *discordgo.Session, event *discordgo.MessageCreate) {
	Log.Trace("--- discordMessageCreate(", event, ") ---", event.ID, event.GuildID, event.ChannelID, event.Member, event.Member.User)
	message, err := session.ChannelMessage(event.ChannelID, event.ID)
	if err != nil {
		Log.Error(message, err)
	}

	cmdResps, err := convoHandler(session, message)
	if err != nil {
		Log.Error(err, cmdResps)
		return
	}

	for i := 0; i < len(cmdResps); i++ {
		if cmdResps[i] == nil || cmdResps[i].Text == "" {
			continue
		}

		cmdResps[i].OnReady(func(r *cmds.CmdResp) {
			Log.Trace("Response to message for convo: " + r.String())
			if r.Title != "" || r.Color != nil || r.Image != "" {
				respEmbed := embed.NewEmbed().
					SetDescription(r.Text)
					if r.Title != "" {
					respEmbed.SetTitle(r.Title)
				}

				if r.Color != nil {
					respEmbed.SetColor(*r.Color)
				}
				if r.Image != "" {
					respEmbed.SetImage(r.Image)
				}

				_, err := session.ChannelMessageSendComplex(event.ChannelID, &discordgo.MessageSend{
					Embed: respEmbed.MessageEmbed,
				})
				if err != nil {
					Log.Error(err)
				}
			} else {
				_, err := session.ChannelMessageSend(event.ChannelID, r.Text)
				if err != nil {
					Log.Error(err)
				}
			}
		})
	}
}

func discordInteractionCreate(session *discordgo.Session, event *discordgo.InteractionCreate) {
	Log.Trace("--- discordInteractionCreate(", event, ") ---", event.ID, event.Type, event.GuildID, event.ChannelID, event.Member, event.User, event.Token, event.Version)

	switch event.Type {
	case discordgo.InteractionApplicationCommand:
		eventData := event.ApplicationCommandData()
		eventOpts := eventData.Options

		cmd := cmds.GetCmd(eventData.Name)
		if cmd == nil {
			Log.Error("Unable to find command " + eventData.Name)
			return
		}

		cmdAlias, cmdResps := cmdHandler(cmd, eventData.Name, eventOpts)
		for i := 0; i < len(cmdResps); i++ {
			if cmdResps[i] == nil || cmdResps[i].Text == "" {
				continue
			}

			cmdResps[i].OnReady(func(r *cmds.CmdResp) {
				Log.Trace("Response to interaction for cmd " + cmdAlias + ": " + r.String())
				resp := &discordgo.InteractionResponse{
					Type: discordgo.InteractionResponseChannelMessageWithSource,
					Data: &discordgo.InteractionResponseData{},
				}

				if r.Title != "" || r.Color != nil || r.Image != "" {
					respEmbed := embed.NewEmbed().
						SetDescription(r.Text)

					if r.Title != "" {
						respEmbed.SetTitle(r.Title)
					}
					if r.Color != nil {
						respEmbed.SetColor(*r.Color)
					}
					if r.Image != "" {
						respEmbed.SetImage(r.Image)
					}

					resp.Data.Embeds = []*discordgo.MessageEmbed{respEmbed.MessageEmbed}
				} else {
					resp.Data.Content = r.Text
				}

				session.InteractionRespond(event.Interaction, resp)
			})
		}
	case discordgo.InteractionMessageComponent:
		msgData := event.MessageComponentData()
		if msgData.ComponentType != discordgo.ButtonComponent {
			return
		}

		switch msgData.CustomID {
		case "example1":
			//do something when this button is pressed
		}

		respEmbed := embed.NewGenericEmbed("Feature Name", "Example response embed from feature")
		featureComponents := []discordgo.MessageComponent{
			discordgo.ActionsRow{
				Components: []discordgo.MessageComponent{
					discordgo.Button{
						Label: "Example 1",
						Style: discordgo.SuccessButton,
						CustomID: "example1",
					},
				},
			},
		}
		resp := &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseUpdateMessage,
			Data: &discordgo.InteractionResponseData{
				CustomID: "featureName",
				Embeds: []*discordgo.MessageEmbed{respEmbed},
				Components: featureComponents,
				Flags: discordgo.MessageFlagsEphemeral,
			},
		}

		err := session.InteractionRespond(event.Interaction, resp)
		if err != nil {
			Log.Error(err)
		} else {
			Log.Debug("Responded to button ", msgData.CustomID, " with response: ", fmt.Sprintf("%v", respEmbed))
		}
	}
}
