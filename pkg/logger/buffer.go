package logger

import (
	"strings"
	"sync"
)

type MultiBuffer struct {
	buffers  [2]*strings.Builder
	writeBuf int
	readBuf  int
	writePos int
	readPos  int
	maxSize  int
	mux      sync.Mutex
}

func NewMultiBuffer(maxSize int) *MultiBuffer {
	return &MultiBuffer{
		buffers:  [2]*strings.Builder{{}, {}},
		writeBuf: 0,
		readBuf:  0,
		writePos: 0,
		readPos:  0,
		maxSize:  maxSize,
	}
}

func (mb *MultiBuffer) Write(message string) {
	mb.mux.Lock()
	defer mb.mux.Unlock()

	// 写入当前缓冲区
	mb.buffers[mb.writeBuf].WriteString(message)
	mb.writePos += len(message)

	// 如果当前缓冲区达到最大大小，切换缓冲区
	if mb.writePos >= mb.maxSize {
		mb.writeBuf = 1 - mb.writeBuf
		mb.writePos = 0
		mb.buffers[mb.writeBuf].Reset()
	}
}

func (mb *MultiBuffer) Read() string {
	mb.mux.Lock()
	defer mb.mux.Unlock()

	// 读取当前缓冲区
	data := mb.buffers[mb.readBuf].String()
	mb.readPos += len(data)

	// 如果当前缓冲区已经读取完毕，切换缓冲区
	if mb.readPos >= mb.buffers[mb.readBuf].Len() {
		mb.readBuf = 1 - mb.readBuf
		mb.readPos = 0
		mb.buffers[mb.readBuf].Reset()
	}

	return data
}
