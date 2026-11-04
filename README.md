# Lean Quiz

A local web app that teaches Lean, one lesson per day. Each day has a short reading and a few questions. Answers are graded in the browser and saved on disk.

There are 21 lessons. The first unlocks on 3 November 2026 at 08:00 local time, then one more each calendar day at 08:00. Before that hour, the previous day's lesson stays open. Later lessons stay locked until their day arrives.

## Run

From `quizapp/`:

```sh
go run .
```

The server listens on `http://127.0.0.1:8787`. Set `LEAN_QUIZ_ADDR` to use another address.

Progress is stored in `~/Library/Application Support/LeanQuiz/state.json`.

## Tests

```sh
cd quizapp
go test ./...
```

## Background service

`script/com.sfeldma.lean-quiz.plist` is a launchd agent that runs the built binary at `quizapp/lean-quiz`, restarts it if it exits, and writes logs under `logs/`. Copy it to `~/Library/LaunchAgents` and load it with `launchctl` to start the server at login. Build the binary first:

```sh
cd quizapp
go build -o lean-quiz .
```
