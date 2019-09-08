package tasks

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/oldfritter/goDCE/utils"
)

type MyPutRet struct {
	Key    string
	Hash   string
	Fsize  int
	Bucket string
	Name   string
}

var (
	logNames = []string{"api", "workers", "schedule"}
)

func UploadLogFileToQiniu() {
	UploadLogFileToQiniuByDay(time.Now().Add(-time.Hour * 24))
	UploadLogFileToQiniuByDay(time.Now())
}

func UploadLogFileToS3() {
	UploadLogFileToS3ByDay(time.Now().Add(-time.Hour * 24))
	UploadLogFileToS3ByDay(time.Now())
}

func UploadLogFileToQiniuByDay(day time.Time) {
	utils.InitQiniuConfig()
	for _, logName := range logNames {
		gzFile := "/tmp/panama/" + logName + day.Format("2006-01-02") + ".tar.gz"

		name, _ := exec.Command("sh", "-c", "hostname").Output()
		hostname := string(name)

		exec.Command("sh", "-c", "mkdir -p /tmp/panama/").Output()
		exec.Command("sh", "-c", "tar -czvf "+gzFile+" "+"logs/"+logName+day.Format("2006-01-02")+".log").Output()

		key := "logs/panama/" + day.Format("01/02") + "/" + logName + "/" + hostname + ".tar.gz"

		err := utils.UploadFileToQiniu(utils.QiniuConfig["backup_bucket"], key, gzFile)
		if err != nil {
			fmt.Println("err: ", err)
		}
		exec.Command("sh", "-c", "rm -rf "+gzFile).Output()

	}
}

func UploadLogFileToS3ByDay(day time.Time) {
	utils.InitAwsS3Config()
	for _, logName := range logNames {
		gzFile := "/tmp/panama/" + logName + day.Format("2006-01-02") + ".tar.gz"

		name, _ := exec.Command("sh", "-c", "hostname").Output()
		hostname := string(name)

		exec.Command("sh", "-c", "mkdir -p /tmp/panama/").Output()
		exec.Command("sh", "-c", "tar -czvf "+gzFile+" "+"logs/"+logName+day.Format("2006-01-02")+".log").Output()

		key := "logs/panama/" + day.Format("01/02") + "/" + logName + "/" + hostname + ".tar.gz"

		err := utils.UploadFileToS3(utils.S3Config["S3_BACKUP_BUCKET"], key, gzFile)
		if err != nil {
			fmt.Println("err: ", err)
		}
		exec.Command("sh", "-c", "rm -rf "+gzFile).Output()

	}
}
func BackupLogFiles() {
	for _, logName := range logNames {
		a, _ := filepath.Abs("logs/" + logName + ".log")
		b, _ := filepath.Abs("logs/" + logName + time.Now().Format("2006-01-02") + ".log")
		exec.Command("sh", "-c", "cat "+fmt.Sprintf(a)+" >> "+fmt.Sprintf(b)).Output()
		exec.Command("sh", "-c", "echo '\n' > "+fmt.Sprintf(a)).Output()
	}
}

func CleanLogs() {
	for _, logName := range logNames {
		day := 2
		for day < 10 {
			str := time.Now().Add(-time.Hour * 24 * time.Duration(day)).Format("2006-01-02")
			b, _ := filepath.Abs("logs/" + logName + str + ".log")
			exec.Command("sh", "-c", "rm -rf "+fmt.Sprintf(b)).Output()
			day += 1
		}
	}
}
