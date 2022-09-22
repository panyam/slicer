# Slicer

# Installation

All instructions here assume you are in the root of the seenema source folder.

#### Install Docker:

Why is this needed - For running test PG DB.

* Install instructions here - https://docs.docker.com/desktop/install/mac-install/

## Setup a Virtualenv (for some tooling)

```
brew install python
```

Note: If you are in a *new* enough installation python2 may not be around so pip may fail.  If this is the case then first do

```
alias pip=pip3
```

```
# Need once
pip install --upgrade pip
pip install virtualenv
virtualenv venv
source venv/bin/activate
pip install --upgrade pip
pip install -r requirements.txt
```

Here on everything also assumes you are in this virtual environment

## Installing Dependencies

```
brew install golang goctl protobuf sqlite

# Optional: This is only the client if using pg as our DB
brew install pgadmin4
```

## Setup Postgres

### Build the pg docker image

A few common utility functions in/for golang
