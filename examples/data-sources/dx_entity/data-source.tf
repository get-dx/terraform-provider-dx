terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.7.0"
    }
  }
}

provider "dx" {}

# Example 1: Look up an existing entity by identifier
data "dx_entity" "payment_service" {
  identifier = "payment-service"
}

# Example 2: Use data source attributes to create related resources
data "dx_entity" "core_service" {
  identifier = "core-service"
}

# You can reference any attribute from the data source
output "service_name" {
  description = "The name of the core service"
  value       = data.dx_entity.core_service.name
}

output "service_type" {
  description = "The type of the core service"
  value       = data.dx_entity.core_service.type
}

output "service_domain" {
  description = "The domain of the core service"
  value       = data.dx_entity.core_service.domain
}

output "service_owner_teams" {
  description = "The owner teams with id and name"
  value       = data.dx_entity.core_service.owner_teams
}

output "service_properties" {
  description = "All service properties"
  value       = data.dx_entity.core_service.properties
}

output "service_aliases" {
  description = "All service aliases"
  value       = data.dx_entity.core_service.aliases
}

# Example 3: Use data source in conditional logic
data "dx_entity" "legacy_api" {
  identifier = "legacy-api"
}

# You can use the data source to make decisions
locals {
  is_production_ready = try(
    data.dx_entity.legacy_api.properties.environment == "production",
    false
  )
}

output "production_status" {
  value = local.is_production_ready
}



