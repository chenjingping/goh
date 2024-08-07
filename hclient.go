package goh

import (
	"errors"
	"net"
	"net/url"

	"github.com/chenjingping/thrift/lib/go/thrift"
	"github.com/chenjingping/goh/hbase1"
)

/*
HClient is wrap of hbase client
*/
type HClient struct {
	addr            string
	Protocol        int
	Trans           thrift.TTransport
	ProtocolFactory thrift.TProtocolFactory
	hbase           *hbase1.HbaseClient
	state           int
}

/*
Dail return hbase client struct

*/
func Dail(name, host, port string) (interface{}, error) {
	cli, err := NewTCPClient(host, port, TBinaryProtocol, false)

	if err == nil {
		if err := cli.Open(); err != nil {
			return nil, err
		}

		return cli, nil
	}
	
	return nil, err
}

/*
CloseCli return hbase close status

*/
func CloseCli(itf interface{}) error {
	if cli, ok := itf.(*HClient); ok {
		return cli.Close()
	}

	return errors.New("client close failed")
}

/*
KeepAlive return hbase online status

*/
func KeepAlive(itf interface{}) error {
	if cnn, ok := itf.(*HClient); ok {
		if tbls, _ := cnn.GetTableNames(); tbls != nil {
			return nil
		}

		return errors.New("the hbase is not available")
	}

	return errors.New("client keepalive failed")
}

/*
NewHTTPClient return a hbase http client instance

*/
func NewHTTPClient(rawurl string, protocol int) (client *HClient, err error) {
	parsedURL, err := url.Parse(rawurl)
	if err != nil {
		return nil, err
	}

	trans, err := thrift.NewTHttpClient(parsedURL.String())
	if err != nil {
		return nil, err
	}

	return newClient(parsedURL.String(), protocol, trans)
}

/*
NewTCPClient return a base tcp client instance

*/
func NewTCPClient(ip string, port string, protocol int, framed bool) (client *HClient, err error) {
	var trans thrift.TTransport
	
	addr := net.JoinHostPort(ip, port)

	trans, err = thrift.NewTSocket(addr)
	if err != nil {
		return nil, err
	}
	if framed {
		trans = thrift.NewTFramedTransport(trans)
	}else {
		trans = thrift.NewTBufferedTransport(trans, 8192)
	}

	return newClient(addr, protocol, trans)
}

/*
newClient create a new hbase client
*/
func newClient(addr string, protocol int, trans thrift.TTransport) (*HClient, error) {
	var client *HClient

	protocolFactory, err := newProtocolFactory(protocol)
	if err != nil {
		return client, err
	}

	client = &HClient{
		addr:            addr,
		Protocol:        protocol,
		ProtocolFactory: protocolFactory,
		Trans:           trans,
		hbase:           hbase1.NewHbaseClientFactory(trans, protocolFactory),
	}
	
	return client, nil
}

/*
Open connection
*/
func (client *HClient) Open() error {
	if client.state == stateDefault {
		if err := client.Trans.Open(); err != nil {
			return err
		}
		client.state = stateOpen
	}
	return nil
}

/*
Close connection
*/
func (client *HClient) Close() error {
	if client.state == stateOpen {
		if err := client.Trans.Close(); err != nil {
			return err
		}
		client.state = stateDefault
	}
	return nil
}

/**
 * Brings a table on-line (enables it)
 *
 * Parameters:
 *  - TableName: name of the table
 */
func (client *HClient) EnableTable(tableName string) error {
	return client.hbase.EnableTable(hbase1.Bytes(tableName))
}

/**
 * Disables a table (takes it off-line) If it is being served, the master
 * will tell the servers to stop serving it.
 *
 * Parameters:
 *  - TableName: name of the table
 */
func (client *HClient) DisableTable(tableName string) (err error) {
	return client.hbase.DisableTable(hbase1.Bytes(tableName))
}

/**
 * @return true if table is on-line
 *
 * Parameters:
 *  - TableName: name of the table to check
 */
func (client *HClient) IsTableEnabled(tableName string) (ret bool, err error) {
	return client.hbase.IsTableEnabled(hbase1.Bytes(tableName))
}

/**
 * Parameters:
 *  - TableNameOrRegionName
 */
func (client *HClient) Compact(tableNameOrRegionName string) (err error) {
	return client.hbase.Compact(hbase1.Bytes(tableNameOrRegionName))
}

/**
 * Parameters:
 *  - TableNameOrRegionName
 */
func (client *HClient) MajorCompact(tableNameOrRegionName string) (err error) {
	return client.hbase.MajorCompact(hbase1.Bytes(tableNameOrRegionName))
}

/**
 * List all the column families assoicated with a table.
 *
 * @return list of column family descriptors
 *
 * Parameters:
 *  - TableName: table name
 */
func (client *HClient) GetTableNames() (tables []string, err error) {
	ret, e1 := client.hbase.GetTableNames()
	if err = checkHbaseError(nil, e1); err != nil {
		return
	}

	tables = textListToStr(ret)
	return
}

/**
 * List all the column families assoicated with a table.
 *
 * @return list of column family descriptors
 *
 * Parameters:
 *  - TableName: table name
 */
func (client *HClient) GetColumnDescriptors(tableName string) (columns map[string]*ColumnDescriptor, err error) {
	ret, e1 := client.hbase.GetColumnDescriptors(hbase1.Text(tableName))
	if err = checkHbaseError(nil, e1); err != nil {
		return
	}
	columns = toColMap(ret)
	return
}

/**
 * List the regions associated with a table.
 *
 * @return list of region descriptors
 *
 * Parameters:
 *  - TableName: table name
 */
func (client *HClient) GetTableRegions(tableName string) (regions []*TRegionInfo, err error) {
	if ret, e1 := client.hbase.GetTableRegions(hbase1.Text(tableName)); e1 == nil {
		return toRegionList(ret), nil
	} else {
		return nil, e1
	}
}

/**
 * Create a table with the specified column families.  The name
 * field for each ColumnDescriptor must be set and must end in a
 * colon (:). All other fields are optional and will get default
 * values if not explicitly specified.
 *
 * @throws IllegalArgument if an input parameter is invalid
 *
 * @throws AlreadyExists if the table name already exists
 *
 * Parameters:
 *  - TableName: name of table to create
 *  - ColumnFamilies: list of column family descriptors
 */
func (client *HClient) CreateTable(tableName string, columnFamilies []*ColumnDescriptor) (exists bool, err error) {
	columns := toHbaseColList(columnFamilies)

	if err = client.hbase.CreateTable(hbase1.Text(tableName), columns); err != nil {
		return
	}
	exists = (err != nil)
	return
}

/**
 * Deletes a table
 *
 * @throws IOError if table doesn't exist on server or there was some other
 * problem
 *
 * Parameters:
 *  - TableName: name of table to delete
 */
func (client *HClient) DeleteTable(tableName string) (err error) {
	return client.hbase.DeleteTable(hbase1.Text(tableName))
}

/**
 * Get a single TCell for the specified table, row, and column at the
 * latest timestamp. Returns an empty list if no such value exists.
 *
 * @return value for specified row/column
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Column: column name
 *  - Attributes: Get attributes
 */
func (client *HClient) Get(tableName string, row []byte, column string, attributes map[string]string) (data []*hbase1.TCell, err error) {
	return client.hbase.Get(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), toHbaseTextMap(attributes))
}

/**
 * Get the specified number of versions for the specified table,
 * row, and column.
 *
 * @return list of cells for specified row/column
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Column: column name
 *  - NumVersions: number of versions to retrieve
 *  - Attributes: Get attributes
 */
func (client *HClient) GetVer(tableName string, row []byte, column string, numVersions int32, attributes map[string]string) (data []*hbase1.TCell, err error) {
	return client.hbase.GetVer(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), numVersions, toHbaseTextMap(attributes))
}

/**
 * Get the specified number of versions for the specified table,
 * row, and column.  Only versions less than or equal to the specified
 * timestamp will be returned.
 *
 * @return list of cells for specified row/column
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Column: column name
 *  - Timestamp: timestamp
 *  - NumVersions: number of versions to retrieve
 *  - Attributes: Get attributes
 */
func (client *HClient) GetVerTs(tableName string, row []byte, column string, timestamp int64, numVersions int32, attributes map[string]string) (data []*hbase1.TCell, err error) {
	return client.hbase.GetVerTs(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), timestamp, numVersions, toHbaseTextMap(attributes))
}

/**
 * Get all the data for the specified table and row at the latest
 * timestamp. Returns an empty list if the row does not exist.
 *
 * @return TRowResult containing the row and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRow(tableName string, row []byte, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRow(hbase1.Text(tableName), hbase1.Text(row), toHbaseTextMap(attributes))
}

/**
 * Get the specified columns for the specified table and row at the latest
 * timestamp. Returns an empty list if the row does not exist.
 *
 * @return TRowResult containing the row and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Columns: List of columns to return, null for all columns
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowWithColumns(tableName string, row []byte, columns []string, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRowWithColumns(hbase1.Text(tableName), hbase1.Text(row), toHbaseTextList(columns), toHbaseTextMap(attributes))
}

/**
 * Get all the data for the specified table and row at the specified
 * timestamp. Returns an empty list if the row does not exist.
 *
 * @return TRowResult containing the row and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of the table
 *  - Row: row key
 *  - Timestamp: timestamp
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowTs(tableName string, row []byte, timestamp int64, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRowTs(hbase1.Text(tableName), hbase1.Text(row), timestamp, toHbaseTextMap(attributes))
}

/**
 * Get the specified columns for the specified table and row at the specified
 * timestamp. Returns an empty list if the row does not exist.
 *
 * @return TRowResult containing the row and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Columns: List of columns to return, null for all columns
 *  - Timestamp
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowWithColumnsTs(tableName string, row []byte, columns []string, timestamp int64, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRowWithColumnsTs(hbase1.Text(tableName), hbase1.Text(row), toHbaseTextList(columns), timestamp, toHbaseTextMap(attributes))
}

/**
 * Get all the data for the specified table and rows at the latest
 * timestamp. Returns an empty list if no rows exist.
 *
 * @return TRowResult containing the rows and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Rows: row keys
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRows(tableName string, rows [][]byte, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRows(hbase1.Text(tableName), toHbaseTextListFromByte(rows), toHbaseTextMap(attributes))
}

/**
 * Get the specified columns for the specified table and rows at the latest
 * timestamp. Returns an empty list if no rows exist.
 *
 * @return TRowResult containing the rows and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Rows: row keys
 *  - Columns: List of columns to return, null for all columns
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowsWithColumns(tableName string, rows [][]byte, columns []string, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	if err = client.Open(); err != nil {
		return nil, err
	}

	return client.hbase.GetRowsWithColumns(hbase1.Text(tableName), toHbaseTextListFromByte(rows), toHbaseTextList(columns), toHbaseTextMap(attributes))
}

/**
 * Get all the data for the specified table and rows at the specified
 * timestamp. Returns an empty list if no rows exist.
 *
 * @return TRowResult containing the rows and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of the table
 *  - Rows: row keys
 *  - Timestamp: timestamp
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowsTs(tableName string, rows [][]byte, timestamp int64, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRowsTs(hbase1.Text(tableName), toHbaseTextListFromByte(rows), timestamp, toHbaseTextMap(attributes))
}

/**
 * Get the specified columns for the specified table and rows at the specified
 * timestamp. Returns an empty list if no rows exist.
 *
 * @return TRowResult containing the rows and map of columns to TCells
 *
 * Parameters:
 *  - TableName: name of table
 *  - Rows: row keys
 *  - Columns: List of columns to return, null for all columns
 *  - Timestamp
 *  - Attributes: Get attributes
 */
func (client *HClient) GetRowsWithColumnsTs(tableName string, rows [][]byte, columns []string, timestamp int64, attributes map[string]string) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.GetRowsWithColumnsTs(hbase1.Text(tableName), toHbaseTextListFromByte(rows), toHbaseTextList(columns), timestamp, toHbaseTextMap(attributes))
}

/*
 * Apply a series of mutations (updates/deletes) to a row in a
 * single transaction.  If an exception is thrown, then the
 * transaction is aborted.  Default current timestamp is used, and
 * all entries will have an identical timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Mutations: list of mutation commands
 *  - Attributes: Mutation attributes
 */
func (client *HClient) MutateRow(tableName string, row []byte, mutations []*hbase1.Mutation, attributes map[string]string) error {
	return client.hbase.MutateRow(hbase1.Text(tableName), hbase1.Text(row), mutations, toHbaseTextMap(attributes))
}

/**
 * Apply a series of mutations (updates/deletes) to a row in a
 * single transaction.  If an exception is thrown, then the
 * transaction is aborted.  The specified timestamp is used, and
 * all entries will have an identical timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Mutations: list of mutation commands
 *  - Timestamp: timestamp
 *  - Attributes: Mutation attributes
 */
func (client *HClient) MutateRowTs(tableName string, row []byte, mutations []*hbase1.Mutation, timestamp int64, attributes map[string]string) error {
	return client.hbase.MutateRowTs(hbase1.Text(tableName), hbase1.Text(row), mutations, timestamp, toHbaseTextMap(attributes))
}

/**
 * Apply a series of batches (each a series of mutations on a single row)
 * in a single transaction.  If an exception is thrown, then the
 * transaction is aborted.  Default current timestamp is used, and
 * all entries will have an identical timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - RowBatches: list of row batches
 *  - Attributes: Mutation attributes
 */
func (client *HClient) MutateRows(tableName string, rowBatches []*hbase1.BatchMutation, attributes map[string]string) error {
	return client.hbase.MutateRows(hbase1.Text(tableName), rowBatches, toHbaseTextMap(attributes))
}

/**
 * Apply a series of batches (each a series of mutations on a single row)
 * in a single transaction.  If an exception is thrown, then the
 * transaction is aborted.  The specified timestamp is used, and
 * all entries will have an identical timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - RowBatches: list of row batches
 *  - Timestamp: timestamp
 *  - Attributes: Mutation attributes
 */
func (client *HClient) MutateRowsTs(tableName string, rowBatches []*hbase1.BatchMutation, timestamp int64, attributes map[string]string) error {
	return client.hbase.MutateRowsTs(hbase1.Text(tableName), rowBatches, timestamp, toHbaseTextMap(attributes))
}

/**
 * Atomically increment the column value specified.  Returns the next value post increment.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row to increment
 *  - Column: name of column
 *  - Value: amount to increment by
 */
func (client *HClient) AtomicIncrement(tableName string, row []byte, column string, value int64) (v int64, err error) {
	return client.hbase.AtomicIncrement(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), value)
}

/**
 * Delete all cells that match the passed row and column.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: Row to update
 *  - Column: name of column whose value is to be deleted
 *  - Attributes: Delete attributes
 */
func (client *HClient) DeleteAll(tableName string, row []byte, column string, attributes map[string]string) error {
	return client.hbase.DeleteAll(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), toHbaseTextMap(attributes))
}

/**
 * Delete all cells that match the passed row and column and whose
 * timestamp is equal-to or older than the passed timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: Row to update
 *  - Column: name of column whose value is to be deleted
 *  - Timestamp: timestamp
 *  - Attributes: Delete attributes
 */
func (client *HClient) DeleteAllTs(tableName string, row []byte, column string, timestamp int64, attributes map[string]string) error {
	return client.hbase.DeleteAllTs(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(column), timestamp, toHbaseTextMap(attributes))
}

/**
 * Completely delete the row's cells.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: key of the row to be completely deleted.
 *  - Attributes: Delete attributes
 */
func (client *HClient) DeleteAllRow(tableName string, row []byte, attributes map[string]string) error {
	return client.hbase.DeleteAllRow(hbase1.Text(tableName), hbase1.Text(row), toHbaseTextMap(attributes))
}

/**
 * Increment a cell by the ammount.
 * Increments can be applied async if hbase.regionserver.thrift.coalesceIncrement is set to true.
 * False is the default.  Turn to true if you need the extra performance and can accept some
 * data loss if a thrift server dies with increments still in the queue.
 *
 * Parameters:
 *  - Increment: The single increment to apply
 */
func (client *HClient) Increment(increment *hbase1.TIncrement) error {
	return client.hbase.Increment(increment)
}

/**
 * Parameters:
 *  - Increments: The list of increments
 */
func (client *HClient) IncrementRows(increments []*hbase1.TIncrement) error {
	return client.hbase.IncrementRows(increments)
}

/**
 * Completely delete the row's cells marked with a timestamp
 * equal-to or older than the passed timestamp.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: key of the row to be completely deleted.
 *  - Timestamp: timestamp
 *  - Attributes: Delete attributes
 */
func (client *HClient) DeleteAllRowTs(tableName string, row []byte, timestamp int64, attributes map[string]string) error {
	return client.hbase.DeleteAllRowTs(hbase1.Text(tableName), hbase1.Text(row), timestamp, toHbaseTextMap(attributes))
}

/**
 * Get a scanner on the current table, using the Scan instance
 * for the scan parameters.
 *
 * Parameters:
 *  - TableName: name of table
 *  - Scan: Scan instance
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpenWithScan(tableName string, scan *TScan, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpenWithScan(hbase1.Text(tableName), toHbaseTScan(scan), toHbaseTextMap(attributes))

	return int32(ret), e1
}

/**
 * Get a scanner on the current table starting at the specified row and
 * ending at the last row in the table.  Return the specified columns.
 *
 * @return scanner id to be used with other scanner procedures
 *
 * Parameters:
 *  - TableName: name of table
 *  - StartRow: Starting row in table to scan.
 * Send "" (empty string) to start at the first row.
 *  - Columns: columns to scan. If column name is a column family, all
 * columns of the specified column family are returned. It's also possible
 * to pass a regex in the column qualifier.
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpen(tableName string, startRow []byte, columns []string, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpen(hbase1.Text(tableName), hbase1.Text(startRow), toHbaseTextList(columns), toHbaseTextMap(attributes))
	
	return int32(ret), e1
}

/**
 * Get a scanner on the current table starting and stopping at the
 * specified rows.  ending at the last row in the table.  Return the
 * specified columns.
 *
 * @return scanner id to be used with other scanner procedures
 *
 * Parameters:
 *  - TableName: name of table
 *  - StartRow: Starting row in table to scan.
 * Send "" (empty string) to start at the first row.
 *  - StopRow: row to stop scanning on. This row is *not* included in the
 * scanner's results
 *  - Columns: columns to scan. If column name is a column family, all
 * columns of the specified column family are returned. It's also possible
 * to pass a regex in the column qualifier.
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpenWithStop(tableName string, startRow []byte, stopRow []byte, columns []string, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpenWithStop(hbase1.Text(tableName), hbase1.Text(startRow), hbase1.Text(stopRow), toHbaseTextList(columns), toHbaseTextMap(attributes))
	
	return int32(ret), e1
}

/**
 * Open a scanner for a given prefix.  That is all rows will have the specified
 * prefix. No other rows will be returned.
 *
 * @return scanner id to use with other scanner calls
 *
 * Parameters:
 *  - TableName: name of table
 *  - StartAndPrefix: the prefix (and thus start row) of the keys you want
 *  - Columns: the columns you want returned
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpenWithPrefix(tableName string, startAndPrefix []byte, columns []string, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpenWithPrefix(hbase1.Text(tableName), hbase1.Text(startAndPrefix), toHbaseTextList(columns), toHbaseTextMap(attributes))

	return int32(ret), e1
}

/**
 * Get a scanner on the current table starting at the specified row and
 * ending at the last row in the table.  Return the specified columns.
 * Only values with the specified timestamp are returned.
 *
 * @return scanner id to be used with other scanner procedures
 *
 * Parameters:
 *  - TableName: name of table
 *  - StartRow: Starting row in table to scan.
 * Send "" (empty string) to start at the first row.
 *  - Columns: columns to scan. If column name is a column family, all
 * columns of the specified column family are returned. It's also possible
 * to pass a regex in the column qualifier.
 *  - Timestamp: timestamp
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpenTs(tableName string, startRow []byte, columns []string, timestamp int64, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpenTs(hbase1.Text(tableName), hbase1.Text(startRow), toHbaseTextList(columns), timestamp, toHbaseTextMap(attributes))

	return int32(ret), e1
}

/**
 * Get a scanner on the current table starting and stopping at the
 * specified rows.  ending at the last row in the table.  Return the
 * specified columns.  Only values with the specified timestamp are
 * returned.
 *
 * @return scanner id to be used with other scanner procedures
 *
 * Parameters:
 *  - TableName: name of table
 *  - StartRow: Starting row in table to scan.
 * Send "" (empty string) to start at the first row.
 *  - StopRow: row to stop scanning on. This row is *not* included in the
 * scanner's results
 *  - Columns: columns to scan. If column name is a column family, all
 * columns of the specified column family are returned. It's also possible
 * to pass a regex in the column qualifier.
 *  - Timestamp: timestamp
 *  - Attributes: Scan attributes
 */
func (client *HClient) ScannerOpenWithStopTs(tableName string, startRow []byte, stopRow []byte, columns []string, timestamp int64, attributes map[string]string) (id int32, err error) {
	ret, e1 := client.hbase.ScannerOpenWithStopTs(hbase1.Text(tableName), hbase1.Text(startRow), hbase1.Text(stopRow), toHbaseTextList(columns), timestamp, toHbaseTextMap(attributes))
	
	return int32(ret), e1
}

/**
 * Returns the scanner's current row value and advances to the next
 * row in the table.  When there are no more rows in the table, or a key
 * greater-than-or-equal-to the scanner's specified stopRow is reached,
 * an empty list is returned.
 *
 * @return a TRowResult containing the current row and a map of the columns to TCells.
 *
 * @throws IllegalArgument if ScannerID is invalid
 *
 * @throws NotFound when the scanner reaches the end
 *
 * Parameters:
 *  - Id: id of a scanner returned by scannerOpen
 */
func (client *HClient) ScannerGet(id int32) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.ScannerGet(hbase1.ScannerID(id))
}

/**
 * Returns, starting at the scanner's current row value nbRows worth of
 * rows and advances to the next row in the table.  When there are no more
 * rows in the table, or a key greater-than-or-equal-to the scanner's
 * specified stopRow is reached,  an empty list is returned.
 *
 * @return a TRowResult containing the current row and a map of the columns to TCells.
 *
 * @throws IllegalArgument if ScannerID is invalid
 *
 * @throws NotFound when the scanner reaches the end
 *
 * Parameters:
 *  - Id: id of a scanner returned by scannerOpen
 *  - NbRows: number of results to return
 */
func (client *HClient) ScannerGetList(id int32, nbRows int32) (data []*hbase1.TRowResult_, err error) {
	return client.hbase.ScannerGetList(hbase1.ScannerID(id), nbRows)
}

/**
 * Closes the server-state associated with an open scanner.
 *
 * @throws IllegalArgument if ScannerID is invalid
 *
 * Parameters:
 *  - Id: id of a scanner returned by scannerOpen
 */
func (client *HClient) ScannerClose(id int32) error {
	return client.hbase.ScannerClose(hbase1.ScannerID(id))
}

/**
 * Get the row just before the specified one.
 *
 * @return value for specified row/column
 *
 * Parameters:
 *  - TableName: name of table
 *  - Row: row key
 *  - Family: column name
 */
func (client *HClient) GetRowOrBefore(tableName string, row string, family string) (data []*hbase1.TCell, err error) {
	ret, e1 := client.hbase.GetRowOrBefore(hbase1.Text(tableName), hbase1.Text(row), hbase1.Text(family))
	if err = checkHbaseError(nil, e1); err != nil {
		return
	}

	data = ret
	return
}

/**
 * Get the regininfo for the specified row. It scans
 * the metatable to find region's start and end keys.
 *
 * @return value for specified row/column
 *
 * Parameters:
 *  - Row: row key
 */
func (client *HClient) GetRegionInfo(row string) (region *TRegionInfo, err error) {
	ret, e1 := client.hbase.GetRegionInfo(hbase1.Text(row))
	if err = checkHbaseError(nil, e1); err != nil {
		return
	}

	region = toRegion(ret)
	return
}
