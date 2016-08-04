# ardielle-go [![Build Status](https://travis-ci.org/ardielle/ardielle-go.svg?branch=master)](https://travis-ci.org/ardielle/ardielle-go)

Go language support for RDL.

## Regenerating models from RDL

The file [rdl.rdl](https://github.com/ardielle/ardielle-common/blob/master/rdl.rdl) must be available to regenerate the sources. Assuming the clone directory for ardielle-common is specified by the environment variable `ARDIELLE_COMMON`:

    rdl -sp generate -te -o /tmp go-model $ARDIELLE_COMMON/rdl.rdl
	cp -p /tmp/rdl_model.go rdl/schema.go
	cp -p /tmp/rdl_schema.go rdl/rdl_schema.go
	go fmt rdl/schema.go rdl/rdl_schema.go



Godoc API references:  
* https://godoc.org/github.com/ardielle/ardielle-go/rdl
* https://godoc.org/github.com/ardielle/ardielle-go/tbin  

## License

Copyright 2015 Yahoo Inc.

Licensed under the terms of the Apache version 2.0 license. See LICENSE file for terms.
