terraform {
  required_providers {
    aikido = {
      source = "github.com/boolean-software/aikido"
    }
  }
}

provider "aikido" {}

resource "aikido_team" "devs" {
    name = "developers"
    responsibilities = [{
      id = 620493
      type = "code_repository"
    }]
}

output "out" {
  value       = aikido_team.devs
}