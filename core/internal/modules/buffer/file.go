package buffer

import (
	"EXSync/core/internal/config"
	serverOption "EXSync/core/option/exsync/server"
	"os"
	"time"
)

// FileWrite 文件写入缓冲区
func (f *File) FileWrite(file *os.File, StartSize, TotalSize int64, filePath string, date int64, verifyManage serverOption.VerifyManage) error {
	buf := make([]byte, 256)                       // 256 * 4096 = 1048576 = 1MB
	unix := time.Unix(date-verifyManage.Offset, 0) // 计算文件时间偏移后的日期, 并转换为time对象
	for {
		result, err := f.TimeChannel.GetTimeout(f.FileMark, config.SocketTimeout)
		if err != nil {
			return err
		}
		buf = append(buf, result...)
		StartSize += f.DataBlock

		if len(buf) < 256 {
			// 如果接收块未满, 此时又接收数据完毕, 将缓冲区数据写入文件并退出
			if StartSize >= TotalSize {
				_, err = file.Write(buf)
				if err != nil {
					return err
				}
				err = os.Chtimes(filePath, unix, unix)
				if StartSize >= TotalSize {
					return err
				}
				return nil
			}
		} else {
			// 接收块已满, 将缓冲区数据写入文件
			// 如果此时数据接收完毕则退出, 否则清空缓冲区继续接收
			_, err = file.Write(buf)
			if err != nil {
				return err
			}
			err = os.Chtimes(filePath, unix, unix)
			if StartSize >= TotalSize {
				return err
			}
			buf = buf[:0]
		}
	}
}
