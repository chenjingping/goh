/*


 */

package goh

import (
	"bytes"

	"github.com/chenjingping/goh/hbase1"
)

/*
HbaseError
*/
type HbaseError struct {
	IOErr  *hbase1.IOError         // IOError
	ArgErr *hbase1.IllegalArgument // IllegalArgument
	Err    error                   // error

}

func newHbaseError(io *hbase1.IOError, arg *hbase1.IllegalArgument, err error) *HbaseError {
	return &HbaseError{
		IOErr:  io,
		ArgErr: arg,
		Err:    err,
	}
}

/*
String
*/
func (e *HbaseError) String() string {
	if e == nil {
		return "<nil>"
	}

	var b bytes.Buffer
	if e.IOErr != nil {
		b.WriteString("IOError:")
		b.WriteString(e.IOErr.Message)
		b.WriteString(";")
	}

	if e.ArgErr != nil {
		b.WriteString("ArgumentError:")
		b.WriteString(e.ArgErr.Message)
		b.WriteString(";")
	}

	if e.Err != nil {
		b.WriteString("Error:")
		b.WriteString(e.Err.Error())
		b.WriteString(";")
	}
	return b.String()
}

/*
Error
*/
func (e *HbaseError) Error() string {
	return e.String()
}

func checkHbaseError(io *hbase1.IOError, err error) error {
	if io != nil || err != nil {
		return newHbaseError(io, nil, err)
	}
	return nil
}

func checkHbaseArgError(io *hbase1.IOError, arg *hbase1.IllegalArgument, err error) error {
	if io != nil || arg != nil || err != nil {
		return newHbaseError(io, arg, err)
	}
	return nil
}
