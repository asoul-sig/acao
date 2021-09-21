# 🦙 acao ![Go](https://github.com/asoul-video/acao/workflows/Go/badge.svg) [![Go Report Card](https://goreportcard.com/badge/github.com/asoul-video/acao)](https://goreportcard.com/report/github.com/asoul-video/acao) [![Sourcegraph](https://img.shields.io/badge/view%20on-Sourcegraph-brightgreen.svg?logo=sourcegraph)](https://sourcegraph.com/github.com/asoul-video/acao)

acao（阿草）, the tool man for data scraping of https://asoul.video/.

## Deploy to Aliyun serverless function with [Raika](https://github.com/serverless-moe/Raika)

### `update_member` Update A-SOUL member profile.

```bash
$ GOOS=linux go build .

$ Raika function create \
    --name asoul_video_update_member \
    --memory 128 \
    --init-timeout 10 \
    --runtime-timeout 10 \
    --binary-file acao \
    --trigger=cron \
    --cron="0 30 * * * *" \
    --env SOURCE_REPORT_TYPE=update_member \
    --env SOURCE_REPORT_URL=https://asoul.video/source/report \
    --env SOURCE_REPORT_KEY=<REDACTED> \
    --platform aliyun
```

### `create_video` Fetch A-SOUL member's videos from Douyin.

```bash
$ GOOS=linux go build .

$ Raika function create \
    --name asoul_video_create_video \
    --memory 128 \
    --init-timeout 10 \
    --runtime-timeout 10 \
    --binary-file acao \
    --trigger=cron \
    --cron="0 30 * * * *" \
    --env SOURCE_REPORT_TYPE=create_video \
    --env SOURCE_REPORT_URL=https://asoul.video/source/report \
    --env SOURCE_REPORT_KEY=<REDACTED> \
    --platform aliyun
```

## License

MIT
