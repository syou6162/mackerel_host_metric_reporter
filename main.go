package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/Songmu/horenso"
	mkr "github.com/mackerelio/mackerel-client-go"
)

func getReport(r io.Reader) (*horenso.Report, error) {
	bytes, err := ioutil.ReadAll(r)
	if err != nil {
		return nil, err
	}

	report := horenso.Report{}

	err = json.Unmarshal(bytes, &report)
	if err != nil {
		return nil, err
	}
	return &report, nil
}

func getElapsedTime(report horenso.Report) (float64, error) {
	if report.ExitCode != 0 {
		return 0.0, fmt.Errorf("report.ExitCode is %d", report.ExitCode)
	}
	return report.EndAt.Sub(*report.StartAt).Seconds(), nil
}

func postReportToMackerel(report horenso.Report, hostID string) error {
	apiKey := os.Getenv("MACKEREL_APIKEY")
	if apiKey == "" {
		return errors.New("Cannot find MACKEREL_APIKEY")
	}

	client := mkr.NewClient(apiKey)
	now := time.Now().Unix()

	t, err := getElapsedTime(report)
	if err != nil {
		return err
	}

	err = client.PostHostMetricValuesByHostID(hostID, []*mkr.MetricValue{
		{
			Name:  fmt.Sprintf("batch_elapsed_time.%s", report.Tag),
			Time:  now,
			Value: t,
		},
	})
	return err
}

func main() {
	if len(os.Args) < 2 {
		return
	}
	hostID := os.Args[1]

	report, err := getReport(os.Stdin)
	if err != nil {
		fmt.Println(err)
		return
	}

	err = postReportToMackerel(*report, hostID)
	if err != nil {
		fmt.Println(err)
		return
	}
}
