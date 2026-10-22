terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.11.0"
    }
  }
}

provider "dx" {}

resource "dx_catalog_relation" "service_depends_on_resource" {
  identifier                    = "service-depends-on-resource"
  type                          = "depends on"
  cardinality                   = "many_to_many"
  source_entity_type_identifier = "service"
  target_entity_type_identifier = "resource"
  description                   = "Maps a service to resources it depends on to operate correctly"
}

resource "dx_catalog_relation_edge" "authn_service_depends_on_authn_service_core_db" {
  relation_identifier      = "service-depends-on-resource"
  source_entity_identifier = "authn-service"
  target_entity_identifier = "authn-service-core-db"
}

resource "dx_catalog_relation_edge" "authn_service_depends_on_authn_service_event_publisher" {
  relation_identifier      = "service-depends-on-resource"
  source_entity_identifier = "authn-service"
  target_entity_identifier = "authn-service-event-publisher"
}
