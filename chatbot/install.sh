#!/bin/bash

# REQ: 
# - brew install pyenv
# - xcode-select --install (or use the way over https://developer.apple.com/download/more/ )
# Command line Tools of Xcode is needed to compile python 3.x (mac)

# Install Python over pyenv
pyenv install 3.6.5
pyenv global 3.6.5

pip3 install --upgrade pip setuptools wheel

pip3 install -U -r requirements.txt