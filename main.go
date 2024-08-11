package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/guptarohit/asciigraph"
)

var data []float64
var dataPoints int
var xAxisDateValues, xAxisTimeValues, key, value string
var promOutputDataResult PromOutputDataResult

func main() {
	lFlag := flag.String("l", "", "key=value label to be used to filter among the metrics present, needed only if \".data.result | length > 1\" ")

	flag.Parse()

	dataIn, err := io.ReadAll(os.Stdin)
	if err != nil {
		fmt.Println("Error reading from stdin:", err)
		return
	}
	promOutput := &PromOutput{}
	err = json.Unmarshal([]byte(dataIn), &promOutput)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if *lFlag != "" {
		parts := strings.SplitN(*lFlag, "=", 2)
		if len(parts) == 2 {
			key = parts[0]
			value = parts[1]
		}
	} else {
		if len(promOutput.Data.Result) > 1 {
			fmt.Println("Invalid format for -l flag. Expected format: key=value, choose one label among the followings:")
			labels := make(map[string]int)
			for _, _promOutputDataResult := range promOutput.Data.Result {
				for key, value := range _promOutputDataResult.Metric {
					pair := fmt.Sprintf("%s=%s", key, value)
					labels[pair] += 1
				}
			}
			var keys []string
			for label, value := range labels {
				if value == 1 {
					keys = append(keys, label)
				}
			}
			sort.Strings(keys)
			fmt.Println("")
			for _, label := range keys {
				fmt.Println("  ", label)
			}
			os.Exit(1)
		}
		if len(promOutput.Data.Result) == 1 {
			promOutputDataResult = promOutput.Data.Result[0]
		}
		fmt.Println("The Prometheus query returned an empty output.")
		os.Exit(1)
	}
	for _, _promOutputDataResult := range promOutput.Data.Result {
		_value, ok := _promOutputDataResult.Metric[key]
		if ok {
			if _value == value {
				promOutputDataResult = _promOutputDataResult
			}
		}
	}

	dataPoints = len(promOutputDataResult.Values)
	if dataPoints < 207 {
		for _, epochValuePair := range promOutputDataResult.Values {
			epoch := epochValuePair[0]
			epochFloat64 := epoch.(float64)
			epochInt64 := int64(epochFloat64)
			formattedTime := time.Unix(epochInt64, 0).Format("15:04")
			fmt.Println("Formatted time:", formattedTime)
			value := epochValuePair[1]
			valueFloat, _ := strconv.ParseFloat(value.(string), 64)
			data = append(data, valueFloat)
		}
	} else {
		for i := 0; i < len(promOutputDataResult.Values); i += 2 {
			epochValuePair := promOutputDataResult.Values[i]
			epoch := epochValuePair[0]
			epochFloat64 := epoch.(float64)
			epochInt64 := int64(epochFloat64)
			formattedTime := time.Unix(epochInt64, 0).Format("15:04")
			value := epochValuePair[1]
			valueFloat, _ := strconv.ParseFloat(value.(string), 64)
			data = append(data, valueFloat)
			formattedDate := time.Unix(epochInt64, 0).Format("01/02")
			if i == 0 {
				xAxisDateValues = xAxisDateValues + "     " + formattedDate
				xAxisTimeValues = xAxisTimeValues + "     " + formattedTime
			} else {
				if i%7 == 0 && i >= 7 {
					xAxisDateValues = xAxisDateValues + "  " + formattedDate
					xAxisTimeValues = xAxisTimeValues + "  " + formattedTime
				}
			}
		}
	}
	graph := asciigraph.Plot(data, asciigraph.Height(25), asciigraph.Width(len(promOutputDataResult.Values)/2), asciigraph.Offset(4))
	fmt.Println("")
	fmt.Println(graph)
	fmt.Println("       └──────┬" + strings.Repeat("──────┬", (dataPoints/2-6)/7) + "─────")
	fmt.Println(xAxisDateValues)
	fmt.Println(xAxisTimeValues)
}

type PromOutput struct {
	Status string         `json:"status"`
	Data   PromOutputData `json:"data"`
}

type PromOutputData struct {
	ResultType string                 `json:"resultType"`
	Result     []PromOutputDataResult `json:"result"`
}

type PromOutputDataResult struct {
	Metric map[string]string `json:"metric"`
	Values []FloatStringPair `json:"values"`
}

type FloatStringPair [2]interface{}
