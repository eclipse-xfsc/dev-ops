# Building Go services with plugin-based dependencies

Some library packages use plugin-driven architecture. The means that the library defines an interface which is used by 
dependent party, but exact implementation of the interface is dedicated to a separate piece of software, "plugin", which 
is loaded at runtime. It comes with a more complicated build step but improves encapsulation and makes it easier to 
switch between different implementations.

## Building using Docker

Here is the proposed [Dockerfile template](Dockerfile) that can be used to build services using such dependencies.
It expects 2 build arguments (```--build-arg```):
1. pluginRepoUrl - a path to git repository where chosen plugin is cloned from
2. pluginTag - a specific branch name or tag, default: main

## Building locally

Building a programme using plugin-based dependency implies two steps

1. Building plugin - See https://pkg.go.dev/cmd/go#hdr-Build_modes
2. Pasting a built .so file to the location, from  where it will be fetched by go tool - See https://pkg.go.dev/plugin
The expected location is implementation specific and should be defined by the library that loads a plugin.
