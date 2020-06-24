# Dragonfly
Dragonfly is a server software for Minecraft Bedrock Edition written in Go. It was written with scalability
and simplicity in mind and aims to make the process of setting up a server and modifying it easy.

## Setup
There are currently no prebuilt executables available. These may be provided later once Dragonfly moves to a
more stable stage.

### Development setup
Installing/compiling Dragonfly requires at least Go 1.13.

##### Instant install, when GOPATH is added to $PATH:
```
go install github.com/df-mc/dragonfly
```
Running:
```
dragonfly
```

##### Installation for editing Dragonfly:
```
git clone https://github.com/df-mc/dragonfly
cd dragonfly
```
Running:
```
go run main.go
```

## Usage
After starting the Dragonfly server, messages will be logged to the console. Console commands are currently
not implemented in Dragonfly, so writing commands will not do anything. The server may be stopped by running
`ctrl+c` at any time.

## Developer info
Dragonfly features a well-documented codebase with an easy to use API. Automatically generated documentation
may be found [here](https://pkg.go.dev/github.com/df-mc/dragonfly/dragonfly?tab=doc) and in the subpackages
found by clicking 'Subdirectories'.
The GitHub wiki will hold examples of frequently used functionality.

## Contributing
We use JetBrains Space to manage our issues, pull requests and code reviews, but we welcome contributions
through GitHub issues and pull requests.

## Contact
[![Chat on Discord](https://img.shields.io/badge/Chat-On%20Discord-738BD7.svg?style=for-the-badge)](https://discord.gg/evzQR4R)
