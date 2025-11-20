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
      options = [
        { value = "platform", color = "#3b82f6" },
        { value = "data", color = "#ef4444" },
        { value = "product", color = "#10b981" },
        { value = "infrastructure", color = "#f59e0b" }
      ]
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
      options = [
        { value = "tier_1", color = "#818cf8" },
        { value = "tier_2", color = "#a78bfa" },
        { value = "tier_3", color = "#c084fc" }
      ]
    }
  ]

  aliases = {
    "github_repo" = true
  }
}
