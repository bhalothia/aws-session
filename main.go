package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"syscall"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
)

var (
	roleArnRegex = regexp.MustCompile(`^arn:aws:iam::(.+):role/([^/]+)(/.+)?$`)
)

func init() {
	flag.Usage = usage
}

func usage() {
	fmt.Fprintf(os.Stderr, "Usage: %s <profile/role_arn> [<options>] [<command> <args...>]\n", os.Args[0])
	flag.PrintDefaults()
}

func main() {
	var err error

	duration := *flag.Duration("duration",  stscreds.DefaultDuration, "The duration that temporary credentials will be valid for.")
	defaultRegion := *flag.String("region", "", "The aws default region.")
	mfaToken := *flag.String("token", "", "The mfa token to use. [only considered if assume by <role_arn>]")
	format := *flag.String("format", defaultEnvFormat(), "The environment variables format. [only considered if no <command> is provided]")
	flag.Parse()
	argv := flag.Args()
	
	if len(argv) < 1 {
		flag.Usage()
		fmt.Fprintf(os.Stderr, "\nERROR: <profile/role_arn> argument is missing\n")
		os.Exit(1)
	}

	stscreds.DefaultDuration = duration
	indentity := argv[0]
	command := argv[1:]

	var credentials *credentials.Value
	if roleArnRegex.MatchString(indentity) {
		credentials, err = assumeRole(indentity, mfaToken)
	} else {
		var profileRegion string
		credentials, profileRegion, err = assumeProfile(indentity)
		if defaultRegion == "" {
			defaultRegion = profileRegion
		}
	}
	exitOnError(err)

	if len(command) > 0 {
		err = executeWithSessionEnv(indentity, credentials, defaultRegion, command)
	} else {
		err = printSessionEnv(indentity, credentials, defaultRegion, format)
	}
	exitOnError(err)
}

func defaultEnvFormat() string {
	var shell = os.Getenv("SHELL")
	if shell == "" {
		if runtime.GOOS == "windows" {
			return "powershell"
		}
	} else if strings.HasSuffix(shell, "fish") {
		return "fish"
	}
	return "bash"
}

// creates temporary STS credentials for given profile in ~/.aws/config
// see https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html
func assumeProfile(profile string) (*credentials.Value, string, error) {
	sess := session.Must(session.NewSessionWithOptions(session.Options{
		Profile:                 profile,
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
	}))
	credentials, err := sess.Config.Credentials.Get()
	var region string
	if sess.Config.Region != nil {
		region = *sess.Config.Region
	}
	
	return &credentials, region, err
}

// creates temporary STS credentials for given role arn
func assumeRole(roleArn, token string) (*credentials.Value, error) {
	sess := session.Must(session.NewSession())
	credentials, err := stscreds.NewCredentials(sess, roleArn, func(p *stscreds.AssumeRoleProvider) {
		p.SerialNumber = aws.String(token)
		p.TokenProvider = stscreds.StdinTokenProvider
	}).Get()
	
	return &credentials, err
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

func executeWithSessionEnv(identity string, credentials *credentials.Value, region string, command []string) error {
	os.Setenv("AWS_IDENTITY", identity)

	os.Setenv("AWS_ACCESS_KEY_ID", credentials.AccessKeyID)
	os.Setenv("AWS_SECRET_ACCESS_KEY", credentials.SecretAccessKey)

	if credentials.SessionToken != "" {
		os.Setenv("AWS_SESSION_TOKEN", credentials.SessionToken)
		os.Setenv("AWS_SECURITY_TOKEN", credentials.SessionToken)
	}

	if region != "" {
		os.Setenv("AWS_DEFAULT_REGION", region)
		os.Setenv("AWS_REGION", region)
	}

	path, err := exec.LookPath(command[0])
	if err != nil {
		return err
	}
	return syscall.Exec(path, command, os.Environ())
}

// prints the credentials in a way that can easily be sourced with bash.
func printSessionEnv(identity string, credentials *credentials.Value, region string, format string) error {
	switch format {
	case "bash":
		printSessionEnvBash(identity, credentials, region)
	case "fish":
		printSessionEnvFish(identity, credentials, region)
	case "powershell":
		printSessionEnvPowerShell(identity, credentials, region)
	default:
		return errors.New(fmt.Sprintf("unsuported format '%s'", format))
	}
	return nil
}

func printSessionEnvBash(identity string, credentials *credentials.Value, region string) {
	fmt.Printf("export AWS_IDENTITY=\"%s\"\n", identity)

	fmt.Printf("export AWS_ACCESS_KEY_ID=\"%s\"\n", credentials.AccessKeyID)
	fmt.Printf("export AWS_SECRET_ACCESS_KEY=\"%s\"\n", credentials.SecretAccessKey)

	if credentials.SessionToken != "" {
		fmt.Printf("export AWS_SESSION_TOKEN=\"%s\"\n", credentials.SessionToken)
		fmt.Printf("export AWS_SECURITY_TOKEN=\"%s\"\n", credentials.SessionToken)
	}

	if region != "" {
		fmt.Printf("export AWS_DEFAULT_REGION=\"%s\"\n", region)
		fmt.Printf("export AWS_REGION=\"%s\"\n", region)
	}

	fmt.Printf("\n")
	fmt.Printf("# Run this to configure your shell:\n")
	fmt.Printf("# eval $(%s)\n", strings.Join(os.Args, " "))
}

func printSessionEnvFish(identity string, credentials *credentials.Value, region string) {
	fmt.Printf("set -gx AWS_IDENTITY \"%s\";\n", identity)

	fmt.Printf("set -gx AWS_ACCESS_KEY_ID \"%s\";\n", credentials.AccessKeyID)
	fmt.Printf("set -gx AWS_SECRET_ACCESS_KEY \"%s\";\n", credentials.SecretAccessKey)

	if credentials.SessionToken != "" {
		fmt.Printf("set -gx AWS_SESSION_TOKEN \"%s\";\n", credentials.SessionToken)
		fmt.Printf("set -gx AWS_SECURITY_TOKEN \"%s\";\n", credentials.SessionToken)
	}

	if region != "" {
		fmt.Printf("set -gx AWS_DEFAULT_REGION \"%s\";\n", region)
		fmt.Printf("set -gx AWS_REGION \"%s\";\n", region)
	}

	fmt.Printf("\n")
	fmt.Printf("# Run this to configure your shell:\n")
	fmt.Printf("# eval (%s)\n", strings.Join(os.Args, " "))
}

func printSessionEnvPowerShell(identity string, credentials *credentials.Value, region string) {
	fmt.Printf("$env:AWS_IDENTITY=\"%s\"\n", identity)

	fmt.Printf("$env:AWS_ACCESS_KEY_ID=\"%s\"\n", credentials.AccessKeyID)
	fmt.Printf("$env:AWS_SECRET_ACCESS_KEY=\"%s\"\n", credentials.SecretAccessKey)

	if credentials.SessionToken != "" {
		fmt.Printf("$env:AWS_SESSION_TOKEN=\"%s\"\n", credentials.SessionToken)
		fmt.Printf("$env:AWS_SECURITY_TOKEN=\"%s\"\n", credentials.SessionToken)
	}

	if region != "" {
		fmt.Printf("$env:AWS_DEFAULT_REGION=\"%s\"\n", region)
		fmt.Printf("$env:AWS_REGION=\"%s\"\n", region)
	}

	fmt.Printf("\n")
	fmt.Printf("# Run this to configure your shell:\n")
	fmt.Printf("# %s | Invoke-Expression \n", strings.Join(os.Args, " "))
}
