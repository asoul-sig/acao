package main

import (
	"bytes"
	"io"
	"net/http"
	"os"

	"github.com/flamego/flamego"
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

	source, ok := source.Sources[reportType]
	if !ok {
		log.Fatal("Report type not found")
	}

	f := flamego.Classic()

	f.NotFound(func(ctx flamego.Context) {
		respData, err := source.Scrap()
		if err != nil {
			log.Error("Failed to scrap source: %v", err)
			_, _ = ctx.ResponseWriter().Write([]byte(err.Error()))
			ctx.ResponseWriter().WriteHeader(http.StatusInternalServerError)
			return
		}

		for _, data := range respData {
			if err := reportData(model.ReportType(reportType), data); err != nil {
				log.Error("Failed to report data: %v", err)
				_, _ = ctx.ResponseWriter().Write([]byte(err.Error()))
				ctx.ResponseWriter().WriteHeader(http.StatusInternalServerError)
				return
			}
		}

		ctx.ResponseWriter().WriteHeader(http.StatusNoContent)
	})

	f.Run(9000)
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
