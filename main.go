package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/matrix-org/gomatrix"

	promlabels "git.cyberia.club/cyberia/promql-parser/pkg/labels"
	promqlparser "git.cyberia.club/cyberia/promql-parser/promql/parser"
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

type AlertRulesBall struct {
	Status string             `json:"status"`
	Data   AlertRulesBallData `json:"data"`
}

type AlertRulesBallData struct {
	Groups []AlertRulesBallGroup `json:"groups"`
}

type AlertRulesBallGroup struct {
	// name
	// file

	Rules []AlertRule `json:"rules"`
}

type AlertRule struct {
	Name string `json:"name"`
	// state
	Query string `json:"query"`
	// duration
	// labels
	// annotations
	Alerts []Alert `json:"alerts"`
	// health
	// lastEvaluation
	// type
}

type Alert struct {
	Labels      map[string]string `json:"labels"`
	Annotations struct {
		Description string `json:"description"`
	} `json:"annotations"`
	State    string    `json:"state"`
	ActiveAt time.Time `json:"activeAt"`
	Value    string    `json:"value"`
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
	_, err := url.Parse(prometheusUrl)
	if prometheusUrl == "" {
		log.Fatalf("JACKAL_PROMETHEUS_URL must be a valid url: %s\n", err)
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
func fetch() AlertRulesBall {
	ball := AlertRulesBall{}
	resp, err := http.Get(prometheusUrl + "/api/v1/rules?type=alert")
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
		rules := fetch()
		if rules.Status != "success" {
			bark("no ball!?!?")
		}

		rulesWithAlertsCurrentlyFiring := []AlertRule{}
		for _, group := range rules.Data.Groups {
			for _, rule := range group.Rules {
				if len(rule.Alerts) > 0 {
					rulesWithAlertsCurrentlyFiring = append(rulesWithAlertsCurrentlyFiring, rule)
				}
			}
		}

		if len(rulesWithAlertsCurrentlyFiring) > 0 {
			for _, rule := range rulesWithAlertsCurrentlyFiring {
				for _, alert := range rule.Alerts {
					generatorURL := fmt.Sprintf("%s/graph?g0.expr=%s", prometheusUrl, url.QueryEscape(rule.Query))

					alertName := "<missing alertname>"
					instance := "<missing instance>"
					if _, has := alert.Labels["alertname"]; has {
						alertName = alert.Labels["alertname"]
					}

					if _, has := alert.Labels["instance"]; has {
						instance = alert.Labels["instance"]
					}

					viewUrl, err := fixupPrometheusGeneratorURL(generatorURL, alert.Labels)

					if err != nil {
						viewUrl = fmt.Sprintf("fixupPrometheusGeneratorURL('%s') failed with '%s'", generatorURL, err)
					}

					bork := fmt.Sprintf("bark!!! %s for %s\n%s\n%s\n",
						alertName, instance, alert.Annotations.Description, viewUrl)

					bark(bork)
				}
			}
			time.Sleep(24 * time.Hour)
		}

		time.Sleep(15 * time.Second)
	}
}

func fixupPrometheusGeneratorURL(urlString string, alertLabels map[string]string) (string, error) {

	originalURL, err := url.Parse(urlString)
	if err != nil {
		return "", err
	}
	originalQuery := originalURL.Query()
	expressionString := originalQuery.Get("g0.expr")

	if expressionString != "" {
		newExpressionString, err := fixupPrometheusExpression(expressionString, alertLabels)
		if err != nil {
			return "", err
		}
		originalQuery.Set("g0.expr", newExpressionString)
	}

	promTimeStamp30MinInTheFuture := time.Now().Add(time.Duration(30) * time.Minute).UTC().Format("2006-01-02 15:04")
	originalQuery.Set("g0.end_input", promTimeStamp30MinInTheFuture)
	originalQuery.Set("g0.moment_input", promTimeStamp30MinInTheFuture)
	originalQuery.Set("g0.range_input", "1h")
	originalQuery.Set("g0.tab", "0")
	originalQuery.Set("g0.stacked", "0")

	promUrl, _ := url.Parse(prometheusUrl)

	newURL := url.URL{
		Scheme:   promUrl.Scheme,
		Host:     promUrl.Host,
		Path:     "/graph",
		RawQuery: originalQuery.Encode(),
	}

	return newURL.String(), nil
}

func fixupPrometheusExpression(expressionString string, alertLabels map[string]string) (string, error) {
	expr, err := promqlparser.ParseExpr(expressionString)
	if err != nil {
		return "", err
	}

	comparisonOperators := map[promqlparser.ItemType]bool{
		promqlparser.GTR:       true,
		promqlparser.GTE:       true,
		promqlparser.LTE:       true,
		promqlparser.LSS:       true,
		promqlparser.EQL:       true,
		promqlparser.EQL_REGEX: true,
		promqlparser.EQLC:      true,
		promqlparser.NEQ:       true,
		promqlparser.NEQ_REGEX: true,
	}

	// This function returns true if niether this expression nor any of its recursive children are query-ish
	// in other words it returns true for ((90+10)*100) but returns false for irate(node_cpu_seconds_total{mode="idle"}[10m])
	isLiteral := func(expr promqlparser.Expr) bool {
		blah := false
		foundDynamicContents := &blah

		promqlparser.Inspect(expr, func(node promqlparser.Node, path []promqlparser.Node) error {
			switch typedNode := node.(type) {
			case *promqlparser.VectorSelector:
				*foundDynamicContents = (typedNode != nil)
			case *promqlparser.SubqueryExpr:
				*foundDynamicContents = (typedNode != nil)
			case *promqlparser.MatrixSelector:
				*foundDynamicContents = (typedNode != nil)
			case *promqlparser.AggregateExpr:
				*foundDynamicContents = (typedNode != nil)
			case *promqlparser.EvalStmt:
				*foundDynamicContents = (typedNode != nil)
			default:

			}
			return nil
		})

		return !*foundDynamicContents
	}

	var newRootOfExpr *promqlparser.Expr = &expr

	setNewRootOfExpr := func(x *promqlparser.Expr) {
		newRootOfExpr = x
	}

	promqlparser.Inspect(expr, func(node promqlparser.Node, path []promqlparser.Node) error {
		if node != nil {
			switch typedNode := node.(type) {
			case *promqlparser.BinaryExpr:

				_, isComparison := comparisonOperators[typedNode.Op]
				isOutermost := len(path) == 0

				//log.Infof("BinaryExpr: Op: %d. isComparison: %t. isOutermost: %t", int(typedNode.Op), isComparison, isOutermost)
				// if the outermost Node is a binary comparison op like x > y, a <= b, etc
				if isComparison && isOutermost {
					// now we have to figure out which side is dynamic and which side is static (only contains literals)
					lhsIsLiteral := isLiteral(typedNode.LHS)
					rhsIsLiteral := isLiteral(typedNode.RHS)

					//log.Infof("BinaryExpr: lhsIsLiteral: %t. rhsIsLiteral: %t", lhsIsLiteral, rhsIsLiteral)
					// if one side is dynamic, make that side be the new root of the expression (trim the comparison off the outside)
					if !lhsIsLiteral && rhsIsLiteral {
						setNewRootOfExpr(&(typedNode.LHS))
					}
					if lhsIsLiteral && !rhsIsLiteral {
						setNewRootOfExpr(&(typedNode.RHS))
					}
				}
			case *promqlparser.VectorSelector:
				alreadyQualifiedLabels := map[string]bool{
					"severity":  true,
					"alertname": true,
				}
				for _, v := range typedNode.LabelMatchers {
					alreadyQualifiedLabels[v.Name] = true
				}

				//bytes, _ := json.MarshalIndent(alreadyQualifiedLabels, "", "  ")
				//log.Infof("VectorSelector: alreadyQualifiedLabels: %s", string(bytes))

				// qualify any vectorSelector in the query with the labels that we know matched the alert condition
				// as long as this specific label is not already matched on
				for k, v := range alertLabels {
					//log.Infof("alertLabels: %s=%s,  apply = %t", k, v, !alreadyQualifiedLabels[k])
					if !alreadyQualifiedLabels[k] {
						typedNode.LabelMatchers = append(typedNode.LabelMatchers, &promlabels.Matcher{
							Type:  promlabels.MatchEqual,
							Name:  k,
							Value: v,
						})

					}
				}
			default:
			}
		}

		return nil
	})

	toReturn := (*newRootOfExpr).(promqlparser.Node).String()
	log.Printf("fixupPrometheusExpression: %s -> %s\n", expressionString, toReturn)

	return toReturn, nil
}
