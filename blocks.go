package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/amoghe/distillog"
	"github.com/bwmarrin/discordgo"
)

var (
	logger         distillog.Logger
	discordSession *discordgo.Session
	config         *Config
)

// CONFIGPATH = relative path to config file, see Config struct
const CONFIGPATH = "./config.json"

// Config struct, filled with config.json data
type Config struct {
	Interval         int
	LastTimestamp    int
	DiscordToken     string
	DiscordChannelID string
	SendGridToken    string
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
	message := fmt.Sprintf("Failed to sent mail:\nCreated at: %d\nEmail: %s \nReason: %s\nStatus: %s\n", block.Created, block.Email, block.Reason, block.Status)
	_, err := discordSession.ChannelMessageSend(config.DiscordChannelID, message)
	if err != nil {
		logger.Errorf("Discord Error: %s", err.Error())
	} else {
		logger.Infof("Successfully send message: %s", strings.ReplaceAll(message, "\n", ";"))
	}

}

// Iterate through blocks, send your log messages and save new latest timestamp
func checkBlocks(blocks []Block) {
	for _, block := range blocks {
		sendMessage(block)
		if block.Created > config.LastTimestamp {
			config.LastTimestamp = block.Created + 1 // + 1 or we would get the last one all the time
			err := saveLastTimestamp(config.LastTimestamp)
			if err != nil {
				logger.Errorf("Failed to save last timestamp: %s", err.Error())
			}
		}
	}
}

// Makes the sendgrid api requests, parses the blocks and calls checkBlocks
func getBlocks() {
	URL := "https://api.sendgrid.com/v3/suppression/blocks?start_time=" + strconv.Itoa(config.LastTimestamp)

	req, err := http.NewRequest("GET", URL, nil)
	req.Header.Set("Authorization", "Bearer "+config.SendGridToken)

	if err != nil {
		logger.Errorf("Failed to build Request: %s", err)
		return
	}

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		logger.Errorf("HTTP Call failed: %s", err.Error())
		return
	}
	defer res.Body.Close()

	decoder := json.NewDecoder(res.Body)
	var blocks []Block
	err = decoder.Decode(&blocks)
	if err != nil {
		logger.Errorf("Failed parsing json: %T\n%s\n%#v\n", err, err, err)
	} else {
		checkBlocks(blocks)
	}
}

// Saves the timestamp as LastTimestamp field in the config file
func saveLastTimestamp(timestamp int) error {
	config.LastTimestamp = timestamp
	file, err := os.Create(CONFIGPATH)
	if err != nil {
		return err
	}
	json, err := json.MarshalIndent(config, "", "    ")
	if err != nil {
		return err
	}
	file.Write(json)
	return nil
}

// Function to parse config file into global config struct
func parseConfig() error {
	file, err := os.Open(CONFIGPATH)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		return err
	}
	return nil
}

func main() {

	lumberjackHandle := &lumberjack.Logger{
		Filename:   "./log.txt",
		MaxSize:    500, // megabytes
		MaxBackups: 3,
		MaxAge:     30, // days
	}

	logger = distillog.NewStreamLogger("tag", lumberjackHandle)
	distillog.SetOutput(lumberjackHandle)
	logger.Infoln("Starting thingy")
	err := parseConfig()
	if err != nil {
		logger.Errorf("Failed to parse config: %s", err.Error())
		panic(err)
	}
	logger.Infof("Successfully parsed config file: %s", CONFIGPATH)

	discordSession, err = discordgo.New("Bot " + config.DiscordToken)
	if err != nil {
		logger.Errorf("Failed to create new Discord bot: %s", err.Error())
		panic(err)
	}
	logger.Infoln("Successfully started Discord bot")

	// Loop all da time!
	for range time.Tick(time.Duration(config.Interval) * time.Second) {
		getBlocks()
	}
}
