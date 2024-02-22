package buffer

import "EXSync/core/internal/modules/timechannel"

type File struct {
	TimeChannel *timechannel.TimeChannel
	FileMark    string
	DataBlock   int64
}
