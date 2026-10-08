module github.com/XanderOnGithub/home-tools

go 1.27.1

// The frontend lives in web/. Ignoring it keeps `go ./...` from walking
// node_modules (some npm packages ship stray .go files that break builds).
ignore ./web
