#!/bin/bash

python3 -c 'import json,pprint,sys; [(pprint.pprint(json.loads(line)),print(),sys.stdout.flush()) for line in sys.stdin if line.startswith("{")]'
