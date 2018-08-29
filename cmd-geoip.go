package main

import (
	"strconv"

	"github.com/JoshuaDoes/goeip"
	"github.com/bwmarrin/discordgo"
)

func commandGeoIP(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	data, err := goeip.Lookup(args[0])
	if err != nil {
		return NewErrorEmbed("GeoIP Error", "There was an error with the GeoIP utility.")
	}
	if data.Error > 0 {
		return NewErrorEmbed("GeoIP Error", data.Details)
	}

	geoipEmbed := NewEmbed().
		SetTitle("GeoIP Lookup for "+args[0]).
		SetDescription(data.Details).
		AddField("IP Address", data.IPAddr).
		AddField("Hostname", data.Hostname).
		SetColor(0x1C1C1C)

	if data.City != "" {
		geoipEmbed.AddField("City", data.City)
	}
	if data.State != "" {
		geoipEmbed.AddField("State", data.State)
	}
	if data.PostalCode != "" {
		geoipEmbed.AddField("Postal Code", data.PostalCode)
	}

	geoipEmbed.AddField("Country", data.Country.Name+" ("+data.Country.Code+")")

	if data.Timezone != "" {
		geoipEmbed.AddField("Timezone", data.Timezone)
	}

	geoipEmbed.AddField("Location", "Latitude: "+strconv.FormatFloat(data.Location.Latitude, 'f', -1, 64)+"\nLongitude: "+strconv.FormatFloat(data.Location.Longitude, 'f', -1, 64))
	geoipEmbed.AddField("ASN", "Number: "+strconv.Itoa(data.ASN.Number)+"\nOrganization: "+data.ASN.Organization)
	geoipEmbed.InlineAllFields()

	return geoipEmbed.MessageEmbed
}
