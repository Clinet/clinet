package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
)

func recoverPanic() {
	if panicReason := recover(); panicReason != nil {
		fmt.Println("Clinet has encountered an unrecoverable error and has crashed.")
		fmt.Println("Some information describing this crash: " + panicReason.(error).Error())
		if botData.SendOwnerStackTraces || configIsBot == "false" {
			stack := make([]byte, 65536)
			l := runtime.Stack(stack, true)
			fmt.Println("Stack trace:\n" + string(stack[:l]))
			err := ioutil.WriteFile("stacktrace.txt", stack[:l], 0644)
			if err != nil {
				fmt.Println("Failed to write stack trace.")
			}
			err = ioutil.WriteFile("crash.txt", []byte(panicReason.(error).Error()), 0644)
			if err != nil {
				fmt.Println("Failed to write crash error.")
			}
		}
		os.Exit(1)
	}
}

func checkPanicRecovery() {
	ownerPrivChannel, err := botData.DiscordSession.UserChannelCreate(botData.BotOwnerID)
	if err != nil {
		debugLog("An error occurred creating a private channel with the bot owner.", false)
	} else {
		ownerPrivChannelID := ownerPrivChannel.ID

		crash, crashErr := ioutil.ReadFile("crash.txt")
		stack, stackErr := os.Open("stacktrace.txt")

		if crashErr == nil && stackErr == nil {
			DowntimeReason = "Crash: " + string(crash)

			botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Clinet has just recovered from an error that caused a crash.")
			botData.DiscordSession.ChannelMessageSend(ownerPrivChannelID, "Crash:\n```"+string(crash)+"```")
			botData.DiscordSession.ChannelFileSendWithMessage(ownerPrivChannelID, "Stack trace:", "stacktrace.txt", stack)
		}

		stack.Close()
		os.Remove("crash.txt")
		os.Remove("stacktrace.txt")
	}
}
