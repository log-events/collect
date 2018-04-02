package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/log-events/collect/rfc5424"
	"github.com/spf13/cobra"
	"gopkg.in/mcuadros/go-syslog.v2"
	"gopkg.in/olivere/elastic.v5"
	"gopkg.in/yaml.v2"
)

var rootCmd = &cobra.Command{
	Use:   "collect",
	Short: "Collects logs to elasticsearch",
	Run:   run,
}
var configPath string

type Config struct {
	Listen  string
	Elastic struct {
		URI         string `yaml:"uri"`
		IndexFormat string `yaml:"index-format"`
		DocType     string `yaml:"doc-type"`
		Fields      map[string]interface{}
		Index       map[interface{}]interface{}
	}
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "./collect.yml", "config file")
}

// Execute executes the root command.
func Execute() {
	rootCmd.Execute()
}

func run(cmd *cobra.Command, args []string) {
	channel := make(syslog.LogPartsChannel)
	handler := syslog.NewChannelHandler(channel)

	configFile, err := ioutil.ReadFile(configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not open config file '%s'\n", configPath)
		os.Exit(1)
	}
	var config Config
	err = yaml.Unmarshal(configFile, &config)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Invalid config file\n")
		os.Exit(1)
	}

	elastic, err := elastic.NewClient(elastic.SetURL(config.Elastic.URI))
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not connect to elastic: %s\n", err)
		os.Exit(1)
	}
	defer elastic.Stop()

	server := syslog.NewServer()
	server.SetFormat(syslog.Automatic)
	server.SetHandler(handler)
	if strings.HasPrefix(config.Listen, "tcp://") {
		err = server.ListenTCP(config.Listen[6:])
	} else {
		fmt.Fprintf(os.Stderr, "Invalid listen address '%s'\n", config.Listen)
		os.Exit(1)
	}
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not start listener: %s\n", err)
		os.Exit(1)
	}
	err = server.Boot()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Could not start listener: %s\n", err)
		os.Exit(1)
	}

	indexBody := stringifyYAMLMapKeys(config.Elastic.Index)

	go func(channel syslog.LogPartsChannel) {
		for logParts := range channel {

			indexName := logParts["timestamp"].(time.Time).UTC().Format(config.Elastic.IndexFormat)

			// Create index if not exists
			elastic.CreateIndex(indexName).BodyJson(indexBody).Do(context.Background())

			// Build the document to insert
			if logParts["structured_data"] != nil {
				sd, err := rfc5424.ParseStructuredData(logParts["structured_data"].(string))
				if err != nil {
					fmt.Fprintf(os.Stderr, "Could not parse structured data: %s\n", err)
				} else {
					var sd2 map[string]interface{}
					sd3, _ := json.Marshal(sd)
					json.Unmarshal(sd3, &sd2)
					logParts["structured_data"] = sd2
				}
			}
			data, documentID := getDocumentFromLogParts(logParts, config.Elastic.Fields)

			_, err = elastic.Index().Index(indexName).Type(config.Elastic.DocType).Id(documentID).BodyJson(data).Do(context.Background())
			if err != nil {
				fmt.Fprintf(os.Stderr, "Could not add log entry: %s\n", err)
			}
		}
	}(channel)

	server.Wait()
}

func resolveProperty(object map[string]interface{}, prop string) string {
	sep := strings.Index(prop, ".")
	var val interface{}
	if sep == -1 {
		val = object[prop]
	} else {
		val = object[prop[:sep]]
	}
	if v, ok := val.(string); ok == true {
		return v
	}
	if v, ok := val.(map[string]interface{}); ok == true {
		return resolveProperty(v, prop[sep+1:])
	}
	return ""
}

func getDocumentFromLogParts(logParts syslog.LogParts, fields map[string]interface{}) (map[string]interface{}, string) {
	res := make(map[string]interface{})
	documentID := ""
	if logParts["timestamp"] != nil {
		// If there is a timestamp, generate deterministic ids
		seqID, _ := strconv.ParseInt(resolveProperty(logParts, "structured_data.meta.sequenceId"), 10, 64)
		documentID = fmt.Sprintf("%d%010d", logParts["timestamp"].(time.Time).UTC().UnixNano()/1000, seqID)
	}
	for k, v := range fields {
		// Timestamp is a special case
		if v == "timestamp" {
			res[k] = logParts["timestamp"].(time.Time).UTC().Format(time.RFC3339Nano)
			continue
		}
		if v == "id" {
			res[k] = documentID
			continue
		}
		switch v.(type) {
		case string:
			vv := resolveProperty(logParts, v.(string))
			if vv != "" {
				res[k] = vv
			}
			continue
		case map[interface{}]interface{}:
			vv := v.(map[interface{}]interface{})
			if vv["field"] != nil {
				vvv := resolveProperty(logParts, vv["field"].(string))
				switch vv["type"] {
				case "int":
					if vvvv, err := strconv.ParseInt(vvv, 10, 64); err == nil {
						res[k] = vvvv
						continue
					}
				}
			}

		}
	}
	return res, documentID
}

// stringifyKeysMapValue recurses into in and changes all instances of
// map[interface{}]interface{} to map[string]interface{}. This is useful to
// work around the impedence mismatch between JSON and YAML unmarshaling that's
// described here: https://github.com/go-yaml/yaml/issues/139
//
// Inspired by https://github.com/stripe/stripe-mock, MIT licensed
func stringifyYAMLMapKeys(in interface{}) interface{} {
	switch in := in.(type) {
	case []interface{}:
		res := make([]interface{}, len(in))
		for i, v := range in {
			res[i] = stringifyYAMLMapKeys(v)
		}
		return res
	case map[interface{}]interface{}:
		res := make(map[string]interface{})
		for k, v := range in {
			res[fmt.Sprintf("%v", k)] = stringifyYAMLMapKeys(v)
		}
		return res
	default:
		return in
	}
}
