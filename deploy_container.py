#!/usr/bin/python3

import json  # for parsing state file
import docker  # for deployment purposes
import socket  # for hostname retrival

#hostname of the current system
hostname = socket.gethostname()



# Parses statefile and returns State Data

def parse_state_file(file):
    with open(file) as f:
        v = json.load(f)
        return v

# Returns terraform managed resources


def get_resources(stateObject):
    return stateObject["resources"]


# Retrives Droplets array from state file
def get_droplets(res):
    droplets = []
    for x in res:
        if x["mode"] == "managed" and x["type"] == "digitalocean_droplet":
            for y in x["instances"]:
                attr = y["attributes"]
                droplets.append((x["name"], attr))
    return droplets


# This retrives droplets for backend stuff
def retrive_backend_droplets(droplets):
    return droplets_with_tag(droplets, "backend")


# This retrives droplets for load balancer stuff
def retrive_lb_droplets(droplets):
    return droplets_with_tag(droplets, "lb")


# This retrives droplets with given tag
def droplets_with_tag(droplets, tag):
    return [(name, attr) for (name, attr) in droplets if tag in attr["tags"]]


# Deploys Backend container of load balancer
# backend is a tuple of (name,attr) where attr stands for Attribute
def deploy_backend(backend):
    name, attr = backend
    c = docker.from_env()
    print("Pulling Docker Image")
    c.images.pull("devansh42/momo", tag="latest")
    print("Image Pulled\nDeploying Container")
    container = c.containers.run(
        "devansh42/momo", command="-port=8000 -alsologtostderr", ports={8000: 8080}, detach=True)
    print("Container Deployed in detach mode")
    c.close()


# This deploys lb container with given backend
def deploy_load_balancer(lb, backends):
    backend_str = ";".join(map(prepare_backend_string, backends))
    health_check_str = prepare_backend_string()
    name, attr = lb
    c = docker.from_env()
    print("Pulling Docker Image")
    c.images.pull("devansh42/momo", tag="latest")
    print("Image Pulled\nDeploying Container")
    container = c.containers.run(
        "devansh42/momo",
        command=["-lb",
                 "-port=8000",
                 "-backend=%s" % (backend_str),
                 "-alsologtostderr",
                 "-health=%s" % (health_check_str)
                 ],
        detach=True, ports={8000: 80})
    print("Container Deployed in detach mode")


def prepare_health_check_str():
    d = {"method": "tcp",
         "port": "8000",  # Port to make health check request
         "timeout": "60s",
         "threshold": 0.7,
         }
    ar = ["%s=%s" % (x, d[x]) for x in d.keys()]
    return ";".join(ar)


def prepare_backend_string(backend):
    name, attr = backend
    return "%s:%s:%s" % (name, attr["ipv4_address"], "8080")


def do_for_lb(lb,droplets):
    print("Deploying lb")
    deploy_load_balancer(lb,droplets)

def do_for_backend(droplets):
    for (name,attr) in droplets:
        if name is hostname:
            deploy_backend((name,attr)) #deploy this node
            break    




# File to retirve terraform infra state
filename = "/tmp/terraform/remoteState"

droplets = get_droplets(get_resources(parse_state_file(filename)))
backends=retrive_backend_droplets(droplets)


#Below written code identifies current node type and 

if hostname is "lb":
    do_for_lb(retrive_lb_droplets(droplets)[0],backends)
else:
    do_for_backend(backends)
