terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.7.0"
    }
  }
}

provider "dx" {}

# Example 1: List all services
data "dx_entities" "all_services" {
  type = "service"
}

output "service_count" {
  description = "Total number of services in the catalog"
  value       = length(data.dx_entities.all_services.entities)
}

output "service_identifiers" {
  description = "All service identifiers"
  value       = [for e in data.dx_entities.all_services.entities : e.identifier]
}

# Example 2: Filter entities using for expressions
data "dx_entities" "all_apis" {
  type = "api"
}

output "api_names" {
  description = "Names of all APIs"
  value       = [for e in data.dx_entities.all_apis.entities : e.name]
}
