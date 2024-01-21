package lan

import (
	"net"
	"time"
)

func ping(host string, onlineHostsChan chan string) {
	var size int
	var timeout int64
	var seq int16 = 1
	const EchoRequestHeadLen = 8

	size = 32
	timeout = 1000

	starttime := time.Now()
	conn, err := net.DialTimeout("ip4:icmp", host, time.Duration(timeout*1000*1000))
	if err != nil {
		return
	}
	defer conn.Close()
	id0, id1 := genidentifier(host)

	var msg = make([]byte, size+EchoRequestHeadLen)
	msg[0] = 8                        // echo
	msg[1] = 0                        // code 0
	msg[2] = 0                        // checksum
	msg[3] = 0                        // checksum
	msg[4], msg[5] = id0, id1         //identifier[0] identifier[1]
	msg[6], msg[7] = gensequence(seq) //sequence[0], sequence[1]

	length := size + EchoRequestHeadLen

	check := checkSum(msg[0:length])
	msg[2] = byte(check >> 8)
	msg[3] = byte(check & 255)

	err = conn.SetDeadline(starttime.Add(time.Duration(timeout * 1000 * 1000)))
	if err != nil {
		return
	}
	_, err = conn.Write(msg[0:length])

	const ECHO_REPLY_HEAD_LEN = 20

	var receive []byte = make([]byte, ECHO_REPLY_HEAD_LEN+length)
	n, err := conn.Read(receive)
	_ = n
	var endduration int = int(int64(time.Since(starttime)) / (1000 * 1000))

	if err != nil || receive[ECHO_REPLY_HEAD_LEN+4] != msg[4] || receive[ECHO_REPLY_HEAD_LEN+5] != msg[5] || receive[ECHO_REPLY_HEAD_LEN+6] != msg[6] || receive[ECHO_REPLY_HEAD_LEN+7] != msg[7] || endduration >= int(timeout) || receive[ECHO_REPLY_HEAD_LEN] == 11 {
		//
	} else {
		onlineHostsChan <- host
		//Log.Print("扫描到主机地址:", host)
	}

}

func checkSum(msg []byte) uint16 {
	sum := 0
	length := len(msg)
	for i := 0; i < length-1; i += 2 {
		sum += int(msg[i])*256 + int(msg[i+1])
	}
	if length%2 == 1 {
		sum += int(msg[length-1]) * 256 // notice here, why *256?
	}

	sum = (sum >> 16) + (sum & 0xffff)
	sum += (sum >> 16)
	var answer uint16 = uint16(^sum)
	return answer
}

func gensequence(v int16) (byte, byte) {
	ret1 := byte(v >> 8)
	ret2 := byte(v & 255)
	return ret1, ret2
}

func genidentifier(host string) (byte, byte) {
	return host[0], host[1]
}
