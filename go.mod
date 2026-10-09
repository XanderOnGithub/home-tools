module github.com/XanderOnGithub/home-tools

go 1.27.1

// The frontend lives in web/. Ignoring it keeps `go ./...` from walking
// node_modules (some npm packages ship stray .go files that break builds).
ignore ./web

require github.com/bwmarrin/discordgo v0.29.0

require (
	github.com/gorilla/websocket v1.5.3 // indirect
	golang.org/x/crypto v0.58.0 // indirect
	golang.org/x/sys v0.49.0 // indirect
)
