package process

type processData struct {
	downloadLinkId uint
	md5            string
}

var channel = make(chan processData, 10)

// 通过消息队列发送消息给子线程，使子线程处理信息
func SendMessage(downloadLinkId uint, md5 string) {
	data := processData{
		downloadLinkId: downloadLinkId,
		md5:            md5,
	}

	channel <- data
}
