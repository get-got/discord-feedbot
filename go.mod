module github.com/get-got/discord-feedbot

go 1.19

require (
	github.com/bwmarrin/discordgo v0.27.1
	github.com/dustin/go-humanize v1.0.1
	github.com/fatih/color v1.14.1
	github.com/gtuk/discordwebhook v1.1.0
	github.com/hako/durafmt v0.0.0-20210608085754-5c1018a4e16b
	github.com/mmcdole/gofeed v1.2.1
	github.com/n0madic/twitter-scraper v0.0.0-20230520222908-ec6e8f3e190e
	gopkg.in/ini.v1 v1.67.0
	gorm.io/driver/sqlite v1.5.0
	gorm.io/gorm v1.25.0
)

require (
	github.com/PuerkitoBio/goquery v1.8.1 // indirect
	github.com/andybalholm/cascadia v1.3.2 // indirect
	github.com/gorilla/websocket v1.4.2 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.17 // indirect
	github.com/mattn/go-sqlite3 v1.14.16 // indirect
	github.com/mmcdole/goxpp v1.1.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/stretchr/testify v1.8.2 // indirect
	golang.org/x/crypto v0.0.0-20210921155107-089bfa567519 // indirect
	golang.org/x/net v0.9.0 // indirect
	golang.org/x/sys v0.7.0 // indirect
	golang.org/x/text v0.9.0 // indirect
)

replace github.com/n0madic/twitter-scraper => github.com/get-got/twitter-scraper v0.0.0-20230525183600
