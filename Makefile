# Load environment variables
include .env
export

# Project variables
PROJECT_NAME := air-social

# Include all makefiles from make/ directory
include make/*.mk

# Default target
.DEFAULT_GOAL := help

## Display this help message
.PHONY: help
help:
	@echo "Usage: make [target]"
	@echo ""
	@echo "Available targets:"
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z0-9_-]+:.*?## / {printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2}' $(MAKEFILE_LIST)