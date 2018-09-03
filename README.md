# aws-session

`aws-session` sets temporary credential environment variables for a given profile.

### Features
* MFA Support 
* Assume Role Support

### Supported Shells
* sh 
* zsh
* fish


## Installation

#### homebrew

```bash
brew install qoomon/tab/aws-session
```

#### Manuel
`curl -o /usr/local/bin/aws-session https://raw.githubusercontent.com/qoomon/aws-session/master/aws-session`

## Configuration
Setup profiles you would like to assume in `~/.aws/config` => [aws cli-roles](https://docs.aws.amazon.com/cli/latest/userguide/cli-roles.html)


## Usage

`aws-session <profile>`

If the session requires MFA, you will be asked for the token

`assume assume` sets following environment variables and then executes the command 
* `AWS_PROFILE`
* `AWS_ACCESS_KEY_ID`
* `AWS_SECRET_ACCESS_KEY`
* `AWS_SESSION_TOKEN`
* `AWS_DEFAULT_REGION`

#### Examples

##### Print environment variables commands
`aws-session company-production`

## Utils
If you use `eval $(aws-session)` frequently, you may want to create a alias for it:

* zsh
```shell
alias aws-session='function(){eval "$(./aws-session $@)"}'
```
* bash
```shell
function aws-session { eval "$( $(which aws-session) "$@")"; }
```
