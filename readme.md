# Cinemanager bot

## Running the bot

The bot requires the following environment variables to function:

### Environment Variables

| Variable            | Description                                                                                    
|---------------------|-------------------------------------------------------------------------------------------------|
| `TELEGRAM_BOT_TOKEN`| The token used to authenticate the bot with the Telegram API. This token is provided by the BotFather when you create a new bot on Telegram. | 
| `OMDB_API_KEY`      | The API key used to authenticate requests to the OMDB API. This key is required to fetch movie data from the OMDB API. | 
| `DEBUG`             | A flag to enable or disable debug mode for the bot. When set to `true`, the bot will log additional debug information. | 


# Development environment

## Watch mode

To get live updates on code changes during development, use [wgo](https://github.com/bokwoon95/wgo)

```shell
go install github.com/bokwoon95/wgo@latest
```

```shell
wgo run ./main.go
```

## Database

You need a connection to a postgres database

