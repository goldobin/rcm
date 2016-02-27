# Tasks

Check node's process is actually running in `Node::IsUp` 

Add logs output for each node

Add flag to `create` command which will signalize to perform `start` and `distribute-slots` with default parameters (or 
parameters specified in env vars) right after cluster start

Add `cli-each` command which will perform the same operation on each known cluster node 

Add master/slave count info to the `list` output

Investigate the ways of distributing application (brew, rpm, deb)

Add reasonable ENV vars to override defaults

Add `-y` (answer yes) flag (for create, remove, damage etc.)

Add specific Redis version download (no redis pre-installation)

Generate and test bash completion
