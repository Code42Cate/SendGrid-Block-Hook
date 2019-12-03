package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kelseyhightower/envconfig"
)

var (
	discordSession *discordgo.Session
	config         Config
)

// Config struct, filled with env variables from our Dockerfile
type Config struct {
	Interval         int    `default:"60"`
	DiscordToken     string `required:"true" split_words:"true"`
	DiscordChannelID string `required:"true" split_words:"true"`
	SendgridToken    string `required:"true" split_words:"true"`
	LastTimestamp    int    `default:"0" split_words:"true"`
}

// Block is an SendGrid block object as specified in:
// https://sendgrid.com/docs/API_Reference/Web_API_v3/blocks.html
type Block struct {
	Created int
	Email   string
	Reason  string
	Status  string
}

// Send a single block to your preferred service, in this case discord:D
func sendMessage(block Block) {
	message := fmt.Sprintf("Failed to send mail:\nCreated at: %d\nEmail: %s \nReason: %s\nStatus: %s\n", block.Created, block.Email, block.Reason, block.Status)
	if _, err := discordSession.ChannelMessageSend(config.DiscordChannelID, message); err != nil {
		fmt.Printf("Discord Error: %s\n", err.Error())
	}
}

// Iterate through blocks, send your log messages and save new latest timestamp
func checkBlocks(blocks []Block) {
	for _, block := range blocks {
		sendMessage(block)
		if block.Created > config.LastTimestamp {
			config.LastTimestamp = block.Created + 1 // + 1 or we would get the last one all the time
		}
	}
}

// Makes the sendgrid api requests, parses the blocks and calls checkBlocks
func getBlocks() {
	URL := "https://api.sendgrid.com/v3/suppression/blocks?start_time=" + strconv.Itoa(config.LastTimestamp)

	req, err := http.NewRequest("GET", URL, nil)
	req.Header.Set("Authorization", "Bearer "+config.SendgridToken)

	if err != nil {
		fmt.Printf("Failed to build Request: %s\n", err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		fmt.Printf("HTTP Call failed: %s\n", err.Error())
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var blocks []Block
	err = decoder.Decode(&blocks)
	if err != nil {
		fmt.Printf("Failed parsing json: %T\n%s\n%#v\n", err, err, err)
		return
	}
	checkBlocks(blocks)
}

// Read environment variables (from docker) into config struct. Fails if variables are missing.
// Does not check for empty variables, that is your responsibility
func parseConfig() error {

	prefix := ""
	if err := envconfig.Process(prefix, &config); err != nil {
		log.Fatal(err.Error())
	}

	if config.LastTimestamp == -1 {
		config.LastTimestamp = int(time.Now().Unix())
	}
	return nil
}

func main() {

	err := parseConfig()

	if err != nil {
		panic(err)
	}
	fmt.Println("Successfully created config")

	discordSession, err = discordgo.New("Bot " + config.DiscordToken)

	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully started Discord bot")
	discordSession.ChannelMessageSend(config.DiscordChannelID, "Sendgrid block thingy is online!")

	// Loop all da time!
	for range time.Tick(time.Duration(config.Interval) * time.Second) {
		getBlocks()
	}
}
