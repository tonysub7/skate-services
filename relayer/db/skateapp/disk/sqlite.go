package disk

import (
	"database/sql"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
	"skatechain.org/lib/db"
	config "skatechain.org/relayer/db"

	"skatechain.org/lib/logging"
)

var (
	TaskLogger = db.NewFileLogger(config.DbDir, "skateapp_tasks.log")
	SkateAppDB *sql.DB
)

const (
	SignedTaskSchema = "SignedTasks"
)

func init() {
	logger := logging.NewLoggerWithConsoleWriter()
	db, err := sql.Open("sqlite3", filepath.Join(config.DbDir, "skateapp.db"))
	if err != nil {
		logger.Fatal("Relayer DB initialization failed")
		panic("Relayer DB initialization failed")
	}
	SkateAppDB = db
}

type SignedTask struct {
	TaskId    uint32
	Message   string
	Initiator string
	ChainType uint32
	ChainId   uint32
	Hash      [32]byte
	Operator  string
	Signature [65]byte
}

func InitializeSkateApp() {
	SkateAppDB.Exec(`CREATE TABLE IF NOT EXISTS ` + SignedTaskSchema + ` (
		id           INTEGER PRIMARY KEY AUTOINCREMENT,
	  taskId       INTEGER,
	  message      TEXT,
	  initiator    TEXT,
	  chainId      INTEGER,
	  chainType    INTEGER,
	  hash         BLOB,
	  operator     TEXT,
	  signature    BLOB
	)`)
}

func InsertSignedTask(signedTask SignedTask) error {
	_, err := SkateAppDB.Exec(
		"INSERT INTO "+SignedTaskSchema+" (taskId, message, initiator, chainId, chainType, hash, operator, signature) VALUES (?,?,?,?,?,?,?,?)",
		signedTask.TaskId, signedTask.Message, signedTask.Initiator, signedTask.ChainId,
		signedTask.ChainType, signedTask.Hash, signedTask.Operator, signedTask.Signature,
	)
	if err != nil {
		TaskLogger.Error("InsertTask failed", "task", signedTask, "err", err)
		return err
	}
	return nil
}

func RetrieveSignedTasks() ([]SignedTask, error) {
	rows, err := SkateAppDB.Query("SELECT * FROM " + SignedTaskSchema)
	if err != nil {
		TaskLogger.Error("SelectAllTasks failed", "err", err)
		return nil, err
	}
	defer rows.Close()

	var resultTasks []SignedTask

	for rows.Next() {
		var task SignedTask
		var entryid int

		err := rows.Scan(
			&entryid, &task.TaskId, &task.Message, &task.Initiator,
			&task.ChainId, &task.ChainType, &task.Hash, &task.Operator, &task.Signature,
		)
		if err != nil {
			return nil, err
		}
		TaskLogger.Info("Task", "task", task)
		resultTasks = append(resultTasks, task)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return resultTasks, nil
}
