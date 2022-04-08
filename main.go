package main

import (
	"embed"
	"fmt"

	"github.com/VonC/barerepo/version"
	"github.com/alecthomas/kong"
	"github.com/ryboe/q"
	"github.com/spewerspew/spew"
)

// CLI stores arguments and subcommands
type CLI struct {
	Debug     bool        `help:"if true, print Debug information." type:"bool" short:"d" env:"DEBUG"`
	Version   VersionFlag `name:"version" help:"Print version information and quit" short:"v" type:"counter"`
	VersionC  VersionCmd  `cmd:"" help:"Show the version information" name:"version" aliases:"ver"`
	versionFs embed.FS
}

type VersionFlag int
type VersionCmd struct{}

// https://github.com/golang/go/issues/41191
// https://stackoverflow.com/a/67357103/6309
//go:embed version/*
var versionFs embed.FS

func main() {

	var cli CLI
	ctx := kong.Parse(&cli)

	if cli.Debug {
		spew.Dump(cli)
		q.Q(cli)
		fmt.Printf("ctx command '%s'\n", ctx.Command())
	}

	cli.versionFs = versionFs

	if ctx.Command() != "version" && cli.Version > 0 {
		fmt.Printf(version.String(int(cli.Version), cli.versionFs))
		ctx.Exit(0)
	}

	fmt.Println("barerepo")
}
