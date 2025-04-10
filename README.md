# Introduction

This repo contains common Helm Charts, CI-Templates and more for XFSC which can be used in the associated projects. E.g. Library Charts.


# Charts

|Name|Purpose|Folder|
|----|-------|------|
|Library| This chart is a library chart which contains helpers for building helm charts.| [Click](/library/deployment/helm) |
|Universal Resolver| This chart creates an deployment for the universal resolver from uport.| [Click](/universalresolver/deployment/helm) |
|ipfs-cluster| This chart creates an deployment for a small ipfs cluster.| [Click](/ipfs-cluster/deployment/helm) |
|Hashicorp Vault| Deployment for a simple vault setup.| [Click](/hashicorp-vault/deployment/helm) |

# Test of Workflows

For testing the workflows, act is used: https://github.com/nektos/act