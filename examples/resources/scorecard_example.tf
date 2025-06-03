terraform {
  required_providers {
    scorecard = {
      source  = "registry.terraform.io/local/scorecard"
      version = "0.1.0"
    }
  }
}

provider "scorecard" {
  api_token = "v1U37UDXfAHtABr7UJXaFdm5HDVqQPYFQ6Bo"
}

resource "scorecard_scorecard" "example" {
  name                           = "Terraform Provider Scorecard"
  description                    = "This is a test scorecard"
  type                           = "LEVEL"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
  evaluation_frequency_hours     = 2
  empty_level_label              = "None"
  empty_level_color              = "#cccccc"
  published                      = true

  levels = [{
    key   = "bronze"
    name  = "Bronze"
    color = "#cd7f32"
    rank  = 1
  }]
  checks = []
  # checks = [
  #   {
  #     name                  = "Test Check"
  #     description           = "This is a test check"
  #     sql                   = "select 'PASS' as status, 123 as output"
  #     scorecard_level_key   = "bronze"
  #     output_enabled        = true
  #     output_type           = "duration_seconds"
  #     output_aggregation    = "median"
  #     ordering              = 0
  #     external_url          = "http://example.com"
  #     published             = true
  #     estimated_dev_days    = 1
  #     filter_message        = ""
  #     filter_sql            = ""
  #     output_custom_options = ""
  #   }
  # ]
}
