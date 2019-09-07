package logger

import "github.com/mitchellh/colorstring"

func LogInfo(message string) {
	colorstring.Println("[blue] [Info]: " + message)
}

func LogCritical(message string){
	colorstring.Println("[yellow] [Critical]: " + message)
}

func LogError(message string){
	colorstring.Println("[red] [Error]: " + message)
}
