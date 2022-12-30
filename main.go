package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/matrix-org/gomatrix"
)

// globalz lol
var (
	matrixUrl   string
	matrixUser  string
	matrixToken string
	matrixRoom  string

	prometheusUrl string

	cli *gomatrix.Client
)

type AlertBall struct {
	Status string `json:"status"`
	Data   struct {
		Alerts []struct {
			Labels struct {
				Alertname string `json:"alertname"`
				Instance  string `json:"instance"`
				Job       string `json:"job"`
				Severity  string `json:"severity"`
			} `json:"labels"`
			Annotations struct {
				Description string `json:"description"`
			} `json:"annotations"`
			State    string    `json:"state"`
			ActiveAt time.Time `json:"activeAt"`
			Value    string    `json:"value"`
		} `json:"alerts"`
	} `json:"data"`
}

func parseEnv() {
	matrixUrl = os.Getenv("JACKAL_MATRIX_URL")
	if matrixUrl == "" {
		log.Fatal("JACKAL_MATRIX_URL is required")
	}
	matrixUser = os.Getenv("JACKAL_MATRIX_USER")
	if matrixUser == "" {
		log.Fatal("JACKAL_MATRIX_USER is required")
	}
	matrixToken = os.Getenv("JACKAL_MATRIX_TOKEN")
	if matrixToken == "" {
		log.Fatal("JACKAL_MATRIX_TOKEN is required")
	}
	matrixRoom = os.Getenv("JACKAL_MATRIX_ROOM")
	if matrixRoom == "" {
		log.Fatal("JACKAL_MATRIX_ROOM is required")
	}
	prometheusUrl = os.Getenv("JACKAL_PROMETHEUS_URL")
	if prometheusUrl == "" {
		log.Fatal("JACKAL_PROMETHEUS_URL is required")
	}
}

func bark(text string) {
	_, err := cli.SendText(matrixRoom, text)
	if err != nil {
		log.Println(err)
	}
}

// fetchAlerts expects a URL to a Prometheus server
// & will construct URLs to fetch alerts and such
func fetch() AlertBall {
	ball := AlertBall{}
	resp, err := http.Get(prometheusUrl + "/api/v1/alerts")
	if err != nil {
		log.Println(err)
		bark("where ball?!??")
		return ball
	}
	jsonBlob, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		bark("how ball?!l???")
		return ball
	}
	err = json.Unmarshal(jsonBlob, &ball)
	if err != nil {
		log.Println(err)
		bark("wrong ball??!??")
		return ball
	}
	return ball
}

func main() {
	parseEnv()
	var err error
	cli, err = gomatrix.NewClient(matrixUrl, matrixUser, matrixToken)
	if err != nil {
		log.Panic(err)
	}

	// lmao, bark bark
	for {
		alerts := fetch()
		if alerts.Status != "success" {
			bark("no ball!?!?")
		}
		if len(alerts.Data.Alerts) > 0 {
			for _, alert := range alerts.Data.Alerts {
				bork := fmt.Sprintf("bark!!! %s for %s\n%s",
					alert.Labels.Alertname, alert.Labels.Instance, alert.Annotations.Description)
				bark(bork)
			}
			time.Sleep(24 * time.Hour)
		}
		time.Sleep(15 * time.Second)
	}
}
