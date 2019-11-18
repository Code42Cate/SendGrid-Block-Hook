## SendGrid Blocked Mails Script Thingy

It literally just makes requests to the SendGrid API in your preferred interval and if it gets a new blocked mail it can send a message somewhere. I only implemented Discord for now. Adding another service for sending the log messages is easy

More README coming later:D

Your config.json file should have this format:
```json
{
    "Interval": 60,
    "LastTimestamp": 0,
    "DiscordToken": "your-bot-token",
    "DiscordChannelID": "your-logging-channel",
    "SendGridToken": "your-sendgrid-token-with-supress-access"
}
```