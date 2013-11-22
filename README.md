## Run locally

```
go get github.com/JacobVorreuter/zendesk-trello-importer
cd $GOPATH/src/github.com/JacobVorreuter/zendesk-trello-importer/
go install
```

```
foreman start importer
```

```
foreman start web
```

## Deploy on Heroku

```
heroku create --buildpack https://github.com/kr/heroku-buildpack-go.git
```

```
heroku config:add \
  ZENDESK_TOKEN=... \
  ZENDESK_URL=https://support.heroku.com \
  ZENDESK_USERNAME=bob@myorg.com \
  ZENDESK_GROUP_ID=... # identifies the zendesk group for which to import tickets \
  ZENDESK_VIEW=... \
  ZENDESK_DUMMY_ACCT=dummy@myorg.com # the zendesk account email for the dummy account responsible for reassigning tickets to your support rotation
```

Use the link below to generate your Trello API token.  Make sure to replace `YOURKEYGOESHERE` with your Trello API key:

[Grant access to Go Zendesk Importer application](https://trello.com/1/authorize?key=YOURKEYGOESHERE&name=Go%20Zendesk%20Importer&expiration=never&response_type=token&scope=read,write)

```
heroku config:add \
  TRELLO_API_KEY=... \
  TRELLO_API_TOKEN=... \
  TRELLO_BOARD_ID=... \
  TRELLO_LIST="Support Tickets"
```

```
heroku config:add \
  PAGERDUTY_URL=... \
  PAGERDUTY_API_KEY=... \
  PAGERDUTY_ROTATION_IDS=... # comma delimited list of on-call schedule IDs
```

```
git push heroku master
```

Configure the importer to run regularly with the Heroku scheduler
