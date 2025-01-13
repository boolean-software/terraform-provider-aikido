terraform {
  required_providers {
    aikido = {
      source = "github.com/boolean-software/aikido"
    }
  }
}

provider "aikido" {}

data "aikido_users" "all" {}

output "out" {
  value       = data.aikido_users.all
}
