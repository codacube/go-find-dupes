# Use official Golang image as base
FROM docker.io/library/golang:1.25 AS debug

# RUN apt-get update && apt-get install -y make

# Install Air (live reload tool) for hot-reloading
RUN go install github.com/air-verse/air@latest
RUN go install github.com/go-delve/delve/cmd/dlv@latest

# Set working directory inside container
WORKDIR /app

# For delve
EXPOSE 2345

# Jump into bash
# dlv debug --headless --listen=127.0.0.1:43000 --api-version=2 -- main.go -- gdrive <folder-id>
CMD ["/bin/bash"]

# OR use this to fire up remote debugging using delve
# Arguments must be passed here when delve launches
# FROM https://github.com/charmbracelet/bubbletea
# dlv debug --headless --api-version=2 --listen=127.0.0.1:43000 .
# CMD ["dlv", "debug", "--headless", "--listen=:43000", "--api-version=2", ".", "--", "<cmd_args_if_any>"] 
# OR No file args
# CMD ["dlv", "debug", "--headless", "--listen=:43000", "--api-version=2", "."] 
# dlv exec ./your_binary --api-version=2 --headless --listen=:2345 --accept-multiclient --continue.

# To build the image
# docker build -t go-find-dupes .

# To run it - mount the $(pwd) to the image /app folder (volume mount)
# docker run --rm -it -v ${pwd}:/app -v <local folder, eg. C:\Projects>:/data -p 2345:2345 go-find-dupes

# Once at the bash shell, you'll need to build the executable for debugging, then run delve:
# go build -gcflags="all=-N -l" -o /tmp/go-find-dupes .
# dlv --listen=:2345 --headless --api-version=2 exec /tmp/go-find-dupes -- local /data
# (or add multiclient to keep delve alive, and logging) dlv --listen=:2345 --headless --api-version=2 --accept-multiclient --log --log-output=debugger,dap,rpc exec /tmp/get-hash -- /data/<file_to_hash.ext>
# OR, one line:
# dlv debug --listen=:2345 --headless --api-version=2 -- local /data