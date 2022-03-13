package process

import (
	db "BooPT/database"
	"BooPT/model"
	"BooPT/storage"
	"fmt"
	"os"
	"os/exec"

	"github.com/sirupsen/logrus"
)

func Start() {
	for {
		data := <-channel
		err := ExtractPdfImage(data.downloadLinkId, data.md5)
		if err != nil {
			logrus.Error(err)
			continue
		}
		if err = db.DB.Model(&model.DownloadLink{}).Update("state", model.DownloadLinkState_Processed).Error; err != nil {
			logrus.Error(err)
		}
	}
}

// 提取pdf的前两页作为图片
func ExtractPdfImage(downloadLinkId uint, md5 string) (err error) {

	getObjectName := "unprocessed/" + md5 + "/file.pdf"
	// putObjectName := "processed/" + md5 + "/file.pdf"
	putPngName := "processed/" + md5 + "/file_%d.png"
	fileName := "/tmp/BooPT/file.pdf"
	pngName := "/tmp/BooPT/file_%d.png"

	if err = storage.DownloadFile(getObjectName, fileName); err != nil {
		return
	}

	cmd := exec.Command("gs",
		"-q",
		"-o", pngName,
		"-sDEVICE=png256",
		"-dLastPage=8",
		"-r144",
		"-dFIXEDMEDIA",
		"-dPDFFitPage",
		"-sPAPERSIZE=a4",
		fileName,
	)

	output, err := cmd.Output()

	logrus.Infof("[Process] %s", string(output))
	if err != nil {
		return err
	}

	for i := 0; i < 8; i++ {
		_, err = os.Stat(fmt.Sprintf(pngName, i))
		if err != nil {
			if os.IsNotExist(err) {
				break
			} else {
				return
			}
		}

		if err = storage.UploadFile(fmt.Sprintf(putPngName, i), fmt.Sprintf(pngName, i)); err != nil {
			return
		}
	}

	return
}
