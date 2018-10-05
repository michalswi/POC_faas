#!/usr/bin/env python

import json

def test():
    # return(json.dumps([1,2,3,4,5,6]))
    return(json.dumps({"person":[{"name":"Jan","surname":"Kowalski"}]}))

print(test())
