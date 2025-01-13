terraform {
  required_providers {
    aikido = {
      source = "github.com/boolean-software/aikido"
    }
  }
}

provider "aikido" {}

data "aikido_example" "all" {}