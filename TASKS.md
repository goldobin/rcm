# Tasks

Add `README.md` file

Check node's process is actually running, and executable is a redis-server in `Node::IsUp` 

Refactor commands: move to separate files, use factory

Add flag to `create` command which will signalize to perform `start` and `distribute-slots` with default parameters (or 
parameters specified in env vars) right after cluster start

Add `cli-each` command which will perform the same operation on each known cluster node 

Add master/slave count info to the `list` output

Implement `damage` command
 
Implement `recover` command

Generate and bash completion

Move to git-hub

Enable CI (travis ci?)

Investigate the ways of distributing application (brew, rpm, deb)

Add reasonable ENV vars to the commands

Add specific redis version download (no redis pre-installation)

Add global 'answer yes' flag
