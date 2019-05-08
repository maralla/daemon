package main

import (
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"
	"text/template"
)

const defaultSystemdDir = "/usr/lib/systemd/system"
const tpl = `[Unit]
Description={{ .Description }}
Wants=network-online.target
After={{ .After }}

[Service]
User=www-data
Group=www-data
Type=simple
Restart=on-failure
ExecStart={{ .Command }}
WorkingDirectory={{ .WorkDir }}

[Install]
WantedBy=multi-user.target
`

type cmdArgs struct {
	Help        bool
	Name        string
	Description string
	WorkDir     string
	Command     string
	SystemdDir  string
	After       string
}

func initArgs() (args cmdArgs, err error) {
	dir := os.Getenv("SYSTEMD_CONFIG_DIR")
	if dir == "" {
		dir = defaultSystemdDir
	}
	flag.StringVar(&args.Name, "name", "", "Name of the daemon")
	flag.StringVar(&args.Description, "d", "", "Daemon description")
	flag.StringVar(&args.WorkDir, "wd", "", "Working directory")
	flag.StringVar(&args.Command, "cmd", "", "The command to run")
	flag.StringVar(&args.After, "after", "", "Systemd After")
	flag.StringVar(&args.SystemdDir, "systemd-config-dir", dir, "Systemd config directory")
	flag.BoolVar(&args.Help, "h", false, "Show help")
	flag.Parse()
	if args.Help {
		flag.Usage()
		os.Exit(0)
	}
	if args.Name == "" {
		err = errors.New("Argument name is required")
	}
	if args.Command == "" {
		err = errors.New("Argument cmd is required")
	}
	args.After = strings.TrimSpace("network-online.target " + args.After)
	return
}

func die(err error) {
	fmt.Fprintf(os.Stderr, "%s\n", err)
	os.Exit(1)
}

func main() {
	args, err := initArgs()
	if err != nil {
		die(err)
	}
	run(args)
}

func run(args cmdArgs) {
	t := template.Must(template.New("systemd-config").Parse(tpl))
	buf := bytes.NewBuffer([]byte{})
	if err := t.Execute(buf, args); err != nil {
		die(err)
	}
	file := path.Join(args.SystemdDir, args.Name+".service")
	if err := ioutil.WriteFile(file, buf.Bytes(), 0644); err != nil {
		die(err)
	}
}
