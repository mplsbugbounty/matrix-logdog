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
export MATRIX_LOGDOG_URL="https://matrix.example.com"
export MATRIX_LOGDOG_USER="@yourbot:example.com"
export MATRIX_LOGDOG_TOKEN="yourbots-token"
export MATRIX_LOGDOG_ROOM="!WAoLCYOOyceAxMaFYU:example.com"
#this next one is the directory in which files are monitored for changes
export MATRIX_LOGDOG_WATCH_DIR="/home/exampleuser/logz/"
#this file is expected to be a newline-delimited set of strings to search for
export MATRIX_LOGDOG_MATCH_FILE="/home/exampleuser/logrulez.txt"
```

then, start the bot:

```
go run main.go
```
