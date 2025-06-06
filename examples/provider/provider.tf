terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/local/dx"
      version = "0.1.0"
    }
  }
}

provider "dx" {
  # Define your Web API token here, or set `DX_WEB_API_TOKEN` in your environment.
  #
  # To manage scorecards, the token must have the following scopes:
  #
  # - scorecards:read
  # - scorecards:write
  #
  api_token = "<your api token>"
}

resource "dx_scorecard" "my_example_scorecard" {
  name                           = "My example scorecard"
  description                    = "This is a test scorecard"
  type                           = "LEVEL"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
  evaluation_frequency_hours     = 2
  empty_level_label              = "Incomplete"
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
      name                = "Test Check"
      scorecard_level_key = "bronze"
      ordering            = 0

      description           = "This is a test check"
      sql                   = <<-EOT
        select
          'PASS' as status,
          123 as output
      EOT
      output_enabled        = true
      output_type           = "duration_seconds"
      output_aggregation    = "median"
      external_url          = "http://example.com"
      published             = true
      estimated_dev_days    = 1.5
      filter_message        = ""
      filter_sql            = ""
      output_custom_options = null
    },
    {
      name                = "Another Check"
      scorecard_level_key = "silver"
      ordering            = 0

      description           = "This is a another test check"
      sql                   = <<-EOT
        with random_number as (
          select ROUND(RANDOM() * 10) as value
        )
        select case
            when value >= 7 then 'PASS'
            when value >= 4 then 'WARN'
            else 'FAIL'
          end as status,
          value as output
        from random_number
      EOT
      output_enabled        = true
      output_type           = "duration_seconds"
      output_aggregation    = "median"
      external_url          = "http://example.com"
      published             = false
      estimated_dev_days    = null
      filter_message        = ""
      filter_sql            = ""
      output_custom_options = null
    }
  ]
}
