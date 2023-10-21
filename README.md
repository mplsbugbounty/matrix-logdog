Based on https://git.cyberia.club/cyberia/jackal

## Plattorm
Linux!

## Description
Watch a directory for filtered changes, receive notifications in Matrix!
Any time a file is written within the selected directory, the new file data is checked against a provided list
of terms. If a match is found, the matched line will be sent as a message to the provided matrix room. 
You may either provide the filter terms as a list in the json configuration file, or
specify a wordlist with the -filterFile flag so that each line will be a search term.

## Installation

First, you need to have Go installed. Installation instructions can be found here: https://go.dev/doc/install
So far, the only tested version of Go is 1.21.3.

libolm-dev is required, and right now it's a manual install process, sorry about that. You should be able to install it with:
```
sudo apt install libolm-dev
```

With those dependencies met, you just clone the repo and enter it:
```
git clone https://github.com/mplsbugbounty/matrix-loggo-doggo.git
cd matrix-loggo-doggo
```

## Usage


Use whatever user you want!
Room IDs can be found under settings -> advanced -> internal room id.

First, you need to configure account info for your matrix bot account.
Configuration is specified in matrix_logdog.json should look like:
```
{
    "MatrixHomeserver": "https://matrix.example.com",
    "MatrixUser": "username_localpart",
    "MatrixRoom": "!WAoLCYOOyceAxMaFYU:example.com",
    "MatrixPassword": "uSeRlOGinPassWurd1",
    "WatchDir": "/home/exampleuser/logz/",
    "Filters": [ "msg:","message:" ],
    "SQLiteDatabase": "test.db"
}
```

You may optionally specifiy a different config file by setting this environment variable, MATRIX_LOGDOG_CONFIG in which case e.g. myconfig.json should look like the above. Use like this:
```
export MATRIX_LOGDOG_CONFIG_FILE="myconfig.json"
```

Once you have finished configuration, start the bot with:
```
go run main.go
```
You may use a wordlist (a newline delimited text file in which each line is presumed 
to be an entry) to specify filters, like:
```
go run main.go -filterFile path/to/file.txt
```

##Multi-user Systems
After you have set the values in your config file, it is recommended to create a new user for the matrix-logdog process, then change the owner of matrix_logdog.json to that user, then set permissions so that only that user can read the file.
This is to prevent other users on the same machine from seeing your account info.
After creating the user matrixLogdogUser, use:
```
chown matrixLogdogUser matrix_logdog.json
chmod 400 matrix_logdog.json
```
