# Introduction

Provides an small docker image which contains basic things for cloning, executing python scripts and helm. This container is mostly used to execute steps on  repositories during the build. The reason for this is, that the gitlab standard ci has some package not, and the most base images missing some thing to execute the required things ,so this new image closes the gap in the build.

# Depedencies

Depends on the python:3.11-alpine base image


# Bootstrap

It's used within ci execution by choosing this image: 

```
mystep:
  stage: build
  image: 
    name: node-654e3bca7fbeeed18f81d7c7.ps-xaas.io/dev-ops/build-executor:latest
  script:
    - echo "execute things now"
    - git clone https://myrepo
    - pip install -r requirements.txt
    - echo "Start Script"
    - python ./myscriptPy.py
    - curl things
    - make something
    - mvn packages
    - openssl things

```

# Developer Information

If anything is missing in the build CI just add more packages by using apk to the docker image section: 

```
RUN apk update
RUN apk add bash
RUN apk add git
RUN apk add curl
RUN apk add openssl
RUN apk add maven
RUN apk add make
RUN apk add build-base
```


