package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <profile> [<options>] [<command> <args...>]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	flag_region := *flag.String("region", "", "The AWS region. Overrides region of profile definition.")
	flag_duration := *flag.Duration("duration",  stscreds.DefaultDuration, "The duration in that temporary credentials will be valid for. (default " + stscreds.DefaultDuration.String() + ")")
	flag_format := *flag.String("format", defaultEnvFormat(), "The environment variables format. [only considered if no <command> is provided]")
	flag.Parse()
	flag_args:= flag.Args()
	
	if len(flag_args) < 1 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nERROR: <profile> argument is missing\n")
		os.Exit(1)
	}

	profile := flag_args[0]
	command := flag_args[1:]
	stscreds.DefaultDuration = flag_duration

	sess, err:= createSession(profile)
	exitOnError(err)
	
	credentials, err := sess.Config.Credentials.Get()
	exitOnError(err)
	
	var defaultRegion string
	if sess.Config.Region != nil {
		defaultRegion = *sess.Config.Region
	}
	if flag_region != "" {
		defaultRegion = flag_region
	}	

	if len(command) > 0 {
		err := executeWithAwsEnv(credentials, defaultRegion, command)
		exitOnError(err)
	} else {
		err := printAwsEnv(credentials, defaultRegion, flag_format)
		exitOnError(err)
		fmt.Println("")
		fmt.Println("# Run this to configure your shell:")
		fmt.Println("# " + evalCommand(flag_format, strings.Join(os.Args, " ")))
	}
}

// determine default environment format based on curret shell. [default sh]
func defaultEnvFormat() string {
	shell := os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			return "powershell"
		}
	} else if strings.HasSuffix(shell, "fish") {
		return "fish"
	}
	return "sh"
}

// creates temporary STS credentials for given profile in ~/.aws/config
// see https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html
func createSession(profile string) (*session.Session, error) {
	os.Unsetenv("AWS_ACCESS_KEY_ID")
	os.Unsetenv("AWS_SECRET_ACCESS_KEY")
	
	return session.NewSessionWithOptions(session.Options{
		Profile:                 profile,
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider, // TODO make use of https://github.com/mattn/go-tty
	})
}

func exitOnError(err error) {
	if err != nil {
		if _, ok := err.(*exec.ExitError); ! ok {
			// Errors are not already on Stderr.
			fmt.Fprintf(os.Stderr, "\nERROR: %v\n", err)
		}
		os.Exit(1)
	}
}

func executeWithAwsEnv(credentials credentials.Value, defaultRegion string, command []string) error {
	
	os.Setenv("AWS_ACCESS_KEY_ID", credentials.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey)

	if credentials.SessionToken != "" {
		os.Setenv("AWS_SESSION_TOKEN", credentials.SessionToken)
		os.Setenv("AWS_SECURITY_TOKEN", credentials.SessionToken)
	}

	if defaultRegion != "" {
		os.Setenv("AWS_DEFAULT_REGION", defaultRegion)
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		return err
	}
	
	err = syscall.Exec(path, command, os.Environ())
	if err != nil {
		return err
	}
	
	return nil
}

// prints the credentials in a way that can easily be sourced.
func printAwsEnv(credentials credentials.Value, defaultRegion string, format string) error {

	fmt.Println(envCommand(format, "AWS_ACCESS_KEY_ID", credentials.AccessKeyID))
	fmt.Println(envCommand(format, "AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey))

	if credentials.SessionToken != "" {
		fmt.Println(envCommand(format, "AWS_SESSION_TOKEN", credentials.SessionToken))
		fmt.Println(envCommand(format, "AWS_SECURITY_TOKEN", credentials.SessionToken))
	}

	if defaultRegion != "" {
		fmt.Println(envCommand(format, "AWS_DEFAULT_REGION", defaultRegion))
		fmt.Println(envCommand(format, "AWS_REGION", defaultRegion))
	}
	
	return nil
}

func envCommand(format string, name string, value string) string {
	switch format {
	case "fish":
		return "set -gx " + name + " \"" + value + "\";"
	case "powershell":
		return "$env:" + name + "=\"" + value + "\";"
	default: // sh
		return "export " + name + "=\"" + value + "\";"
	}
}

func evalCommand(format string, command string) string {
	switch format {
	case "fish":
		return "eval (" + command + ")"
	case "powershell":
		return command + " | Invoke-Expression"
	default: // sh
		return "eval $(" + command + ")"
	}
}
