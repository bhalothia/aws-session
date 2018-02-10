# aws-assume

`aws-assume` sets temporary credentials environment variables for a given profile or role.

This project is heavily inspired by [assume-role](https://github.com/remind101/assume-role), so kudos to [remind101](https://github.com/remind101)

## Installation

#### homebrew

```bash
brew install qoomon/tab/aws-assume
```

#### Go

```bash
go get -u github.com/qoomon/aws-assume
```

#### Scoop

```bash
scoop bucket add qoomon https://github.com/qoomon/scoop-bucket.git
scoop install aws-assume
```

#### Manuel

Download from [Releases](https://github.com/qoomon/aws-assume/releases/latest)
* [macOS](https://github.com/qoomon/aws-assume/releases)
* [Linux](https://github.com/qoomon/aws-assume/releases)
* [Windows](https://github.com/qoomon/aws-assume/releases)

## Configuration

Setup profiles you would like to assume in `~/.aws/config` => [aws cli-roles](https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html)


## Usage

`aws-assume <profile/role_arn> [<options>] [<command> <args...>]`


If the role requires MFA, you will be asked for the token

*Options*
* `-duration` - The duration that temporary credentials will be valid for. (default 15m0s)
* `-format` - The environment variables format. [only consider if no \<command> is provided]
  * bash 
  * fish
  * powershell
* `-region` - The AWS default region.
* `-token` - The MFA token to use. [only considered if assume by \<role_arn>]


`assume assume` sets following environment variables and then executes the command 
* `AWS_IDENTITY` 
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `AWS_SESSION_TOKEN`
* `AWS_SECURITY_TOKEN`
* `AWS_DEFAULT_REGION`
* `AWS_REGION`

#### Examples

##### Execute command with environment variables
`aws-assume company-production aws ec2 describe-instances`

##### Print environment variables commands
`aws-assume company-production`

## Utils
If you use `eval $(aws-assume)` frequently, you may want to create a alias for it:

* zsh
```shell
alias aws-assume='function(){eval $(command aws-assume $1) && shift; $@}'

```
* bash
```shell
function aws-assume { eval $( $(which aws-assume) $1) && shift; $@ }
```

## TODO

* [ ] Cache credentials.
