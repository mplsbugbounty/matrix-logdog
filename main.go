package main

import (
    "context"
    "sync"
    "errors"
	"log"
	"os"
    "strings"
    "regexp"
    "bufio"
    "flag"
    "encoding/json"
    "crypto/sha256"

	_ "github.com/mattn/go-sqlite3"
    "github.com/fsnotify/fsnotify"
    "maunium.net/go/mautrix"
	"maunium.net/go/mautrix/crypto/cryptohelper"
    id "maunium.net/go/mautrix/id"

)

type StringSet map[string]struct{}
type HashByFilepath map[string][32]byte

type FilepathHashPair struct{
    Filename string
    Sha256Sum [32]byte
}

//for unmarshalling the config file
type CredzNPathz struct {
    MatrixHomeserver string
    MatrixUser id.UserID
    MatrixRoom id.RoomID
    MatrixPassword string

    WatchDir string
    Filters []string
    SQLiteDatabase string
}

// globalz lol
var (
    matchTermsFile string
    configFile string
    barkedSet StringSet
	cli *mautrix.Client
)


//possibly could further abstract these
//but why?
func logFCheck(e error){
    if e != nil{
        log.Fatal(e)
    }
}

func printCheck(e error){
    if e != nil{
        log.Println(e)
    }
}

func panicCheck(e error){
    if e != nil{
        log.Panicln(e)
    }
}

func cloneHashByFile( mapIn HashByFilepath ) ( mapOut HashByFilepath ) {
    mapOut = make(HashByFilepath)
    for k, v := range mapIn {
            mapOut[k] = v
        }
        return
}

func isMemberOfSet( set StringSet, strIn string ) bool {
    for item,_ := range set{
        if strIn == item {
            return true
        }
    }
    return false
}

func parseEnv() {
	configFile = os.Getenv("MATRIX_LOGDOG_CONFIG_FILE")
    if configFile == "" {
        configFile = "matrix_logdog.json"
	}
}

func parseConfigJson ( jsonFilepath string) ( configOut CredzNPathz ) {
    content, err := os.ReadFile(jsonFilepath)
    panicCheck(err)
    err = json.Unmarshal(content, &configOut)
    return
}

func parseTermsFile ( filepathIn string ) ( fileLines []string ) {

    file, err := os.Open(filepathIn)
    defer file.Close()
    printCheck(err)
    fileScanner := bufio.NewScanner(file)
    fileScanner.Split(bufio.ScanLines)
  
    for fileScanner.Scan() {
        if fileScanner.Text() != "" {
            fileLines = append(fileLines, fileScanner.Text())
        }
    }
    return
}

func bark(text string, matrixRoom id.RoomID) {
	_, err := cli.SendText(matrixRoom, text)
    printCheck(err)
}

func watch( watchDir string, eventChan chan<- string) {
    watcher, err := fsnotify.NewWatcher()
    logFCheck(err)
    err = watcher.Add( watchDir )
    logFCheck(err)
    log.Println("Watching...")
    defer watcher.Close()
    for {
        select {
        case event, ok := <-watcher.Events:
            if !ok {
                log.Println("Error reading fsnotify event in watcher.Events")
                return
            }
            if event.Has(fsnotify.Write) {
                eventChan <-event.Name
            }
        case err, ok := <-watcher.Errors:
            if !ok {
                log.Println("Error reading fsnotify event in watcher.Erros")
                return
            }
            log.Println("error:", err)
        }
    }
}

func barkIfFound( term string , line string, matrixRoom id.RoomID ) {

    found , _ := regexp.MatchString(term, line)
    if found {
        if !isMemberOfSet(barkedSet, line) {
            barkedSet[line] = struct{}{}
            log.Println("barking: ", line )
            bark(line, matrixRoom)
        }
    }
}

func checkLinesAgainstTermsBarkIfFound( fileContent []byte, terms []string, matrixRoom id.RoomID  ) {

        newText := strings.Split(string(fileContent), "\n")
        for _, line := range newText {
            for _, term := range terms {
                barkIfFound(term , line , matrixRoom )
            }
        }
}

func search( terms []string , currentFileHashes HashByFilepath, event string, hashChan chan<- FilepathHashPair , matrixRoom id.RoomID ) {

        file, err := os.ReadFile( event )
        printCheck(err)
        if err != nil{
            return
        }

        var fileHashPair FilepathHashPair
        fileHashPair.Filename = event
        fileHashPair.Sha256Sum = sha256.Sum256( file )
        if storedFileHash, notFirstHashing := currentFileHashes[ event ]; notFirstHashing {
            if fileHashPair.Sha256Sum != storedFileHash {
                hashChan <- fileHashPair
                checkLinesAgainstTermsBarkIfFound( file, terms , matrixRoom )
            } else {
                return
            }
        } else /*if firstHashing*/ {
            hashChan <- fileHashPair
            checkLinesAgainstTermsBarkIfFound( file, terms, matrixRoom )
        }
}

func main() {

	parseEnv()
	var err error
    var event string
    configStruct := parseConfigJson( configFile )

    matchTermsFile := flag.String("filterFile","","Optional newline-delimited list of filter terms.")
    var terms []string
    if *matchTermsFile != "" {
        terms = parseTermsFile( *matchTermsFile )
    } else {
        terms = configStruct.Filters
    }
    log.Println("terms: ", terms)
	cli, err = mautrix.NewClient(configStruct.MatrixHomeserver, "","")
    panicCheck(err)
    
    barkedSet = StringSet{}
    userLocalPart := string(configStruct.MatrixUser)
    
	cryptoHelper, err := cryptohelper.NewCryptoHelper(cli, []byte("meow"), configStruct.SQLiteDatabase)
	cryptoHelper.LoginAs = &mautrix.ReqLogin{
		Type:       mautrix.AuthTypePassword,
		Identifier: mautrix.UserIdentifier{Type: mautrix.IdentifierTypeUser, User: userLocalPart},
		Password:   configStruct.MatrixPassword,
	}
    panicCheck(err)
	err = cryptoHelper.Init()
    panicCheck(err)
    cli.Crypto = cryptoHelper

	syncCtx, _ := context.WithCancel(context.Background())
	var syncStopWait sync.WaitGroup
	syncStopWait.Add(1)

	go func() {
		err = cli.SyncWithContext(syncCtx)
		defer syncStopWait.Done()
		if err != nil && !errors.Is(err, context.Canceled) {
			panic(err)
		}
	}()

    eventChan :=  make(chan string)
    hashChan :=  make(chan FilepathHashPair )
	// lmao, bark bark
    currentFileHashes := make(HashByFilepath) 
    go watch( configStruct.WatchDir , eventChan )
	for {

            select {
            case event = <-eventChan:
                log.Println("event triggered")
                currentHashesCopy := cloneHashByFile(currentFileHashes)
                go search( terms, currentHashesCopy , event , hashChan, configStruct.MatrixRoom )
            case newHash := <-hashChan:
                currentFileHashes[newHash.Filename] = newHash.Sha256Sum
                log.Println("Change detected in ", newHash.Filename)
            }
		}
    err = cryptoHelper.Close()
    logFCheck(err)
}

