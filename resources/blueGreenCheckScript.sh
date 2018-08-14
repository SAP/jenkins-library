#!/usr/bin/env bash
# this is simply testing if the application root returns HTTP STATUS_CODE
curl -so /dev/null -w '%{response_code}' https://$1 | grep $STATUS_CODE
