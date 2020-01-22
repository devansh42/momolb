#!/bin/sh
#Script to run lb
docker run -it -p 80:80 devansh42/momo /
	-lb=true /
        -backend="b1:134.209.157.47:8080;b2:134.209.157.51:8080;" /
        -health="method=tcp;port=8080;timeout=30;interval=300;threshold=0.6"

