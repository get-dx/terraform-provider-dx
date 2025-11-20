# Example: Create a basic entity type
resource "dx_entity_type" "service" {
  identifier  = "service"
  name        = "Service"
  description = "A deployable service in our infrastructure"
}

# Example: Create an entity type with properties and aliases
resource "dx_entity_type" "repository" {
  identifier  = "repository"
  name        = "Repository"
  description = "A source code repository"

  properties = [
    {
      identifier  = "team"
      name        = "Owning Team"
      description = "The team that owns this repository"
      type        = "multi_select"
      visibility  = "visible"
      ordering    = 0
      options     = ["platform", "data", "product", "infrastructure"]
    },
    {
      identifier  = "language"
      name        = "Primary Language"
      description = "The main programming language used"
      type        = "text"
      visibility  = "visible"
      ordering    = 1
    },
    {
      identifier = "tier"
      name       = "Service Tier"
      type       = "multi_select"
      visibility = "hidden"
      ordering   = 2
      options    = ["tier_1", "tier_2", "tier_3"]
    }
  ]

  aliases = {
    "github_repository" = true
    "gitlab_project"    = true
  }
}

# Example: Minimal entity type with just required fields
resource "dx_entity_type" "team" {
  identifier = "team"
  name       = "Team"
}
