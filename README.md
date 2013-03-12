goh
===

golang client of hbase (via thrift)

this project is suspend, thrift4go has some bugs, need to fix them first, or figure out another approach.

Install
===

The official thrift package is out of date, need install [thrift4go](https://github.com/pomack/thrift4go) first,
then

	go get github.com/sdming/goh

Usage
===

	host, port := "192.168.17.129", "9090"
	
	client, err := goh.NewTcpClient(host, port, goh.TBinaryProtocol, false)
	if err != nil {
		fmt.Println(err)
		return
	}

	if err = client.Open(); err != nil {
		fmt.Println(err)
		return
	}

	defer client.Close()

	fmt.Println(client.IsTableEnabled(table))
	fmt.Println(client.DisableTable(table))
	fmt.Println(client.EnableTable(table))
	fmt.Println(client.Compact(table))
	

\demo\client.go for more example	

Files
===

* \thrift  
  thrift4go fork 

* \Hbase  
  generated by thrift compiler

  thrift --gen go hbase-root/src/main/resources/org/apache/hadoop/hbase/thrift/Hbase.thrift

* \demo  
  demo code of goh usage  


Start/Stop thrift 
===

	./bin/hbase-daemon.sh start thrift

	./bin/hbase-daemon.sh stop thrift



Links
===

* http://wiki.apache.org/hadoop/Hbase/ThriftApi
* http://hbase.apache.org/book/thrift.html
* http://hbase.apache.org/apidocs/org/apache/hadoop/hbase/thrift/package-summary.html
* https://github.com/pomack/thrift4go
* https://github.com/samuel/go-thrift



Protocol
===

TBinaryProtocol  	
TCompactProtocol  	
TJSONProtocol 
TSimpleJSONProtocol
TDebugProtocoal  

Transport 
===

TSocket  
TFramedTransport   
TFileTransport   
TMemoryTransport   
TZlibTransport 

License
===

Apache License v2.0  