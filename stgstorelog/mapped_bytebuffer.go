// Copyright (c) 2015-2018 All rights reserved.
// 本软件源代码版权归 my.oschina.net/tantexian 所有,允许复制与学习借鉴.
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/6
package stgstorelog

import "bytes"
import (
	"errors"

	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/byteutil"
	"git.oschina.net/cloudzone/smartgo/stgstorelog/mmap"
)

type MappedByteBuffer struct {
	MMapBuf  mmap.MMap
	ReadPos  int // read at &buf[ReadPos]
	WritePos int // write at &buf[WritePos]
	Limit    int // MMapBuf's max Size
}

func NewMappedByteBuffer(mMap mmap.MMap) *MappedByteBuffer {
	mappedByteBuffer := &MappedByteBuffer{}
	mappedByteBuffer.MMapBuf = mMap
	mappedByteBuffer.ReadPos = 0
	mappedByteBuffer.WritePos = 0
	mappedByteBuffer.Limit = cap(mMap)
	return mappedByteBuffer
}

func (m *MappedByteBuffer) Bytes() []byte { return m.MMapBuf[:m.WritePos] }

// Write appends the contents of data to the buffer
func (m *MappedByteBuffer) Write(data []byte) (n int, err error) {
	if m.WritePos+len(data) > m.Limit {
		logger.Errorf("m.WritePos + len(data)(%v > %v) data: %s", m.WritePos+len(data), m.Limit, string(data))
		//panic(errors.New("m.WritePos + len(data)"))
		return 0, errors.New("m.WritePos + len(data)")
	}
	for index, value := range data {
		m.MMapBuf[m.WritePos+index] = value
	}
	m.WritePos += len(data)
	return len(data), nil
}

func (self *MappedByteBuffer) WriteInt8(i int8) (mappedByteBuffer *MappedByteBuffer) {
	toBytes := byteutil.Int8ToBytes(i)
	self.Write(toBytes)
	return self
}

func (self *MappedByteBuffer) WriteInt16(i int16) (mappedByteBuffer *MappedByteBuffer) {
	toBytes := byteutil.Int16ToBytes(i)
	self.Write(toBytes)
	return self
}

func (self *MappedByteBuffer) WriteInt32(i int32) (mappedByteBuffer *MappedByteBuffer) {
	toBytes := byteutil.Int32ToBytes(i)
	self.Write(toBytes)
	return self
}

func (self *MappedByteBuffer) WriteInt64(i int64) (mappedByteBuffer *MappedByteBuffer) {
	toBytes := byteutil.Int64ToBytes(i)
	self.Write(toBytes)
	return self
}

// Read reads the next len(p) bytes from the buffer or until the buffer
// is drained. The return value n is the number of bytes read. If the
// buffer has no data to return, err is io.EOF (unless en(p) is zero);
// otherwise it is nil.
func (m *MappedByteBuffer) Read(data []byte) (n int, err error) {
	n = copy(data, m.MMapBuf[m.ReadPos:])
	m.ReadPos += n
	return
}

func (self *MappedByteBuffer) ReadInt8() (i int8) {
	int8bytes := make([]byte, 1)
	self.Read(int8bytes)
	i = byteutil.BytesToInt8(int8bytes)
	return i
}

func (self *MappedByteBuffer) ReadInt16() (i int16) {
	int16bytes := make([]byte, 2)
	self.Read(int16bytes)
	i = byteutil.BytesToInt16(int16bytes)
	return i
}

func (self *MappedByteBuffer) ReadInt32() (i int32) {
	int32bytes := make([]byte, 4)
	self.Read(int32bytes)
	i = byteutil.BytesToInt32(int32bytes)
	return i
}

func (self *MappedByteBuffer) ReadInt64() (i int64) {
	int64bytes := make([]byte, 8)
	self.Read(int64bytes)
	i = byteutil.BytesToInt64(int64bytes)
	return i
}

// slice 返回当前MappedByteBuffer.byteBuffer中从开始位置到len的分片buffer
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/6
func (self *MappedByteBuffer) slice() *bytes.Buffer {
	return bytes.NewBuffer(self.MMapBuf[:self.ReadPos])
}

func (self *MappedByteBuffer) flush() {
	self.MMapBuf.Flush()
}

func (self *MappedByteBuffer) unmap() {
	self.MMapBuf.Unmap()
}

func (self *MappedByteBuffer) flip() {
	self.Limit = self.WritePos
	self.WritePos = 0
}

func (self *MappedByteBuffer) limit(newLimit int) {
	self.Limit = newLimit
	if self.WritePos > self.Limit {
		self.WritePos = newLimit
	}
}
