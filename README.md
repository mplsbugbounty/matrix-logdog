Based on https://git.cyberia.club/cyberia/jackal

Watch a directory for filtered changes, receive notifications in Matrix!
Any time a file is written within the selected directory, the new file data is checked against a provided list
of terms. If a match is found, the matched line will be sent as a message to thee provided matrix room. 

Matrix encryption is not yet supported.

## usage

first, set up proper variables for MATRIX_LOGDOG:

```
# use whatever user you want! can even be from different homeserver
# token can be found in element at settings -> help -> advanced -> token
# room IDs can be found under settings -> advanced -> internal room id

export MATRIX_LOGDOG_CONFIG_FILE="myconfig.json"
```
Then, myconfig.json should look like:
```
{
    "MatrixHomeserver": "https://matrix.example.com",
    "MatrixUser": "user@example.com",
    "MatrixRoom": "!WAoLCYOOyceAxMaFYU:example.com",
    "MatrixPassword": "uSeRlOGinPassWurd1",
    "WatchDir": "/home/exampleuser/logz/",
    "Filters": [ "msg:","message:" ],
    "SQLiteDatabase": "test.db"
}
```


then, start the bot:

```
go run main.go
```
