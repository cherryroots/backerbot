## Kickstarter backer bot
Rename the `sample.env` file to `.env` then populate the token field with your discord bot token.

The bot requires
- All the intents set
- `bot` and `application.commands` scope
- At least `Manage Roles`,  `Read Messages/View Channels`, `Send messages`, `Read Message History` and `Use Slash Commands` permissions

The bot will give every role up to their donation amount

Basic setup of the csv currently is as follows
The pledge is cleaned up
€100.00 -> 100.00
$1,000.00 -> 1000.00

| Email            | Reward Title   | Pledge Amount | Pledge Status |
|------------------|----------------|---------------|---------------|
| email1@gmail.com | Some Reward    | €100.0        | collected     |
| email2@gmail.com | Another reward | €15.00        | errored       |
| email3@gmail.com | A cool reward  | €50.00        | collected     |
