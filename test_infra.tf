#Infrastructure to run deploy testing infrastructure

variable "doToken" {}



provider "digitalocean" {
  token = var.doToken
}
  

data "digitalocean_ssh_key" "keys" {
  name = "key"
}


#Droplet for LB
resource "digitalocean_droplet" "lb" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "lb"
  tags               = ["testing", "momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
}

#Droplets for Backend

resource "digitalocean_droplet" "b1" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "b1"
  tags               = ["testing", "momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
}

resource "digitalocean_droplet" "b2" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "b1"
  tags               = ["testing", "momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
}
