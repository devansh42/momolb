#Infrastructure to run deploy testing infrastructure

variable "doToken" {}

variable "bash_script" {
  type        = string
  description = "Bash Script to run on Provisioned vms"
  default     = <<EOT
    #cloud-config
    runcmd:

      - sudo apt update -y
      - sudo apt install -y python3-pip docker.io 
      - sudo docker pull devansh42/momo
      - sudo docker pull devansh42/momorunner
      - sudo mkdir -p /tmp/terraform/
      - sudo docker run -it devansh42/momorunner > /tmp/terraform/remoteState
      - sudo mkdir -p /momo
      - sudo cd /momo
      - git clone -b testing https://github.com/devansh42/momolb.git
      - cd momolb
      - sudo pip3 install docker
      - sudo python3 deploy_container.py 

  EOT
}




provider "digitalocean" {
  token = var.doToken
}


data "digitalocean_ssh_key" "keys" {
  name = "key"
}



#Droplets for Backend

resource "digitalocean_droplet" "b1" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "b1"
  tags               = ["backend","testing", "momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
  user_data          = var.bash_script
}

resource "digitalocean_droplet" "b2" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "b2"
  tags               = ["testing","backend" ,"momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
  user_data          = var.bash_script
}



#Droplet for LB
resource "digitalocean_droplet" "lb" {
  region             = "blr1"
  image              = "ubuntu-18-04-x64"
  size               = "s-1vcpu-1gb"
  name               = "lb"
  tags               = ["lb","testing", "momo"]
  monitoring         = "true"
  private_networking = "true"
  ssh_keys           = [data.digitalocean_ssh_key.keys.id]
  user_data          = var.bash_script
}





terraform {
  backend "remote" {
    organization = "Momo"

    workspaces {
      name = "momo-land"
    }
  }
}