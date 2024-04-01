# Lsportal(go)
A language server wrapper for running a langauge server on a subset of another language 

## Examples

The examples folder can be used as a demonstation of this function within go template files. The `.helix/langauges.toml` has a config for running lsportal within that file, to use it:

- install helix,go,vscode-html-language-server,tailwindcss-language-server
- make lsportal available in your PATH
- run helix and try writing html in `example/main.go`

## Setup
To understand using lsportal I'll dissect the example config, this should be easy to adapt to neovim, emacs or whatever other lsp supporting editor you use. 
```toml
[language-server]
#We define a new language server that runs lsportal
html-templ= { command = "lsportal-go",
 args = 
['--exclusion', '({{[\s\S]*?}})', # We exclude the templating parts eg:{{.Count}}
'html', #set the file extension the lang server expects
'htmlT[\n\s]*?\([\n\s]*?`([\s\S]*?)`[\s\n\,]*?\)', #A regex to match the region that should be sent to the server, in this case: " htmlT(``) "
"vscode-html-language-server", #command to run our language server
"--","--stdio"],# anything after -- are args used when running the server
 config = {  hostInfo="helix" } }

[[language]]
name = "go"
language-servers = ["gopls","html-templ","tailwindcss-ls" ] # Then add this server to the list that will be run for golang

````
