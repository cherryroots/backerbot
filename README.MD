This project requires two external go modules

discordgo
```sh
go get github.com/bwmarrin/discordgo
```

skv
```sh
go get github.com/rapidloop/skv
```

Rename the `sample.env` file to `.env` then populate the token field with your discord bot token.

The bot requires
- All the intents set
- `bot` and `application.commands` scope
- At least `Manage Roles`,  `Read Messages/View Channels`, `Send messages`, `Read Message History` and `Use Slash Commands` permissions

The bot doesn't have configurable roles yet, they can be changed in the `utility.go` file
First field is just for readability, second field is the role id and the third field is what's in the tier field in the csv.

Basic setup of the csv currently is
|id |email          |tier  |
|---|---------------|------|
|1  |test1@gmail.com|Tier 1|
|2  |test2@gmail.com|Tier 2|
|3  |test3@gmail.com|Tier 1|
|4  |test4@gmail.com|Tier 2|
|5  |test5@gmail.com|Tier 1|