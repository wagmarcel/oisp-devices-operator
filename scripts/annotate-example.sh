kubectl annotate --overwrite node 18dc19a79fd0 oisp.org/deviceSpec=[{"id":"1234567890","name":"testdevice","hostPath":"/dev/ttyS1","containerPath":"/dev/ttyTest","permission":"rw","gid":1003}]
