terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/local/dx"
      version = "0.1.0"
    }
  }
}

provider "dx" {
  api_token = "<your api token>"
}

resource "dx_scorecard" "example" {
  name                           = "Terraform Provider Scorecard"
  description                    = "This is a test scorecard"
  type                           = "LEVEL"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
  evaluation_frequency_hours     = 2
  empty_level_label              = "None"
  empty_level_color              = "#cccccc"
  published                      = true

  levels = [
    {
      key   = "bronze"
      name  = "Bronze"
      color = "#FB923C"
      rank  = 1
    },
    {
      key   = "silver"
      name  = "Silver"
      color = "#9CA3AF"
      rank  = 2
    },
    {
      key   = "gold"
      name  = "Gold"
      color = "#FBBF24"
      rank  = 3
    },
  ]
  checks = [
    {
      name                  = "Test Check"
      description           = "This is a test check"
      sql                   = "select 'PASS' as status, 123 as output"
      scorecard_level_key   = "bronze"
      output_enabled        = true
      output_type           = "duration_seconds"
      output_aggregation    = "median"
      ordering              = 0
      external_url          = "http://example.com"
      published             = true
      estimated_dev_days    = 1.5
      filter_message        = ""
      filter_sql            = ""
      output_custom_options = null
    }
  ]
}
