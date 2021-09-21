package main

import (
	"bytes"
	"io"
	"net/http"
	"os"

	jsoniter "github.com/json-iterator/go"
	"github.com/pkg/errors"
	log "unknwon.dev/clog/v2"

	"github.com/asoul-video/asoul-video/pkg/model"

	"github.com/asoul-video/acao/source"
)

func main() {
	defer log.Stop()
	err := log.NewConsole()
	if err != nil {
		panic(err)
	}

	reportType := os.Getenv("SOURCE_REPORT_TYPE")

	src, ok := source.Sources[reportType]
	if !ok {
		log.Fatal("Report type not found")
	}

	resultChan := make(chan source.Result, 5)
	go src.Scrap(resultChan)

	for result := range resultChan {
		if result.End {
			break
		}

		var err error
		for i := 1; i <= 5; i++ { // Retry 5 times.
			if err = reportData(model.ReportType(reportType), result.Data); err != nil {
				log.Warn("Failed to report data: %v, retry %d / 5", err, i)
				continue
			}
		}
		if err != nil {
			log.Error("Failed to report data: %v", err)
		}
	}
}

func reportData(reportType model.ReportType, reportData jsoniter.RawMessage) error {
	reportURL := os.Getenv("SOURCE_REPORT_URL")
	reportKey := os.Getenv("SOURCE_REPORT_KEY")

	bodyBytes, err := jsoniter.Marshal(map[string]interface{}{
		"type": reportType,
		"data": reportData,
	})
	if err != nil {
		return errors.Wrap(err, "encode JSON")
	}

	req, err := http.NewRequest(http.MethodPost, reportURL, bytes.NewReader(bodyBytes))
	if err != nil {
		return errors.Wrap(err, "new request")
	}

	req.Header.Set("Authorization", reportKey)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return errors.Wrap(err, "request")
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode/100 != 2 {
		bodyBytes, err := io.ReadAll(resp.Body)
		if err != nil {
			return errors.Wrap(err, "read response body")
		}
		return errors.Errorf("unexpected status code %d: %q", resp.StatusCode, string(bodyBytes))
	}

	return nil
}
