module backerbot

go 1.21.0

require (
	github.com/boltdb/bolt v1.3.1
	github.com/bwmarrin/discordgo v0.27.1
	github.com/joho/godotenv v1.5.1
)

replace github.com/rapidloop/skv => ../skv

require (
	github.com/gorilla/websocket v1.5.0 // indirect
	golang.org/x/crypto v0.13.0 // indirect
	golang.org/x/sys v0.12.0 // indirect
)
