terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.6.0"
    }
  }
}

provider "dx" {}

# Example 1: Create a service entity with properties and aliases
resource "dx_entity" "payment_service" {
  identifier     = "payment-service"
  type           = "service"
  name           = "Payment Service"
  description    = "Core payment processing service handling all payment transactions"
  owner_team_ids = ["MzI1NTk"]
  owner_user_ids = ["MQ"]
  domain         = "payments"

  properties = {
    service_tier      = "Tier-1"
    language          = ["Go", "Python"]
    slack_channel_url = "https://slack.com/channels/payment-service"
    architecture      = "microservices"
    deployment_env    = ["production", "staging"]
  }

  aliases = {
    github_repo = [
      {
        identifier = "1234567890"
      }
    ]
    pagerduty_service = [
      {
        identifier = "PD12345"
      }
    ]
  }
}

# Example 2: Create an API entity with minimal configuration
resource "dx_entity" "user_api" {
  identifier     = "user-api"
  type           = "api"
  name           = "User Management API"
  description    = "RESTful API for user management operations"
  owner_team_ids = ["MzI1NTk"]
  domain         = "platform"

  properties = {
    api_version   = "v2"
    protocol      = "REST"
    documentation = "https://api.company.com/docs/user-api"
    rate_limit    = "1000/hour"
  }

  aliases = {
    github_repo = [
      {
        identifier = "962275774"
      }
    ]
  }
}

