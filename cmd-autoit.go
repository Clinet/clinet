package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/JoshuaDoes/autogo/autoit"
)

func commandAutoIt(args []string, env *CommandEnvironment) *discordgo.MessageEmbed {
	script := ";DISCORD\n" + strings.Join(args, " ") + "\n"
	//fmt.Println([]byte(script))
	//return NewGenericEmbed("test", script)
	
	vm, err := autoit.NewAutoItScriptVM("discord.au3", []byte(script), nil)
	if err != nil {
		return NewErrorEmbed("AutoIt Script Error", "%v", err)
	}
	
	timeStart := time.Now()
	err = vm.Run()
	if err != nil {
		return NewErrorEmbed("AutoIt Runtime Error", "%v", err)
	}
	timeEnd := time.Now()
	runtime := timeEnd.Sub(timeStart)
	
	output := "Script execution succeeded."
	if vm.Stdout() != "" {
		output += "\nStdout:\n```\n" + vm.Stdout() + "\n```"
	}
	if vm.Stderr() != "" {
		output += "\nStderr:\n```\n" + vm.Stderr() + "\n```"
	}
	
	return NewGenericEmbed(fmt.Sprintf("AutoIt VM (%.4fms)", float64(runtime.Nanoseconds()) / 1000000), output)
}