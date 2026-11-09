# go-find-dupes

Finds duplicate files within a given folder. Current sources are searching the local hdd (Full path) and Google Drive (Folder ID). Uses a strategy pattern for the different file sources, and a pipeline pattern for processing the files.

- Gets a list of all files for the given folder
- Groups all files by file size
- Uses sha256 to find the duplicates within the file size groups
- The user is presented the dupicate groups and can select files to delete

~~# Run tests for all packages~~

~~go test ./...~~

~~# OR run with verbose output to see individual test names~~

~~go test -v ./internal/processor~~

~~`make`~~
~~`make run-mock`~~
~~`make test`~~
~~`make clean`~~

~~`make build-all`~~

# Live Reload whilst debugging

- Air doesn't work, missing a TTY
- Watchexec works, but has to be run locally
  - For Manjaro:
    - `pamac install delve`
    - `pamac install watchexec`
    - `watchexec -r -e go -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient --continue -- local <local_folder>` (remove --continue if debug session starts immediately and doesn't stop on breakpoints)
    - To stop `watchexec`: `CTRL-C`, `CTRL-\`, `pkill watchexec`, or `pkill -9 watchexec`

- `watchexec -r -e go -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- gdrive -c <path-to-token.json> -f <folder-id>`

- `watchexec -r -e go -- dlv debug --headless --listen=:2345 --api-version=2 --accept-multiclient -- local <path-to-scan>`

`dlv debug --headless --listen=127.0.0.1:2345 --api-version=2 -- local "<path-to-scan>"`

## Compiling

`go build -o ./bin/ ./...`
