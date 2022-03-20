<!--suppress ALL -->
<img height="310" alt="image" src="https://user-images.githubusercontent.com/16114089/121805566-0cd81280-cc4c-11eb-9b7d-b5f8a6db4f8d.png" align="right">

# Dragonfly

Dragonfly is a heavily asynchronous server software for Minecraft Bedrock Edition written in Go. It was written with scalability
and simplicity in mind and aims to make the process of setting up a server and modifying it easy. Unlike other
Minecraft server software, Dragonfly is generally used as a library to extend.

[![Discord Banner 2](https://discordapp.com/api/guilds/623638955262345216/widget.png?style=banner2)](https://discord.gg/U4kFWHhTNR)

## Getting started
Running Dragonfly requires at least **Go 1.18**. After starting the server through one of the methods below,
**ctrl+c** may be used to shut down the server. Also check out the [wiki](https://github.com/df-mc/dragonfly/wiki) for
more detailed info.

#### Installation as library
```
go mod init github.com/<user>/<module name>
go get github.com/df-mc/dragonfly
```

![SetupLibrary](https://user-images.githubusercontent.com/16114089/121804512-0f843900-cc47-11eb-9320-d195393b5a1f.gif)

#### Installation of the latest commit
```
git clone https://github.com/df-mc/dragonfly
cd dragonfly
go run main.go
```

![SetupClone](https://user-images.githubusercontent.com/16114089/121804495-ff6c5980-cc46-11eb-8e31-df4d94782e5b.gif)


## Developer info
[![Go Reference](https://pkg.go.dev/badge/github.com/df-mc/dragonfly/server.svg)](https://pkg.go.dev/github.com/df-mc/dragonfly/server)

Dragonfly features a well-documented codebase with an easy-to-use API. Documentation may be found
[here](https://pkg.go.dev/github.com/df-mc/dragonfly/server) and in the subpackages found by clicking *Directories*.

Publishing your project on GitHub? Consider adding the **[#df-mc](https://github.com/topic/df-mc)** topic to your
repository to improve visibility of your project.

## Contributing
Contributions are very welcome! Issues, pull requests and feature requests are highly appreciated. Opening a pull
request? Consider joining our [Discord server](https://discord.gg/U4kFWHhTNR) to discuss your changes! Also have a read through the
[CONTRIBUTING.md](https://github.com/df-mc/dragonfly/blob/master/.github/CONTRIBUTING.md) for more info.
