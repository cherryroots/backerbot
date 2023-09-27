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

The bot will give every role up to their donation amount

Basic setup of the csv currently is
|id |email          |donation|
|---|---------------|--------|
|1  |test1@gmail.com|5       |
|2  |test2@gmail.com|15      |
|3  |test3@gmail.com|30      |
|4  |test4@gmail.com|40      |
|5  |test5@gmail.com|50      |
