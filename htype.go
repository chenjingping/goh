/*
 */

package goh

import (
	"fmt"
	"strconv"
	"github.com/chenjingping/goh/hbase1"
)

//type Text []byte

func textListToStr(list []hbase1.Text) []string {
	if list == nil {
		return nil
	}

	l := len(list)
	data := make([]string, l)
	for i := 0; i < l; i++ {
		data[i] = string(list[i])
	}

	return data
}

func toHbaseTextList(list []string) []hbase1.Text {
	if list == nil {
		return nil
	}

	l := len(list)
	data := make([]hbase1.Text, l)
	for i := 0; i < l; i++ {
		data[i] = hbase1.Text(list[i])
	}
	return data
}

func toHbaseTextListFromByte(list [][]byte) []hbase1.Text {
	if list == nil {
		return nil
	}

	l := len(list)
	data := make([]hbase1.Text, l)
	for i := 0; i < l; i++ {
		data[i] = hbase1.Text(list[i])
	}
	return data
}

func toHbaseTextMap(source map[string]string) map[string]hbase1.Text {
	if source == nil {
		return nil
	}

	data := make(map[string]hbase1.Text, len(source))
	for k, v := range source {
		data[k] = hbase1.Text(v)
	}

	return data
}

//type Bytes []byte

//type ScannerID int32

/**
 * An HColumnDescriptor contains information about a column family
 * such as the number of versions, compression settings, etc. It is
 * used as input when creating a table or adding a column.
 *
 * Attributes:
 *  - Name
 *  - MaxVersions
 *  - Compression
 *  - InMemory
 *  - BloomFilterType
 *  - BloomFilterVectorSize
 *  - BloomFilterNbHashes
 *  - BlockCacheEnabled
 *  - TimeToLive
 */
type ColumnDescriptor struct {
	Name                  string "name"                  // 1
	MaxVersions           int32  "maxVersions"           // 2
	Compression           string "compression"           // 3
	InMemory              bool   "inMemory"              // 4
	BloomFilterType       string "bloomFilterType"       // 5
	BloomFilterVectorSize int32  "bloomFilterVectorSize" // 6
	BloomFilterNbHashes   int32  "bloomFilterNbHashes"   // 7
	BlockCacheEnabled     bool   "blockCacheEnabled"     // 8
	TimeToLive            int32  "timeToLive"            // 9
}

func toColumn(col *hbase1.ColumnDescriptor) *ColumnDescriptor {
	return &ColumnDescriptor{
		Name:                  string(col.Name),
		MaxVersions:           col.MaxVersions,
		Compression:           col.Compression,
		InMemory:              col.InMemory,
		BloomFilterType:       col.BloomFilterType,
		BloomFilterVectorSize: col.BloomFilterVectorSize,
		BloomFilterNbHashes:   col.BloomFilterNbHashes,
		BlockCacheEnabled:     col.BlockCacheEnabled,
		TimeToLive:            col.TimeToLive,
	}

}

func toColMap(cols map[string]*hbase1.ColumnDescriptor) map[string]*ColumnDescriptor {
	if cols == nil {
		return nil
	}
	data := make(map[string]*ColumnDescriptor, len(cols))
	for k, v := range cols {
		data[k] = toColumn(v)
	}
	return data
}

func toHbaseColumn(col *ColumnDescriptor) *hbase1.ColumnDescriptor {
	return &hbase1.ColumnDescriptor{
		Name:                  hbase1.Text(col.Name),
		MaxVersions:           col.MaxVersions,
		Compression:           col.Compression,
		InMemory:              col.InMemory,
		BloomFilterType:       col.BloomFilterType,
		BloomFilterVectorSize: col.BloomFilterVectorSize,
		BloomFilterNbHashes:   col.BloomFilterNbHashes,
		BlockCacheEnabled:     col.BlockCacheEnabled,
		TimeToLive:            col.TimeToLive,
	}

}

func toHbaseColList(cols []*ColumnDescriptor) []*hbase1.ColumnDescriptor {
	if cols == nil {
		return nil
	}

	l := len(cols)
	data := make([]*hbase1.ColumnDescriptor, l)
	for i := 0; i < l; i++ {
		data[i] = toHbaseColumn(cols[i])
	}
	return data
}

func NewColumnDescriptorDefault(name string) *ColumnDescriptor {
	output := &ColumnDescriptor{
		MaxVersions:           3,
		Compression:           "NONE",
		InMemory:              false,
		BloomFilterType:       "NONE",
		BloomFilterVectorSize: 0,
		BloomFilterNbHashes:   0,
		BlockCacheEnabled:     false,
		TimeToLive:            -1,
		Name:                  name,
	}

	return output
}

/**
 * A TRegionInfo contains information about an HTable region.
 *
 * Attributes:
 *  - StartKey
 *  - EndKey
 *  - Id
 *  - Name
 *  - Version
 *  - ServerName
 *  - Port
 */
type TRegionInfo struct {
	StartKey   string "startKey"   // 1
	EndKey     string "endKey"     // 2
	Id         int64  "id"         // 3
	Name       string "name"       // 4
	Version    int8   "version"    // 5
	ServerName string "serverName" // 6
	Port       int32  "port"       // 7
}

func toRegion(region *hbase1.TRegionInfo) *TRegionInfo {
	return &TRegionInfo{
		StartKey:   string(region.StartKey),
		EndKey:     string(region.EndKey),
		Id:         region.ID,
		Name:       string(region.Name),
		Version:    region.Version,
		ServerName: string(region.ServerName),
		Port:       region.Port,
	}

}

func toRegionList(regions []*hbase1.TRegionInfo) []*TRegionInfo {

	if regions == nil {
		return nil
	}

	l := len(regions)
	data := make([]*TRegionInfo, l)
	for i := 0; i < l; i++ {
		data[i] = toRegion(regions[i])
	}

	return data
}

// /**
//  * A Mutation object is used to either update or delete a column-value.
//  *
//  * Attributes:
//  *  - IsDelete
//  *  - Column
//  *  - Value
//  *  - WriteToWAL
//  */
// type Mutation struct {
// 	IsDelete   bool   "isDelete"   // 1
// 	Column     []byte "column"     // 2
// 	Value      []byte "value"      // 3
// 	WriteToWAL bool   "writeToWAL" // 4
// }

func NewMutation(column string, value []byte) *hbase1.Mutation {
	return &hbase1.Mutation{
		IsDelete:   false,
		WriteToWAL: true,
		Column:     hbase1.Text(column),
		Value:      hbase1.Text(value),
	}

}

// /**
//  * A BatchMutation object is used to apply a number of Mutations to a single row.
//  *
//  * Attributes:
//  *  - Row
//  *  - Mutations
//  */
// type BatchMutation struct {
// 	Row       []byte      "row"       // 1
// 	Mutations []*Mutation "mutations" // 2
// }
func NewBatchMutation(row []byte, mutations []*hbase1.Mutation) *hbase1.BatchMutation {
	return &hbase1.BatchMutation{
		Row:       []byte(row),
		Mutations: mutations,
	}

}

// /**
//  * For increments that are not incrementColumnValue
//  * equivalents.
//  *
//  * Attributes:
//  *  - Table
//  *  - Row
//  *  - Column
//  *  - Ammount
//  */
// type TIncrement struct {
// 	Table   []byte "table"   // 1
// 	Row     []byte "row"     // 2
// 	Column  []byte "column"  // 3
// 	Ammount int64  "ammount" // 4
// }

func NewTIncrement(table string, row []byte, column string, ammount int64) *hbase1.TIncrement {
	return &hbase1.TIncrement{
		Table:   hbase1.Text(table),
		Row:     hbase1.Text(row),
		Column:  hbase1.Text(column),
		Ammount: ammount,
	}
}

/**
 * Holds row name and then a map of columns to cells.
 *
 * Attributes:
 *  - Row
 *  - Columns
 */
// type TRowResult struct {
// 	Row     string            "row"     // 1
// 	Columns map[string]*TCell "columns" // 2
// }

/**
 * A Scan object is used to specify scanner parameters when opening a scanner.
 *
 * Attributes:
 *  - StartRow
 *  - StopRow
 *  - Timestamp
 *  - Columns
 *  - Caching
 *  - FilterString
 */
type TScan struct {
	StartRow     []byte   "startRow"     // 1
	StopRow      []byte   "stopRow"      // 2
	Timestamp    int64    "timestamp"    // 3
	Columns      []string "columns"      // 4
	Caching      int32    "caching"      // 5
	FilterString string   "filterString" // 6
	BatchSize    int32    "batchSize"    // 7
}

func toHbaseTScan(scan *TScan) *hbase1.TScan {
	if scan == nil { return nil }
	
	tscan := hbase1.NewTScan()

	tscan.StartRow = hbase1.Text(scan.StartRow)
	tscan.StopRow = hbase1.Text(scan.StopRow)

	if len(scan.Columns) > 0 {
		tscan.Columns = toHbaseTextList(scan.Columns)
	}
	
	if scan.Timestamp > 0 {
		tscan.Timestamp = &scan.Timestamp
	}

	if scan.Caching > 0 {
		tscan.Caching = &scan.Caching
	}
	
	if scan.BatchSize > 0 {
		tscan.BatchSize =  &scan.BatchSize
	}
		
	if scan.FilterString == "" {
		tscan.FilterString = nil
	} else {
		tscan.FilterString = hbase1.Text(scan.FilterString)
	}

	
	return tscan
}

/*
ScanHbaseColumns get hbase cell value

*/
func ScanHbaseColumns(cols map[string]*hbase1.TCell, nams []string, dest... interface{}) error {
	for idx, nam := range nams {
		if val, ok := cols[nam]; ok {
			tmp := string(val.Value)
			
			d := dest[idx]
			switch dt := d.(type) {
			case *int:
				i, _ := strconv.ParseInt(tmp, 10, 0)
				*dt = int(i)
			case *int64:
				i, _ := strconv.ParseInt(tmp, 10, 0)
				*dt = int64(i)
			case *int32:
				i, _ := strconv.ParseInt(tmp, 10, 0)
				*dt = int32(i)
			case *int16:
				i, _ := strconv.ParseInt(tmp, 10, 0)
				*dt = int16(i)
			case *int8:
				i, _ := strconv.ParseInt(tmp, 10, 0)
				*dt = int8(i)
			case *float64:
				*dt, _ = strconv.ParseFloat(tmp, 64)
			case *float32:
				f, _ := strconv.ParseFloat(tmp, 32)
				*dt = float32(f)
			case *string:
				*dt = tmp
			case *[]byte:
				*dt = []byte(val.Value)
			case *bool:
				b, _ := strconv.ParseBool(tmp)
				*dt = b
			default:
				return fmt.Errorf("Can't scan value of type %T with value %v", dt, tmp)
			}
		}
	}

	return nil
}

// /**
//  * TCell - Used to transport a cell value (byte[]) and the timestamp it was
//  * stored with together as a result for get and getRow methods. This promotes
//  * the timestamp of a cell to a first-class value, making it easy to take
//  * note of temporal data. Cell is used all the way from HStore up to HTable.
//  *
//  * Attributes:
//  *  - Value
//  *  - Timestamp
//  */
// type TCell struct {
// 	Value     []byte "value"     // 1
// 	Timestamp int64  "timestamp" // 2
// }

// /**
//  * An IOError exception signals that an error occurred communicating
//  * to the Hbase master or an Hbase region server.  Also used to return
//  * more general Hbase error conditions.
//  *
//  * Attributes:
//  *  - Message
//  */
// type IOError struct {
// 	Message string "message" // 1
// }

// /**
//  * An IllegalArgument exception indicates an illegal or invalid
//  * argument was passed into a procedure.
//  *
//  * Attributes:
//  *  - Message
//  */
// type IllegalArgument struct {
// 	Message string "message" // 1
// }

// /**
//  * An AlreadyExists exceptions signals that a table with the specified
//  * name already exists
//  *
//  * Attributes:
//  *  - Message
//  */
// type AlreadyExists struct {
// 	Message string "message" // 1
// }

// func (t *AlreadyExists) String() string {
// 	if t == nil {
// 		return "<nil>"
// 	}
// 	return t.Message
// }

// func name() {
// 	fmt.Println("...")
// }
