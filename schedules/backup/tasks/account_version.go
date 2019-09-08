package tasks

import (
	"fmt"
	"time"

	. "github.com/oldfritter/goDCE/models"
	"github.com/oldfritter/goDCE/utils"
)

func BackupAccountVersions() {
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var first, last AccountVersion
	mainDB.Where("created_at < ?", time.Now().Add(-time.Hour*24*30)).First(&first)
	mainDB.Where("created_at < ?", time.Now().Add(-time.Hour*24*30)).Last(&last)
	limit := 1000
	maxThreads := 16

	c := make(chan int)
	quit := make(chan int)
	i := 0
	for i < maxThreads {
		go func(j int) {
			if (first.Id + j*limit) < last.Id {
				backupRowsAccountVersions(first.Id+j*limit, last.Id, limit)
				c <- 1
			}
		}(i)
		if (first.Id + i*limit) > last.Id {
			break
		}
		i += 1
	}

	for _ = range c {
		go func(j int) {
			if (first.Id + j*limit) > last.Id {
				time.Sleep(10 * time.Second)
				quit <- 1
			} else {
				backupRowsAccountVersions(first.Id+j*limit, last.Id, limit)
				c <- 1
			}
		}(i)
		if (first.Id + i*limit) > last.Id {
			break
		}
		i += 1
	}

	<-quit
	fmt.Println("quiting...")
	time.Sleep(10 * time.Second)
	return

}

func backupRowsAccountVersions(begin, lastId, limit int) {
	fmt.Println(begin)
	end := begin + limit - 1
	if end > lastId {
		end = lastId
	}
	mainDB := utils.MainDbBegin()
	defer mainDB.DbRollback()
	var avs []AccountVersion
	mainDB.Where("id between ? AND ?", begin, end).Find(&avs)
	for _, av := range avs {
		insertRowToBackup(&av, 9)
	}

	for _, av := range avs {
		mainDB.Delete(&av)
	}
	mainDB.DbCommit()
	return
}

func insertRowToBackup(av *AccountVersion, times int) {
	backupDB := utils.BackupDbBegin()
	defer backupDB.DbRollback()
	var avBackup AccountVersion
	if backupDB.Where("id = ?", av.Id).First(&avBackup).RecordNotFound() {
		sql := fmt.Sprintf("INSERT INTO account_versions (id, created_at, updated_at, user_id, account_id, reason, balance, locked, fee, amount, modifiable_id, modifiable_type, currency_id, fun) VALUES (%v, '%v', '%v', %v, %v, %v, %v, %v, %v, %v, %v, '%v', %v, %v)",
			(*av).Id,
			(*av).CreatedAt.Format("2006-01-02 15:04:05"),
			(*av).UpdatedAt.Format("2006-01-02 15:04:05"),
			(*av).UserId,
			(*av).AccountId,
			(*av).Reason,
			(*av).Balance,
			(*av).Locked,
			(*av).Fee,
			(*av).Amount,
			(*av).ModifiableId,
			(*av).ModifiableType,
			(*av).CurrencyId,
			(*av).Fun,
		)
		result := backupDB.Exec(sql)
		if result.RowsAffected == 1 {
			backupDB.DbCommit()
			return
		}
		backupDB.DbRollback()
		if times > 0 {
			insertRowToBackup(av, times-1)
		}
	}
	return
}
