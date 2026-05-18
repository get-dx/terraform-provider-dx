resource "dx_catalog_relation" "team_manages_github_team" {
  identifier                    = "team-manages-github-team"
  type                          = "manages"
  cardinality                   = "one_to_many"
  source_entity_type_identifier = "team"
  target_entity_type_identifier = "github-team"
  description                   = "Maps DX teams to their GitHub team"
}
