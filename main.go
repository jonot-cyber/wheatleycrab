package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

var stopChannel chan bool

func download(s string, c chan string) {
	fmt.Println("Downloading:", s)
	downloadPath := "audio/" + strings.ReplaceAll(s, "/", "SLASH") + ".opus"
	go func() {
		cmd := exec.Command("youtube-dl", "-x", "--default-search", "ytsearch", "--audio-format", "opus", s, "-o", downloadPath)
		err := cmd.Run()
		if err != nil {
			return
		}
		fmt.Println("Download of " + s + " complete!")
		os.Rename(downloadPath, downloadPath+".part")
	}()
	time.Sleep(time.Second * 2)
	c <- downloadPath + ".part"
}

var commands = []*discordgo.ApplicationCommand{
	{
		Name:        "play",
		Description: "Play a song",
		Options: []*discordgo.ApplicationCommandOption{
			{
				Type:        discordgo.ApplicationCommandOptionString,
				Name:        "song",
				Description: "The song to be played",
				Required:    true,
			},
		},
	},
	{
		Name:        "skip",
		Description: "Skip this song",
	},
}

var handlers = map[string]func(s *discordgo.Session, i *discordgo.InteractionCreate, c chan string){
	"play": func(s *discordgo.Session, i *discordgo.InteractionCreate, c chan string) {
		songName := i.ApplicationCommandData().Options[0].StringValue()

		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "**Downloading and playing song: **" + songName,
			},
		})
		go download(songName, c)
	},
	"skip": func(s *discordgo.Session, i *discordgo.InteractionCreate, c chan string) {
		fmt.Println("SKIPPING")
		s.InteractionRespond(i.Interaction, &discordgo.InteractionResponse{
			Type: discordgo.InteractionResponseChannelMessageWithSource,
			Data: &discordgo.InteractionResponseData{
				Content: "**SKIPPED**",
			},
		})
		stopChannel <- true
	},
}

func main() {
	token := os.Getenv("DISCORD_TOKEN")

	var server = flag.String("server", "811598800606461962", "The server to use")
	var channel = flag.String("channel", "811598800606461968", "The channel to use")

	flag.Parse()

	c := make(chan string)
	stopChannel := make(chan bool)

	discord, err := discordgo.New("Bot " + token)
	if err != nil {
		fmt.Println(err)
		return
	}

	discord.AddHandler(func(s *discordgo.Session, i *discordgo.InteractionCreate) {
		if h, ok := handlers[i.ApplicationCommandData().Name]; ok {
			h(s, i, c)
		}
	})

	err = discord.Open()
	if err != nil {
		fmt.Println(err)
		return
	}
	defer discord.Close()

	for _, v := range commands {
		_, err := discord.ApplicationCommandCreate(discord.State.User.ID, *server, v)
		if err != nil {
			fmt.Println(err)
			return
		}

	}

	discord.UpdateGameStatus(0, "some tunez! 🎵")

	for song := range c {
		dgv, err := discord.ChannelVoiceJoin(*server, *channel, false, true)
		if err != nil {
			fmt.Println(err)
			return
		}
		dgvoice.PlayAudioFile(dgv, song, stopChannel)
		dgv.Disconnect()
		dgv.Close()
		os.Remove(song)
	}
}
