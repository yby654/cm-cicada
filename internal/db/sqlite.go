package db

import (
	"strconv"

	"github.com/cloud-barista/cm-cicada/common"
	"github.com/cloud-barista/cm-cicada/internal/domain"
	"github.com/cloud-barista/cm-cicada/internal/lib/config"
	"github.com/jollaman999/utils/logger"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var DB *gorm.DB

func Open() error {
	var err error
	DB, err = gorm.Open(sqlite.Open(common.RootPath+"/"+common.ModuleName+".db?_journal_mode=WAL&_busy_timeout=10000"), &gorm.Config{})
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	err = DB.AutoMigrate(
		&domain.WorkflowTemplate{},
		&domain.Workflow{},
		&domain.WorkflowVersion{},
		&domain.TaskComponent{},
		&domain.TaskGroup{},
		&domain.Task{},
		&domain.WorkflowRun{},
		&domain.TaskRun{},
	)
	if err != nil {
		logger.Panicln(logger.ERROR, true, err)
	}

	logger.Println(logger.INFO, false, "Loading workflow templates...")
	err = WorkflowTemplateInit()
	if err != nil {
		logger.Println(logger.ERROR, true, err)
	}

	taskComponentLoadExamples, _ := strconv.ParseBool(config.CMCicadaConfig.CMCicada.TaskComponent.LoadExamples)
	if taskComponentLoadExamples {
		logger.Println(logger.INFO, false, "Loading task components...")
		err = TaskComponentInit()
		if err != nil {
			logger.Println(logger.ERROR, true, err)
		}
	}

	return err
}

func Close() {
	if DB != nil {
		sqlDB, _ := DB.DB()
		_ = sqlDB.Close()
	}
}

func BeginTransaction() (*gorm.DB, error) {
	tx := DB.Begin()
	if tx.Error != nil {
		return nil, tx.Error
	}
	return tx, nil
}
