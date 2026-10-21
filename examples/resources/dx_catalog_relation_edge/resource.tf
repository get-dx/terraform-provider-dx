resource "dx_catalog_relation_edge" "login_frontend_depends_on_authn_backend" {
  relation_identifier      = dx_catalog_relation.service_depends_on_service.identifier
  source_entity_identifier = dx_entity.login_frontend.identifier
  target_entity_identifier = dx_entity.authn_backend.identifier
}
