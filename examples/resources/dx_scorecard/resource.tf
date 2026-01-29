terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.8.0"
    }
  }
}

provider "dx" {}

resource "dx_scorecard" "level_based_example" {
  name                           = "Terraform Provider Scorecard"
  description                    = "This is a test scorecard"
  type                           = "LEVEL"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
  evaluation_frequency_hours     = 2
  empty_level_label              = "Incomplete"
  empty_level_color              = "#cccccc"
  published                      = true

  tags = [
    { value = "test" },
    { value = "production" },
  ]

  levels = {
    bronze = {
      name  = "Bronze"
      color = "#FB923C"
      rank  = 1
    },
    silver = {
      name  = "Silver"
      color = "#9CA3AF"
      rank  = 2
    },
    gold = {
      name  = "Gold"
      color = "#FBBF24"
      rank  = 3
    },
  }

  checks = {
    test_check = {
      name                = "Test Check"
      scorecard_level_key = "bronze"
      ordering            = 0

      description    = "This is a test check"
      sql            = <<-EOT
        select 'PASS' as status, 123 as output
      EOT
      output_enabled = true
      output_type    = "custom"
      output_custom_options = {
        unit     = "widget"
        decimals = 0
      }
      output_aggregation = "median"
      external_url       = "http://example.com"
      published          = true
      estimated_dev_days = 1.5
    },

    another_check = {
      name                = "Another Check"
      scorecard_level_key = "bronze"
      ordering            = 1

      sql                = <<-EOT
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
      output_enabled     = true
      output_type        = "duration_seconds"
      output_aggregation = "median"
      published          = false
      estimated_dev_days = null
    },

    neat_silver_check = {
      name                = "Neat silver check"
      scorecard_level_key = "silver"
      ordering            = 0

      description        = "This is a neat silver check"
      sql                = <<-EOT
        select 'PASS' as status
      EOT
      output_enabled     = false
      published          = false
      estimated_dev_days = 1.5
    },
  }
}

resource "dx_scorecard" "points_based_example" {
  name                           = "Terraform Provider Scorecard - points"
  description                    = "This is a test scorecard"
  type                           = "POINTS"
  entity_filter_type             = "entity_types"
  entity_filter_type_identifiers = ["service"]
  evaluation_frequency_hours     = 2
  published                      = true

  check_groups = {
    group_1 = {
      name     = "First group"
      ordering = 0
    },
    group_2 = {
      name     = "Second group"
      ordering = 1
    },
  }

  checks = {
    check_1 = {
      name                      = "Check 1"
      scorecard_check_group_key = "group_1"
      ordering                  = 0

      description        = "This is a check in the first group"
      sql                = <<-EOT
        select 'PASS' as status
      EOT
      output_enabled     = false
      external_url       = "http://example.com"
      published          = true
      estimated_dev_days = 1.5
      points             = 10
    },

    check_2 = {
      name                      = "Check 2"
      scorecard_check_group_key = "group_2"
      ordering                  = 0

      sql                = <<-EOT
        select 'PASS' as status, 123 as output
      EOT
      output_enabled     = true
      output_type        = "duration_seconds"
      output_aggregation = "median"
      external_url       = "http://example.com"
      published          = true
      estimated_dev_days = 1.5
      points             = 20
    },
  }
}
