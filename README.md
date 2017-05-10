Overview
===

golang client of hbase (via thrift)


Install
===

The official thrift package is out of date, need install [thrift4go](https://github.com/apesternikov/thrift4go) first,
then

	go get github.com/chenjingping/goh

Usage
===

	address := "192.168.17.129:9090"
	
	client, err := goh.NewTcpClient(address, goh.TBinaryProtocol, false)
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
	


Start/Stop thrift 
===

	#start
	./bin/hbase-daemon.sh start thrift

	#set parameter
	./bin/hbase-daemon.sh start thrift –threadpool -m 200 -w 500

	#stop
	./bin/hbase-daemon.sh stop thrift

	#set HEAPSIZE (conf/hbase-env.sh)
	export HBASE_HEAPSIZE=1000MB

Changes
=== 
* 1 apache thrift 0.10.0
* 2 fix error

License
===

Apache License v2.0  