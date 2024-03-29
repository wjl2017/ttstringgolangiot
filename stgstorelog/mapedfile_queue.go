// Copyright (c) 2015-2018 All rights reserved.
// 本软件源代码版权归 my.oschina.net/tantexian 所有,允许复制与学习借鉴.
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/7
package stgstorelog

import (
	"container/list"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"git.oschina.net/cloudzone/smartgo/stgcommon/logger"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/fileutil"
	"git.oschina.net/cloudzone/smartgo/stgcommon/utils/timeutil"
)

type MapedFileQueue struct {
	// 每次触发删除文件，最多删除多少个文件
	DeleteFilesBatchMax int
	// 文件存储位置
	storePath string
	// 每个文件的大小
	mapedFileSize int64
	// 各个文件
	mapedFiles *list.List
	// 读写锁（针对mapedFiles）
	rwLock *sync.RWMutex
	// 预分配MapedFile对象服务
	allocateMapedFileService *AllocateMapedFileService
	// 刷盘刷到哪里
	committedWhere int64
	// 最后一条消息存储时间
	storeTimestamp int64
}

func NewMapedFileQueue(storePath string, mapedFileSize int64,
	allocateMapedFileService *AllocateMapedFileService) *MapedFileQueue {
	self := &MapedFileQueue{}
	self.storePath = storePath         // 存储路径
	self.mapedFileSize = mapedFileSize // 文件size
	self.mapedFiles = list.New()
	// 根据读写请求队列requestQueue/readQueue中的读写请求，创建对应的mappedFile文件
	self.allocateMapedFileService = allocateMapedFileService
	self.DeleteFilesBatchMax = 10
	self.committedWhere = 0
	self.storeTimestamp = 0
	self.rwLock = new(sync.RWMutex)
	return self
}

func (self *MapedFileQueue) getMapedFileByTime(timestamp int64) (mf *MapedFile) {
	mapedFileSlice := self.copyMapedFiles(0)
	if mapedFileSlice == nil {
		return nil
	}
	for _, mf := range mapedFileSlice {
		if mf != nil {
			fileInfo, err := os.Stat(mf.fileName)
			if err != nil {
				logger.Warn("maped file queue get maped file by time error:", err.Error())
				continue
			}

			modifiedTime := fileInfo.ModTime().UnixNano() / 1000000
			if modifiedTime >= timestamp {
				return mf
			}
		}
	}
	return mapedFileSlice[len(mapedFileSlice)-1]
}

// copyMapedFiles 获取当前mapedFiles列表中的副本slice
// Params: reservedMapedFiles 只有当mapedFiles列表中文件个数大于reservedMapedFiles才返回副本
// Return: 返回大于等于reservedMapedFiles个元数的MapedFile切片
// Author: tantexian, <tantexian@qq.com>
// Since: 17/8/9
func (self *MapedFileQueue) copyMapedFiles(reservedMapedFiles int) []*MapedFile {
	mapedFileSlice := make([]*MapedFile, self.mapedFiles.Len())
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	if self.mapedFiles.Len() <= reservedMapedFiles {
		return nil
	}
	// Iterate through list and print its contents.
	for e := self.mapedFiles.Front(); e != nil; e = e.Next() {
		mf := e.Value.(*MapedFile)
		if mf != nil {
			mapedFileSlice = append(mapedFileSlice, mf)
		}
	}
	return mapedFileSlice
}

// truncateDirtyFiles recover时调用，不需要加锁
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/7
func (self *MapedFileQueue) truncateDirtyFiles(offset int64) {
	willRemoveFiles := list.New()
	// Iterate through list and print its contents.
	for e := self.mapedFiles.Front(); e != nil; e = e.Next() {
		mf := e.Value.(*MapedFile)
		fileTailOffset := mf.fileFromOffset + mf.fileSize
		if fileTailOffset > offset {
			if offset >= mf.fileFromOffset {
				pos := offset % int64(self.mapedFileSize)
				mf.wrotePostion = pos
				mf.mappedByteBuffer.WritePos = int(pos)
				mf.committedPosition = pos
			} else {
				mf.destroy(1000)
				willRemoveFiles.PushBack(mf)
			}
		}
	}
	self.deleteExpiredFile(willRemoveFiles)
}

// deleteExpiredFile 删除过期文件只能从头开始删
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/7
func (self *MapedFileQueue) deleteExpiredFile(mfs *list.List) {
	if mfs != nil && mfs.Len() > 0 {
		self.rwLock.Lock()
		defer self.rwLock.Unlock()
		for de := mfs.Front(); de != nil; de = de.Next() {
			for e := self.mapedFiles.Front(); e != nil; e = e.Next() {
				deleteFile := de.Value.(*MapedFile)
				file := e.Value.(*MapedFile)

				if deleteFile.fileName == file.fileName {
					success := self.mapedFiles.Remove(e)
					if success == false {
						logger.Error("deleteExpiredFile remove failed.")
					}
				}
			}
		}
	}
}

// load 从磁盘加载mapedfile到内存映射
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/8
func (self *MapedFileQueue) load() bool {
	exist, err := PathExists(self.storePath)
	if err != nil {
		logger.Infof("maped file queue load store path error:", err.Error())
		return false
	}

	if exist {
		files, err := fileutil.ListFilesOrDir(self.storePath, "FILE")
		if err != nil {
			logger.Error(err.Error())
			return false
		}
		if len(files) > 0 {
			// 按照文件名，升序排列
			sort.Strings(files)
			for _, path := range files {
				file, error := os.Stat(path)
				if error != nil {
					logger.Errorf("maped file queue load file %s error: %s", path, error.Error())
				}

				if file == nil {
					logger.Errorf("maped file queue load file not exist: ", path)
				}

				size := file.Size()
				// 校验文件大小是否匹配
				if size != int64(self.mapedFileSize) {
					logger.Warn("filesize(%d) mapedFileSize(%d) length not matched message store config value, ignore it", size, self.mapedFileSize)
					return true
				}

				// 恢复队列
				mapedFile, error := NewMapedFile(path, int64(self.mapedFileSize))
				if error != nil {
					logger.Error("maped file queue load file error:", error.Error())
					return false
				}

				mapedFile.wrotePostion = self.mapedFileSize
				mapedFile.committedPosition = self.mapedFileSize
				mapedFile.mappedByteBuffer.WritePos = int(mapedFile.wrotePostion)
				self.mapedFiles.PushBack(mapedFile)
				logger.Infof("load mapfiled %v success.", mapedFile.fileName)
			}
		}
	}

	return true
}

// howMuchFallBehind 刷盘进度落后了多少
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/8
func (self *MapedFileQueue) howMuchFallBehind() int64 {
	if self.mapedFiles.Len() == 0 {
		return 0
	}
	committed := self.committedWhere
	if committed != 0 {
		mapedFile, error := self.getLastMapedFile(0)
		if error != nil {
			logger.Error(error.Error())
		}
		return mapedFile.fileFromOffset + mapedFile.wrotePostion - committed
	}
	return 0
}

// getLastMapedFile 获取最后一个MapedFile对象，如果一个都没有，则新创建一个，
// 如果最后一个写满了，则新创建一个
// Params: startOffset 如果创建新的文件，起始offset
// Author: tantexian, <tantexian@qq.com>
// Since: 2017/8/8
func (self *MapedFileQueue) getLastMapedFile(startOffset int64) (*MapedFile, error) {
	var createOffset int64 = -1
	var mapedFile, mapedFileLast *MapedFile
	self.rwLock.RLock()
	if self.mapedFiles.Len() == 0 {
		createOffset = startOffset - (startOffset % int64(self.mapedFileSize))
	} else {
		mapedFileLastObj := (self.mapedFiles.Back().Value).(*MapedFile)
		mapedFileLast = mapedFileLastObj
	}
	self.rwLock.RUnlock()
	if mapedFileLast != nil && mapedFileLast.isFull() {
		createOffset = mapedFileLast.fileFromOffset + self.mapedFileSize
	}

	if createOffset != -1 {
		nextPath := self.storePath + string(filepath.Separator) + fileutil.Offset2FileName(createOffset)
		nextNextPath := self.storePath + string(filepath.Separator) +
			fileutil.Offset2FileName(createOffset+self.mapedFileSize)
		if self.allocateMapedFileService != nil {
			var err error
			mapedFile, err = self.allocateMapedFileService.putRequestAndReturnMapedFile(nextPath, nextNextPath, self.mapedFileSize)
			if err != nil {
				logger.Errorf("put request and return maped file, error:%s ", err.Error())
				return nil, err
			}
		} else {
			var err error
			mapedFile, err = NewMapedFile(nextPath, self.mapedFileSize)
			if err != nil {
				logger.Errorf("maped file create maped file error: %s", err.Error())
				return nil, err
			}
		}

		if mapedFile != nil {
			self.rwLock.Lock()
			if self.mapedFiles.Len() == 0 {
				mapedFile.firstCreateInQueue = true
			}
			self.mapedFiles.PushBack(mapedFile)
			self.rwLock.Unlock()
		}

		return mapedFile, nil
	}

	return mapedFileLast, nil
}

func (self *MapedFileQueue) getMinOffset() int64 {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	if self.mapedFiles.Len() > 0 {
		mappedfile := self.mapedFiles.Front().Value.(MapedFile)
		return mappedfile.fileFromOffset
	}

	return -1
}

func (self *MapedFileQueue) getMaxOffset() int64 {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	if self.mapedFiles.Len() > 0 {
		mappedfile := self.mapedFiles.Back().Value.(*MapedFile)
		return mappedfile.fileFromOffset + mappedfile.wrotePostion
	}

	return 0
}

// deleteLastMapedFile 恢复时调用
func (self *MapedFileQueue) deleteLastMapedFile() {
	if self.mapedFiles.Len() != 0 {
		last := self.mapedFiles.Back()
		lastMapedFile := last.Value.(*MapedFile)
		lastMapedFile.destroy(1000)
		self.mapedFiles.Remove(last)
		logger.Infof("on recover, destroy a logic maped file %v", lastMapedFile.fileName)
	}
}

// deleteExpiredFileByTime 根据文件过期时间来删除物理队列文件
// Return: 删除过期文件的数量
// Author: tantexian, <tantexian@qq.com>
// Since: 17/8/9
func (self *MapedFileQueue) deleteExpiredFileByTime(expiredTime int64, deleteFilesInterval int,
	intervalForcibly int64, cleanImmediately bool) int {
	// 获取当前MapedFiles列表中所有元素副本的切片
	files := self.copyMapedFiles(0)
	if len(files) == 0 {
		return 0
	}
	toBeDeleteMfList := list.New()
	var delCount int = 0
	// 最后一个文件处于写状态，不能删除
	mfsLength := len(files) - 1
	for i := 0; i < mfsLength; i++ {
		mf := files[i]
		if mf != nil {
			liveMaxTimestamp := mf.storeTimestamp + expiredTime
			if timeutil.CurrentTimeMillis() > liveMaxTimestamp || cleanImmediately {
				if mf.destroy(intervalForcibly) {
					toBeDeleteMfList.PushBack(mf)
					delCount++
					// 每次触发删除文件，最多删除多少个文件
					if toBeDeleteMfList.Len() >= self.DeleteFilesBatchMax {
						break
					}
					// 删除最后一个文件不需要等待
					if deleteFilesInterval > 0 && (i+1) < mfsLength {
						time.Sleep(time.Second * time.Duration(deleteFilesInterval))
					}
				} else {
					break
				}
			}
		}
	}

	self.deleteExpiredFile(toBeDeleteMfList)

	return delCount
}

// deleteExpiredFileByOffset 根据物理队列最小Offset来删除逻辑队列
// Params: offset 物理队列最小offset
// Params: unitsize ???
// Author: tantexian, <tantexian@qq.com>
// Since: 17/8/9
func (self *MapedFileQueue) deleteExpiredFileByOffset(offset int64, unitsize int) int {
	toBeDeleteFileList := list.New()
	deleteCount := 0
	mfs := self.copyMapedFiles(0)

	if mfs != nil && len(mfs) > 0 {
		// 最后一个文件处于写状态，不能删除
		mfsLength := len(mfs) - 1

		for i := 0; i < mfsLength; i++ {
			destroy := true
			mf := mfs[i]

			if mf == nil {
				continue
			}

			result := mf.selectMapedBuffer(self.mapedFileSize - int64(unitsize))

			if result != nil {
				maxOffsetInLogicQueue := result.MappedByteBuffer.ReadInt64()
				result.Release()
				// 当前文件是否可以删除
				destroy = maxOffsetInLogicQueue < offset
				if destroy {
					logger.Infof("physic min offset %d, logics in current mapedfile max offset %d, delete it",
						offset, maxOffsetInLogicQueue)
				}
			} else {
				logger.Warn("this being not excuted forever.")
				break
			}

			if destroy && mf.destroy(1000*60) {
				toBeDeleteFileList.PushBack(mf)
				deleteCount++
			}
		}
	}

	self.deleteExpiredFile(toBeDeleteFileList)
	return deleteCount
}

func (self *MapedFileQueue) commit(flushLeastPages int32) bool {
	result := true

	mapedFile := self.findMapedFileByOffset(self.committedWhere, true)
	if mapedFile != nil {
		tmpTimeStamp := mapedFile.storeTimestamp
		offset := mapedFile.Commit(flushLeastPages)
		where := mapedFile.fileFromOffset + offset
		result = (where == self.committedWhere)
		self.committedWhere = where

		if 0 == flushLeastPages {
			self.storeTimestamp = tmpTimeStamp
		}
	}

	return result
}

func (self *MapedFileQueue) getFirstMapedFile() *MapedFile {
	if self.mapedFiles.Len() == 0 {
		return nil
	}

	element := self.mapedFiles.Front()
	result := element.Value.(*MapedFile)

	return result
}

func (self *MapedFileQueue) getLastMapedFile2() *MapedFile {
	if self.mapedFiles.Len() == 0 {
		return nil
	}

	lastElement := self.mapedFiles.Back()
	result, ok := lastElement.Value.(*MapedFile)
	if !ok {
		logger.Info("mapedfile queue get last maped file type conversion error")
	}

	return result
}

func (self *MapedFileQueue) findMapedFileByOffset(offset int64, returnFirstOnNotFound bool) *MapedFile {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()

	mapedFile := self.getFirstMapedFile()
	if mapedFile != nil {
		index := (offset / self.mapedFileSize) - (mapedFile.fileFromOffset / self.mapedFileSize)
		if index < 0 || index >= int64(self.mapedFiles.Len()) {
			logger.Warnf("maped file queue find maped file by offset, offset not matched, request Offset: %d, index: %d, mapedFileSize: %d, mapedFiles count: %d",
				offset, index, self.mapedFileSize, self.mapedFiles.Len())
		}

		i := 0
		for e := self.mapedFiles.Front(); e != nil; e = e.Next() {
			if i == int(index) {
				result := e.Value.(*MapedFile)
				return result
			}
			i++
		}

		if returnFirstOnNotFound {
			return mapedFile
		}

	}

	return nil
}

func (self *MapedFileQueue) getLastAndLastMapedFile() *MapedFile {
	if self.mapedFiles.Len() == 0 {
		return nil
	}

	element := self.mapedFiles.Back()
	result := element.Value.(*MapedFile)

	return result
}

func (self *MapedFileQueue) getMapedMemorySize() int64 {
	return 0
}

func (self *MapedFileQueue) retryDeleteFirstFile(intervalForcibly int64) bool {
	mapFile := self.getFirstMapedFileOnLock()
	if mapFile != nil {
		if !mapFile.isAvailable() {
			logger.Warn("the mapedfile was destroyed once, but still alive, ", mapFile.fileName)

			result := mapFile.destroy(intervalForcibly)
			if result {
				logger.Info("the mapedfile redelete OK, ", mapFile.fileName)
				tmps := list.New()
				tmps.PushBack(mapFile)
				self.deleteExpiredFile(tmps)
			} else {
				logger.Warn("the mapedfile redelete Failed, ", mapFile.fileName)
			}

			return result
		}
	}

	return false
}

func (self *MapedFileQueue) getFirstMapedFileOnLock() *MapedFile {
	self.rwLock.RLock()
	defer self.rwLock.RUnlock()
	return self.getFirstMapedFile()
}

// shutdown 关闭队列，队列数据还在，但是不能访问
func (self *MapedFileQueue) shutdown(intervalForcibly int64) {

}

// destroy 销毁队列，队列数据被删除，此函数有可能不成功
func (self *MapedFileQueue) destroy() {
	self.rwLock.Lock()
	self.rwLock.Unlock()

	for element := self.mapedFiles.Front(); element != nil; element = element.Next() {
		mapedFile, ok := element.Value.(*MapedFile)
		if !ok {
			logger.Warnf("maped file queue destroy type conversion error")
			continue
		}
		mapedFile.destroy(1000 * 3)
	}

	self.mapedFiles.Init()
	self.committedWhere = 0

	// delete parent director
	exist, err := PathExists(self.storePath)
	if err != nil {
		logger.Warn("maped file queue destroy check store path is exists, error:", err.Error())
	}

	if exist {
		if storeFile, _ := os.Stat(self.storePath); storeFile.IsDir() {
			os.RemoveAll(self.storePath)
		}
	}
}
