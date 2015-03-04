package main

import (
	"bytes"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"launchpad.net/gomaasapi"
)

const (
	envServerURL = "MAAS_SERVER_URL"
	envOAuthKey  = "MAAS_OAUTH_KEY"

	maasAPIVersion = "1.0"

	cmdUsage = `
Usage:

  maas-utils [-h] [-d] [-u <url>] [-o <oauth-key>] <command>

Accepted flags:

  -h
    Display this help information. Also supported: --help.
  -d
    Enable verbose output for debugging.

  -u <url>
    Required, unless the %s environment variable is set.
    <url> is the MAAS server URL (e.g. http://192.168.50.2/MAAS).

  -o <oauth-key>
    Required, unless the %s environment variable is set.
    <oauth-key> is needed to authenticate with the MAAS API.
    Expected format: 'xxx:yyy:zzz'.

Supported commands:

%s
`
)

// Flags used as arguments.
var (
	serverURL = flag.String("u",
		os.Getenv(envServerURL),
		fmt.Sprintf("MAAS server URL (or %s env var)", envServerURL),
	)
	oauthKey = flag.String("o",
		os.Getenv(envOAuthKey),
		fmt.Sprintf("MAAS OAuth key (or %s env var)", envOAuthKey),
	)
	debug = flag.Bool("d",
		false,
		"enable verbose output for debugging",
	)
)

// Supported subcommands.
var subcommands = map[string]string{
	"list-ips":    "Lists all statically allocated IP addresses",
	"release-ips": "Releases all statically allocated IP addresses",
}

func main() {
	// Silence the default output.
	out := bytes.NewBuffer(nil)
	flag.CommandLine.SetOutput(out)

	flag.Usage = func() {
		outStr := strings.TrimSuffix(out.String(), "\n")
		allArgs := strings.Join(flag.Args(), " ")
		switch {
		case flag.NArg() == 1:
			logf("unknown command: %s", flag.Arg(0))
		case outStr != "":
			logf(outStr)
		case flag.NArg() == 0:
			logf("no command specified.")
		default:
			logf("unrecognized argument(s): %s", allArgs)
		}

		var cmds []string
		ind := "  "
		for cmd, desc := range subcommands {
			cmds = append(cmds, ind+cmd+"\n"+ind+ind+desc+"\n")
		}
		sort.Strings(cmds)
		fmt.Printf(cmdUsage, envServerURL, envOAuthKey, strings.Join(cmds, "\n"))
		os.Exit(2)
	}

	flag.Parse()

	if _, ok := subcommands[flag.Arg(0)]; !ok || flag.NArg() != 1 {
		flag.Usage()
	}

	if *serverURL == "" {
		fatalf("MAAS server URL not specified.")
	}
	if *oauthKey == "" {
		fatalf("MAAS OAuth key not specified.")
	}

	switch flag.Arg(0) {
	case "list-ips":
		listIPs()
	case "release-ips":
		releaseIPs()
	}
}

func debugf(f string, a ...interface{}) {
	if !*debug {
		return
	}
	logf(f, a...)
}

func logf(f string, a ...interface{}) {
	cmd := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "%s: %s\n", cmd, fmt.Sprintf(f, a...))
}

func fatalf(f string, a ...interface{}) {
	logf(f, a...)
	fmt.Fprintln(os.Stderr)
	os.Exit(2)
}

func getClient() *gomaasapi.Client {
	client, err := gomaasapi.NewAuthenticatedClient(*serverURL, *oauthKey, maasAPIVersion)
	if err != nil {
		fatalf("cannot connect: %v", err)
	}
	debugf("connected to %q", *serverURL)
	return client
}
