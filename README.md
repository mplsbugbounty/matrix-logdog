# 🐶 jackal

a humble doggy roboto

## what is jackal?

jackal is a watchdog - it watches things & barks
to alert people of events.

## usage

first, set up proper variables for jackal:

```
# use whatever user you want! can even be from different homeserver
# token can be found in element at settings -> help -> advanced -> token
# room IDs can be found under settings -> advanced -> internal room id
export JACKAL_MATRIX_URL="https://matrix.example.com"
export JACKAL_MATRIX_USER="@yourbot:example.com"
export JACKAL_MATRIX_TOKEN="yourbots-token"
export JACKAL_MATRIX_ROOM="!WAoLCYOOyceAxMaFYU:example.com"
export JACKAL_PROMETHEUS_URL="https://prometheus.example.com"
```

then, start the bot:

```
go run main.go
```
