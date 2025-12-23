terraform {
  required_providers {
    dx = {
      source  = "registry.terraform.io/get-dx/dx"
      version = "~> 0.7.0"
    }
  }
}

provider "dx" {}

# Example: Create an entity type with properties and aliases
resource "dx_entity_type" "repository" {
  identifier  = "repository"
  name        = "Repository"
  description = "A source code repository"

  properties = {
    team = {
      name        = "Owning Team"
      description = "The team that owns this repository"
      type        = "multi_select"
      visibility  = "visible"
      ordering    = 0
      options = [
        { value = "platform", color = "#3b82f6" },
        { value = "data", color = "#ef4444" },
        { value = "product", color = "#10b981" },
        { value = "infrastructure", color = "#f59e0b" }
      ]
    }
    language = {
      name        = "Primary Language"
      description = "The main programming language used"
      type        = "text"
      visibility  = "visible"
      ordering    = 1
    }
    tier = {
      name       = "Service Tier"
      type       = "multi_select"
      visibility = "hidden"
      ordering   = 2
      options = [
        { value = "tier_1", color = "#818cf8" },
        { value = "tier_2", color = "#a78bfa" },
        { value = "tier_3", color = "#c084fc" }
      ]
    }
    active_entities_count = {
      name        = "Active Entities Count"
      description = "Number of active entities for this repository"
      type        = "computed"
      visibility  = "visible"
      ordering    = 3
      sql         = "SELECT COUNT(*) FROM portal_entities WHERE identifier = $entity_identifier"
      output_type = "number"
    }
    repository_url = {
      name                = "Repository URL"
      description         = "Link to the source code repository"
      type                = "url"
      visibility          = "visible"
      ordering            = 4
      call_to_action      = "View Repository"
      call_to_action_type = "text"
    }
  }

  aliases = {
    "github_repo" = true
  }
}
