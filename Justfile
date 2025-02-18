default: build

name := "strafe"
version := "0.1.0"

build_dir := "dist"
build_flags := "-trimpath -ldflags='-s -w'"

list:
    @just --list

clean:
    rm -rf {{build_dir}}

setup:
    mkdir -p {{build_dir}}

generate:
    sqlc generate

tidy:
    go mod tidy

build: clean setup tidy generate
    #!/usr/bin/env sh
    GOOS=linux GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-linux-amd64 ./cmd
    GOOS=linux GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-linux-arm64 ./cmd

    GOOS=darwin GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-darwin-amd64 ./cmd
    GOOS=darwin GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-darwin-arm64 ./cmd

    GOOS=windows GOARCH=amd64 go build {{build_flags}} -o {{build_dir}}/{{name}}-windows-amd64.exe ./cmd
    GOOS=windows GOARCH=arm64 go build {{build_flags}} -o {{build_dir}}/{{name}}-windows-arm64.exe ./cmd

    chmod +x {{build_dir}}/{{name}}-linux-*
    chmod +x {{build_dir}}/{{name}}-darwin-*

build-current: tidy generate setup
    go build {{build_flags}} -o {{build_dir}}/{{name}} ./cmd
    chmod +x {{build_dir}}/{{name}}



package: build
    #!/usr/bin/env sh
    cd {{build_dir}}
    
    tar czf {{name}}-linux-amd64.tar.gz {{name}}-linux-amd64
    tar czf {{name}}-linux-arm64.tar.gz {{name}}-linux-arm64
    
    tar czf {{name}}-darwin-amd64.tar.gz {{name}}-darwin-amd64
    tar czf {{name}}-darwin-arm64.tar.gz {{name}}-darwin-arm64
    
    zip {{name}}-windows-amd64.zip {{name}}-windows-amd64.exe
    zip {{name}}-windows-arm64.zip {{name}}-windows-arm64.exe

docker:
    docker build -t {{name}}:{{version}} .