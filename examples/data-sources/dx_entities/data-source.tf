terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.8.0"
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

# Example 2: Access entity properties
# The properties field is JSON-encoded, use jsondecode() to access values
output "service_tiers" {
  description = "Tier property for each service"
  value = [
    for e in data.dx_entities.all_services.entities :
    e.properties != null ? jsondecode(e.properties)["tier"] : null
  ]
}
