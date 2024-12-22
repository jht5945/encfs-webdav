_:
  @just --list

# Build encfs-webdav for Linux-x86-64
build-linux-amd64:
  env GOOS=linux GOARCH=amd64 go build -o encfs-webdav-linux-amd64 main.go


