package main

import (
	"log"
	"os"
    "strings"
    "regexp"
    "bufio"

    "github.com/fsnotify/fsnotify"
	//"github.com/matrix-org/gomatrix"
    "maunium.net/go/mautrix"
    id "maunium.net/go/mautrix/id"

)

// globalz lol
var (
	matrixUrl   string
	matrixUser  id.UserID
	matrixToken string
	matrixRoom  id.RoomID

    watchDir string
    matchTermsFile string

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

func parseEnv() {
	matrixUrl = os.Getenv("MATRIX_LOGDOG_URL")
	if matrixUrl == "" {
		log.Fatal("MATRIX_LOGDOG_URL is required")
	}
	matrixUser = id.UserID(os.Getenv("MATRIX_LOGDOG_USER"))
	if matrixUser == "" {
		log.Fatal("MATRIX_LOGDOG_USER is required")
	}
	matrixToken = os.Getenv("MATRIX_LOGDOG_TOKEN")
	if matrixToken == "" {
		log.Fatal("MATRIX_LOGDOG_TOKEN is required")
	}
	matrixRoom = id.RoomID(os.Getenv("MATRIX_LOGDOG_ROOM"))
	if matrixRoom == "" {
		log.Fatal("MATRIX_LOGDOG_ROOM is required")
	}
	watchDir = os.Getenv("MATRIX_LOGDOG_WATCH_DIR")
    if watchDir == "" {
		log.Fatal("MATRIX_LOGDOG_WATCH_DIR is required")
	}
	matchTermsFile = os.Getenv("MATRIX_LOGDOG_MATCH_FILE")
    if matchTermsFile == "" {
		log.Fatal("MATRIX_LOGDOG_MATCH_FILE is required")
	}
}

func parseTermsFile ( filepathIn string ) ( fileLines []string ) {

    file, err := os.Open(filepathIn)
    defer file.Close()
    printCheck(err)
    fileScanner := bufio.NewScanner(file)
    fileScanner.Split(bufio.ScanLines)
  
    for fileScanner.Scan() {
        fileLines = append(fileLines, fileScanner.Text())
    }
    return
}

func bark(text string) {
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

func search( terms []string , byteOffsetForRead int64, event string, sizeChan chan<- int64 ) {
        file, err := os.Open( event )
        printCheck(err)
        defer file.Close()
        fileinfo, err := file.Stat()
        printCheck(err)
        fileSize := fileinfo.Size()
        sizeChan <-fileSize
        bufferSize := fileSize - byteOffsetForRead
        //fully reload the file if its length is less than last offset
        if bufferSize < 0 {
            bufferSize = fileSize
            byteOffsetForRead = 0
        }
        buffer := make([]byte, bufferSize)
        bytesRead , err := file.ReadAt(buffer, byteOffsetForRead)
        printCheck(err)
        if err == nil{
            log.Println("bytesRead:", bytesRead)
            newText := strings.Split(string(buffer), "\n")
            for _, line := range newText {
                for _, term := range terms {
                    found , _ := regexp.MatchString(term, line)
                    if found {
                        log.Println("barking: ", line )
                        bark(line)
                    }
                }
            }
        }
}

func main() {
	parseEnv()
	var err error
    var event string
    terms := parseTermsFile( matchTermsFile )
    log.Println("terms: ", terms)
	cli, err = mautrix.NewClient(matrixUrl, matrixUser, matrixToken)
    panicCheck(err)
    
    eventChan :=  make(chan string)
    sizeChan :=  make(chan int64)
	// lmao, bark bark
    currentFileSize := int64(0)
    go watch( watchDir , eventChan )
	for {

            select {
            case event = <-eventChan:
                log.Println("event triggered")
                go search( terms, currentFileSize , event , sizeChan )
            case currentFileSize = <-sizeChan:
                log.Println("filesize change")
            }
		}
}

